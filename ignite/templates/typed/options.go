package typed

import (
	"github.com/bearnetworkchain/core/ignite/pkg/multiformatname"
	"github.com/bearnetworkchain/core/ignite/templates/field"
)

// Options ...
type Options struct {
	AppName      string
	AppPath      string
	ModuleName   string
	ModulePath   string
	TypeName     multiformatname.Name
	MsgSigner    multiformatname.Name
	Fields       field.Fields
	Indexes      field.Fields
	NoMessage    bool
	NoSimulation bool
	IsIBC        bool
}

// Validate that options are usable
func (opts *Options) Validate() error {
	return nil
}
