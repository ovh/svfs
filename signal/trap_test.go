package signal

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TrapTestSuite struct {
	suite.Suite
	pid  int
	proc os.Process
}

func (suite *TrapTestSuite) SetupSuite() {
	suite.pid = os.Getpid()
	suite.proc = os.Process{Pid: suite.pid}
}

func TestRunTrapSuite(t *testing.T) {
	suite.Run(t, new(TrapTestSuite))
}

func (suite *TrapTestSuite) TestTrap() {
	signals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
	}

	cleanupFunc := func(p interface{}) {
		c := p.(chan bool)
		c <- true
	}

	c := make(chan bool)

	for _, signal := range signals {
		Trap(cleanupFunc, c)

		suite.proc.Signal(signal)

		select {
		case <-c:
			break
		// Note that a few signals like SIGKILL, SIGSTOP, SIGHUP and so on will
		// trigger a go runtime exit, preventing this code branch from running.
		case <-time.After(100 * time.Millisecond):
			suite.Fail(fmt.Sprintf("Signal '%v' not trapped", signal))
		}
	}
}
