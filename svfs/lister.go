package svfs

// DirLister is a concurrent processor for segmented objects.
// Its job is to get information about manifests stored within
// directories.
type DirLister struct {
	concurrency uint64
	taskChan    chan DirListerTask
}

// DirListerTask represents a manifest ready to be processed by
// the DirLister. Every task must provide a manifest object and
// a result channel to which retrieved information will be send.
type DirListerTask struct {
	n  Node
	rc chan<- Node
}

// Start spawns workers waiting for tasks. Once a task comes
// in the task channel, one worker will process it by opening
// a connection to swift and asking information about the
// current manifest. The real size of the object is modified
// then it sends the modified object into the task result
// channel.
func (dl *DirLister) Start() {
	dl.taskChan = make(chan DirListerTask, dl.concurrency)
	for i := 0; uint64(i) < dl.concurrency; i++ {
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
func (dl *DirLister) AddTask(n Node, rc chan Node) {
	go func() {
		dl.taskChan <- DirListerTask{
			n:  n,
			rc: rc,
		}
	}()
}
