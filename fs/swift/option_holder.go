package swift

import "github.com/ovh/svfs/fs"

type OptionHolder struct {
	options fs.MountOptions
}

func NewOptionHolder(opts fs.MountOptions) *OptionHolder {
	return &OptionHolder{options: opts}
}

func (o *OptionHolder) Get(opt fs.MountOption) interface{} {
	return o.options[opt]
}

func (o *OptionHolder) GetString(opt fs.MountOption) (val string) {
	if o.options[opt] != nil {
		val = o.options[opt].(string)
	}
	return
}

func (o *OptionHolder) GetUint32(opt fs.MountOption) (val uint32) {
	if o.options[opt] != nil {
		val = o.options[opt].(uint32)
	}
	return
}

func (o *OptionHolder) GetUint64(opt fs.MountOption) (val uint64) {
	if o.options[opt] != nil {
		val = o.options[opt].(uint64)
	}
	return
}

func (o *OptionHolder) IsSet(opt fs.MountOption) bool {
	return o.options[opt] != nil
}

func (o *OptionHolder) Set(opt fs.MountOption, val interface{}) {
	o.options[opt] = val
}

func (o *OptionHolder) Unset(opt fs.MountOption) {
	delete(o.options, opt)
}
