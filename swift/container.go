package swift

import (
	"math"
	"strconv"
	"strings"
	"time"

	. "github.com/ahmetalpbalkan/go-linq"
	lib "github.com/xlucas/swift"
)

// Container represents a wrapper over the native golang Swift library.
type Container struct {
	*lib.Container
	lib.Headers
}

func (c *Container) timestamp() (secs int64, nsecs int64, err error) {
	timestamp, err := strconv.ParseFloat(c.Headers[TimestampHeader], 64)
	if err != nil {
		return
	}

	secs = int64(timestamp)
	nsecs = int64((timestamp - float64(secs)) * math.Pow10(9))

	return
}

func (c *Container) CreationTime() (t time.Time) {
	secs, nsecs, _ := c.timestamp()
	return time.Unix(secs, nsecs)
}

func (c *Container) SelectHeaders(prefix string) lib.Headers {
	headers := make(lib.Headers)

	From(c.Headers).WhereT(func(kv KeyValue) bool {
		return strings.HasPrefix(kv.Key.(string), prefix)
	}).ToMap(&headers)

	return headers
}

// ContainerList is a collection of containers that provides filtering.
type ContainerList map[string]*Container

// Filter retrieves a new list by dropping items not matching the predicate.
func (l ContainerList) Filter(predicate interface{}) ContainerList {
	results := make(ContainerList)

	From(l).WhereT(predicate).ToMap(&results)

	return results
}

// FilterByStoragePolicy drops containers not using a storage policy.
func (l ContainerList) FilterByStoragePolicy(policy string) ContainerList {
	return l.Filter(func(e KeyValue) bool {
		return e.Value.(*Container).Headers[StoragePolicyHeader] == policy
	})
}

func (l ContainerList) Slice() (slice []*lib.Container) {
	From(l).SelectT(
		func(kv KeyValue) *lib.Container {
			return kv.Value.(*Container).Container
		},
	).ToSlice(&slice)

	return
}
