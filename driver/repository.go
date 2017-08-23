package driver

import "fmt"

var repository = make(map[string]*Group)

func GetGroup(name string) *Group {
	if group, ok := repository[name]; ok {
		return group
	}

	panic(fmt.Sprintf("Unknwown driver group '%s'", name))
}

func RegisterGroup(name string, group *Group) {
	repository[name] = group
}
