package driver

import (
	"fmt"
	"reflect"

	linq "github.com/ahmetalpbalkan/go-linq"
)

type Group struct {
	ifaceType reflect.Type
	types     []reflect.Type
}

func NewGroup(iface interface{}) *Group {
	return &Group{
		ifaceType: reflect.TypeOf(iface).Elem(),
	}
}

func (g *Group) Get(driverName string) interface{} {
	driverType := g.getType(driverName)
	return reflect.New(driverType.Elem()).Interface()
}

func (g *Group) Register(driver interface{}) {
	driverType := reflect.TypeOf(driver)
	if driverType.Implements(g.ifaceType) {
		g.types = append(g.types, driverType)
	}
}

func (g *Group) getType(driverName string) (driver reflect.Type) {
	match := linq.From(g.types).FirstWithT(func(t reflect.Type) bool {
		return t.Elem().Name() == driverName
	})

	if match == nil {
		panic(fmt.Sprintf("Driver '%s' not available", driverName))
	}

	return match.(reflect.Type)
}
