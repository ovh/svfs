package svfs

var (
	ListerConcurrency uint64
)

// Lister is a concurrent processor of direntries.
// Its job is to get extra information about files.
type Lister struct {
	taskChan chan ListerTask
}

// ListerTask represents a manifest ready to be processed by
// the Lister. Every task must provide a manifest object and
// a result channel to which retrieved information will be sent.
type ListerTask struct {
	n  Node
	rc chan<- Node
}

// Start spawns workers waiting for tasks. Once a task comes
// in the task channel, one worker will process it by opening
// a connection to swift and asking information about the
// current object.
func (dl *Lister) Start() {
	dl.taskChan = make(chan ListerTask, ListerConcurrency)
	for i := 0; uint64(i) < ListerConcurrency; i++ {
		go func() {
			for t := range dl.taskChan {
				// Standard swift object
				if o, ok := t.n.(*Object); ok {
					ro, h, _ := SwiftConnection.Object(o.c.Name, o.so.Name)
					if SegmentPathRegex.Match([]byte(h[ManifestHeader])) {
						o.segmented = true
					}
					o.sh = &h
					o.so = &ro
					t.rc <- o
				}
				// Directory
				if d, ok := t.n.(*Directory); ok {
					rd, h, _ := SwiftConnection.Object(d.c.Name, d.so.Name)
					d.sh = &h
					d.so = &rd
					t.rc <- d
				}
			}
		}()
	}
}

// AddTask asynchronously adds a new task to be processed. It
// returns immediately with no guarantee that the task has been
// added to the channel nor retrieved by a worker.
func (dl *Lister) AddTask(n Node, rc chan Node) {
	go func() {
		dl.taskChan <- ListerTask{
			n:  n,
			rc: rc,
		}
	}()
}
