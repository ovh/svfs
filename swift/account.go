package swift

import (
	"math"
	"strconv"
	"time"

	lib "github.com/xlucas/swift"
)

type Account struct {
	*lib.Account
	lib.Headers
}

func (a *Account) timestamp() (secs int64, nsecs int64, err error) {
	timestamp, err := strconv.ParseFloat(a.Headers[TimestampHeader], 64)
	if err != nil {
		return
	}

	secs = int64(timestamp)
	nsecs = int64((timestamp - float64(secs)) * math.Pow10(9))

	return
}

func (a *Account) CreationTime() (t time.Time) {
	secs, nsecs, _ := a.timestamp()
	return time.Unix(secs, nsecs)
}
