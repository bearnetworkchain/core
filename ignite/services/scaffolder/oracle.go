package scaffolder

import (
	"context"
	"fmt"

	"github.com/gobuffalo/genny"

	"github.com/ignite-hq/cli/ignite/pkg/cache"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/gocmd"
	"github.com/ignite-hq/cli/ignite/pkg/multiformatname"
	"github.com/ignite-hq/cli/ignite/pkg/placeholder"
	"github.com/ignite-hq/cli/ignite/pkg/xgenny"
	"github.com/ignite-hq/cli/ignite/templates/ibc"
)

const (
	bandImport  = "github.com/bandprotocol/bandchain-packet"
	bandVersion = "v0.0.2"
)

// OracleOption 為 AddOracle 配置選項。
type OracleOption func(*oracleOptions)

type oracleOptions struct {
	signer string
}

// newOracleOptions 返回一個帶有默認選項的 oracleOptions
func newOracleOptions() oracleOptions {
	return oracleOptions{
		signer: "creator",
	}
}

// OracleWithSigner為消息提供自定義簽名者名稱
func OracleWithSigner(signer string) OracleOption {
	return func(m *oracleOptions) {
		m.signer = signer
	}
}

// AddOracle添加了一個新的 BandChain oracle envtest。
func (s *Scaffolder) AddOracle(
	cacheStorage cache.Storage,
	tracer *placeholder.Tracer,
	moduleName,
	queryName string,
	options ...OracleOption,
) (sm xgenny.SourceModification, err error) {
	if err := s.installBandPacket(); err != nil {
		return sm, err
	}

	o := newOracleOptions()
	for _, apply := range options {
		apply(&o)
	}

	mfName, err := multiformatname.NewName(moduleName, multiformatname.NoNumber)
	if err != nil {
		return sm, err
	}
	moduleName = mfName.LowerCase

	name, err := multiformatname.NewName(queryName)
	if err != nil {
		return sm, err
	}

	if err := checkComponentValidity(s.path, moduleName, name, false); err != nil {
		return sm, err
	}

	mfSigner, err := multiformatname.NewName(o.signer, checkForbiddenOracleFieldName)
	if err != nil {
		return sm, err
	}

	// 模塊必須實現 IBC
	ok, err := isIBCModule(s.path, moduleName)
	if err != nil {
		return sm, err
	}
	if !ok {
		return sm, fmt.Errorf("模塊 %s 沒有實現 IBC 模塊接口", moduleName)
	}

	// 生成數據包
	var (
		g    *genny.Generator
		opts = &ibc.OracleOptions{
			AppName:    s.modpath.Package,
			AppPath:    s.path,
			ModulePath: s.modpath.RawPath,
			ModuleName: moduleName,
			QueryName:  name,
			MsgSigner:  mfSigner,
		}
	)
	g, err = ibc.NewOracle(tracer, opts)
	if err != nil {
		return sm, err
	}
	sm, err = xgenny.RunWithValidation(tracer, g)
	if err != nil {
		return sm, err
	}
	return sm, finish(cacheStorage, opts.AppPath, s.modpath.RawPath)
}

func (s Scaffolder) installBandPacket() error {
	return cmdrunner.New().
		Run(context.Background(),
			step.New(step.Exec(gocmd.Name(), "get", gocmd.PackageLiteral(bandImport, bandVersion))),
		)
}
