package message

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
	MsgName      multiformatname.Name
	MsgSigner    multiformatname.Name
	MsgDesc      string
	Fields       field.Fields
	ResFields    field.Fields
	NoSimulation bool
}

// Validate that options are usuable
func (opts *Options) Validate() error {
	return nil
}
