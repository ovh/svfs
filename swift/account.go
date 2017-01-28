package swift

import (
	lib "github.com/xlucas/swift"
)

type Account struct {
	*lib.Account
	lib.Headers
}
