package signal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Trap sets up appropriate signal "trap" for common behavior expected from
// vanilla unix command-line tool in general.
//
// When SIGINT or SIGTERM is received, `cleanup` is executed and `cleanupArgs`
// is passed to the call and the signal trap exits.
func Trap(cleanup func(interface{}), cleanupArgs interface{}) {
	signals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
	}
	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, signals...)

	go func() {
		signal := <-signalChan
		fmt.Fprintf(os.Stdout, "Processing signal '%v'\n", signal)
		cleanup(cleanupArgs)
		return
	}()
}
