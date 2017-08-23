package drivers

import (
	"github.com/ovh/svfs/driver"
	"github.com/ovh/svfs/store"
)

func init() {
	driver.RegisterGroup("store", driver.NewGroup((*store.Store)(nil)))
}
