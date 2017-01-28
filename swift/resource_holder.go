package swift

import "sync"

type ResourceHolder struct {
	capacity uint32
	borrows  uint32
	mutex    sync.RWMutex
	resource interface{}
}

func NewResourceHolder(capacity uint32, resource interface{}) *ResourceHolder {
	return &ResourceHolder{
		capacity: capacity,
		resource: resource,
	}
}

func (h *ResourceHolder) Borrow() interface{} {
	var granted bool

Acquire:
	h.onWriteLock(func() {
		if h.borrows < h.capacity {
			h.borrows++
			granted = true
		}
	})
	if granted == false {
		goto Acquire
	}

	return h.resource
}

func (h *ResourceHolder) Return() {
	h.onWriteLock(func() { h.borrows-- })
}

func (h *ResourceHolder) onLock(lock func(), unlock func(), handler func()) {
	lock()
	defer unlock()
	handler()
}

func (h *ResourceHolder) onReadLock(handler func()) {
	h.onLock(h.mutex.RLock, h.mutex.RUnlock, handler)
}

func (h *ResourceHolder) onWriteLock(handler func()) {
	h.onLock(h.mutex.Lock, h.mutex.Unlock, handler)
}
