package svfs

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

const (
	objectSizeHeader  = objectMetaHeader + "Crypto-Origin-Size"
	objectNonceHeader = objectMetaHeader + "Crypto-Nonce"
)

func newCipher(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func newNonce(cipher cipher.AEAD) ([]byte, error) {
	nonce := make([]byte, cipher.NonceSize())
	_, err := io.ReadFull(rand.Reader, nonce)
	return nonce, err
}

func updateHeaders(object *Object, nonce string) (err error) {
	// Crypto headers
	headers := map[string]string{
		objectSizeHeader:  fmt.Sprintf("%d", object.so.Bytes),
		objectNonceHeader: nonce,
	}

	// Update current node headers
	h := object.sh.ObjectMetadata().Headers(objectMetaHeader)
	for k, v := range headers {
		object.sh[k] = v
		h[k] = v
	}

	// Update object
	if SwiftConnection.ObjectUpdate(object.c.Name, object.path, h) != nil {
		err = fmt.Errorf("Failed to update object crypto headers")
	}

	return err
}

// CryptoHandler is the parent struct of all svfs cryptographic handlers.
type CryptoHandler struct {
	cipher    cipher.AEAD
	Nonce     []byte
	blockSize int64
	overhead  int64
	cellSize  int64
	offset    int64
	cellID    int64
}

// SetCipher register the specified AEAD cipher and nonce to use for
// data decryption or encryption.
func (h *CryptoHandler) SetCipher(cipher cipher.AEAD, nonce []byte) {
	h.cipher = cipher
	h.Nonce = nonce
}

// CryptoReadSeeker reads encrypted swift objects. It supports
// seeking for a specific offset. It reads data produced with a
// static cipher block and padding size (i.e. overhead).
type CryptoReadSeeker struct {
	CryptoHandler
	reader io.ReadSeeker
	fetch  bool   // Force cell fetch
	block  []byte // Current block
	cell   []byte // Current cell
}

// NewCryptoReadSeeker returns a new CryptoReadSeeker.
func NewCryptoReadSeeker(reader io.ReadSeeker, blockSize, overhead int64) *CryptoReadSeeker {
	return &CryptoReadSeeker{
		CryptoHandler: CryptoHandler{
			blockSize: blockSize,
			overhead:  overhead,
			cellSize:  blockSize + overhead,
			cellID:    -1,
			offset:    -1,
		},
		cell:   make([]byte, blockSize+overhead),
		reader: reader,
	}
}

func (r *CryptoReadSeeker) decrypt(buf []byte) ([]byte, error) {
	n, _ := io.ReadFull(r.reader, buf)
	if n == 0 {
		return buf[:n], nil
	}
	return r.cipher.Open(nil, r.Nonce, buf[:n], nil)
}

// Seek jumps to the specified offset in the stream. It expects an absolute offset
// value (whence should always be 0).
func (r *CryptoReadSeeker) Seek(offset int64, whence int) (newPos int64, err error) {
	if whence != 0 {
		return 0, fmt.Errorf("Bad whence given, expecting 0")
	}

	newPos = offset
	cell := newPos / r.blockSize

	if newPos == r.offset || cell == r.cellID {
		r.offset = newPos
		return
	}

	r.fetch = true
	r.cellID = cell
	r.offset = newPos
	pOffset := r.cellID * r.cellSize

	return r.reader.Seek(pOffset, 0)
}

// Read fills a slice with consecutive bytes decrypted from the stream,
// starting at the current offset.
func (r *CryptoReadSeeker) Read(p []byte) (n int, err error) {
	var (
		readBuf   []byte
		startPos  = int64(r.offset - r.cellID*r.blockSize)
		rSize     = int64(len(p))
		lastCell  = (r.offset + rSize - 1) / r.blockSize
		firstCell = r.cellID
	)

	// Seeked to a non-prefetched position
	if r.fetch {
		r.block, err = r.decrypt(r.cell)
		if err != nil {
			return 0, err
		}
		r.fetch = false
	}

	// Append decrypted data and prefetch next cell
	for i := firstCell; i <= lastCell; i++ {
		endPos := startPos + r.blockSize

		if i == firstCell {
			endPos = startPos + (i+1)*r.blockSize - r.offset
		}
		if i == lastCell {
			endPos = startPos + (r.offset + rSize - i*r.blockSize)
		}
		if i == firstCell && i == lastCell {
			endPos = startPos + rSize
		}

		// Make sure read request boundary is always contained within
		// the object length.
		if endPos > int64(len(r.block)) {
			endPos = int64(len(r.block))
		}

		// In case there's nothing left to read, stop here.
		if startPos > endPos {
			return len(p), nil
		}

		// Append decrypted data
		readBuf = append(readBuf, r.block[startPos:endPos]...)

		// Prefetch the next block
		if endPos == r.blockSize {
			r.block, err = r.decrypt(r.cell)
			if err != nil {
				return 0, err
			}
			r.cellID = i + 1
			startPos = 0
		} else {
			startPos = endPos
		}

	}

	// Copy decrypted data to the read buffer
	copy(p, readBuf)

	// Move the offset forward
	r.offset += int64(len(readBuf))

	return len(p), err
}

// Close teardowns the cryptographic reader.
func (r *CryptoReadSeeker) Close() error {
	if closer, ok := r.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// CryptoWriter is a cryptographic handler that manages
// writing encrypted data.
type CryptoWriter struct {
	CryptoHandler
	writer io.WriteCloser
	block  []byte // Current block
}

// NewCryptoWriter allocates a new cryptographic writer that will handle
// writing encrypted data in an existing stream.
func NewCryptoWriter(writer io.WriteCloser, blockSize, overhead int64) *CryptoWriter {
	return &CryptoWriter{
		CryptoHandler: CryptoHandler{
			blockSize: blockSize,
			overhead:  overhead,
			cellSize:  blockSize + overhead,
			cellID:    0,
			offset:    0,
		},
		writer: writer,
	}
}

func (w *CryptoWriter) encrypt(buf []byte) error {
	data := w.cipher.Seal(nil, w.Nonce, buf, nil)
	_, err := w.writer.Write(data)
	return err
}

// Write encrypts the specified data and stores it as regularly
// padded blocks, called cells.
func (w *CryptoWriter) Write(p []byte) (n int, err error) {
	var (
		startPos = int64(0)
		wSize    = int64(len(p))
		lastCell = (w.offset + wSize - 1) / w.blockSize
	)

	// Write encrypted data contiguously
	for i := w.cellID; i <= lastCell; i++ {
		endPos := startPos + w.blockSize

		if i == w.cellID {
			endPos = (i+1)*w.blockSize - w.offset
		}
		if i == lastCell {
			endPos = startPos + (w.offset + wSize - i*w.blockSize)
		}
		if i == w.cellID && i == lastCell {
			endPos = startPos + wSize
		}

		w.block = append(w.block, p[startPos:endPos]...)

		// Write a block only if we entirely filled it
		if i == w.cellID && (w.offset-i*w.blockSize+endPos == w.blockSize) ||
			i != w.cellID && (endPos == w.blockSize) {
			err = w.encrypt(w.block)
			if err != nil {
				return 0, err
			}
			w.block = w.block[:0]
		}

		startPos = endPos
	}

	// Shift offset
	w.offset += wSize

	// Save current cell index
	w.cellID = w.offset / w.blockSize

	return len(p), err
}

// Close tears down the writer, appending any pending
// data not forming a block with a sufficient size to
// be handled by previous write requests.
func (w *CryptoWriter) Close() error {
	if len(w.block) != 0 {
		err := w.encrypt(w.block)
		if err != nil {
			return err
		}
		w.block = w.block[:0]
	}
	return w.writer.Close()
}

var (
	_ io.Reader      = (*CryptoReadSeeker)(nil)
	_ io.Seeker      = (*CryptoReadSeeker)(nil)
	_ io.ReadCloser  = (*CryptoReadSeeker)(nil)
	_ io.Writer      = (*CryptoWriter)(nil)
	_ io.WriteCloser = (*CryptoWriter)(nil)
)
