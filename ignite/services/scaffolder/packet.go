package scaffolder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gobuffalo/genny"

	"github.com/bearnetworkchain/core/ignite/pkg/cache"
	"github.com/bearnetworkchain/core/ignite/pkg/multiformatname"
	"github.com/bearnetworkchain/core/ignite/pkg/placeholder"
	"github.com/bearnetworkchain/core/ignite/pkg/xgenny"
	"github.com/bearnetworkchain/core/ignite/templates/field"
	"github.com/bearnetworkchain/core/ignite/templates/field/datatype"
	"github.com/bearnetworkchain/core/ignite/templates/ibc"
)

const (
	ibcModuleImplementation = "module_ibc.go"
)

// packetOptions 表示數據包腳手架的配置
type packetOptions struct {
	withoutMessage bool
	signer         string
}

// newPacketOptions返回一個帶有默認選項的 packetOptions
func newPacketOptions() packetOptions {
	return packetOptions{
		signer: "creator",
	}
}

// PacketOption 配置數據包腳手架
type PacketOption func(*packetOptions)

// PacketWithoutMessage 禁用生成 sdk 兼容消息和 tx 相關 API。
func PacketWithoutMessage() PacketOption {
	return func(o *packetOptions) {
		o.withoutMessage = true
	}
}

// PacketWithSigner 為數據包提供自定義簽名者名稱
func PacketWithSigner(signer string) PacketOption {
	return func(m *packetOptions) {
		m.signer = signer
	}
}

// AddPacket 使用可選類型字段向腳手架應用程序添加新類型 stype。
func (s Scaffolder) AddPacket(
	ctx context.Context,
	cacheStorage cache.Storage,
	tracer *placeholder.Tracer,
	moduleName,
	packetName string,
	packetFields,
	ackFields []string,
	options ...PacketOption,
) (sm xgenny.SourceModification, err error) {
	// 應用選項。
	o := newPacketOptions()
	for _, apply := range options {
		apply(&o)
	}

	mfName, err := multiformatname.NewName(moduleName, multiformatname.NoNumber)
	if err != nil {
		return sm, err
	}
	moduleName = mfName.LowerCase

	name, err := multiformatname.NewName(packetName)
	if err != nil {
		return sm, err
	}

	if err := checkComponentValidity(s.path, moduleName, name, o.withoutMessage); err != nil {
		return sm, err
	}

	mfSigner, err := multiformatname.NewName(o.signer)
	if err != nil {
		return sm, err
	}

	// Module 必須實施 IBC
	ok, err := isIBCModule(s.path, moduleName)
	if err != nil {
		return sm, err
	}
	if !ok {
		return sm, fmt.Errorf("模塊 %s 沒有實現 IBC 模塊接口", moduleName)
	}

	signer := ""
	if !o.withoutMessage {
		signer = o.signer
	}

	// 檢查和解析數據包字段
	if err := checkCustomTypes(ctx, s.path, moduleName, packetFields); err != nil {
		return sm, err
	}
	parsedPacketFields, err := field.ParseFields(packetFields, checkForbiddenPacketField, signer)
	if err != nil {
		return sm, err
	}

	// 檢查和解析確認字段
	if err := checkCustomTypes(ctx, s.path, moduleName, ackFields); err != nil {
		return sm, err
	}
	parsedAcksFields, err := field.ParseFields(ackFields, checkGoReservedWord, signer)
	if err != nil {
		return sm, err
	}

	//生成數據包
	var (
		g    *genny.Generator
		opts = &ibc.PacketOptions{
			AppName:    s.modpath.Package,
			AppPath:    s.path,
			ModulePath: s.modpath.RawPath,
			ModuleName: moduleName,
			PacketName: name,
			Fields:     parsedPacketFields,
			AckFields:  parsedAcksFields,
			NoMessage:  o.withoutMessage,
			MsgSigner:  mfSigner,
		}
	)
	g, err = ibc.NewPacket(tracer, opts)
	if err != nil {
		return sm, err
	}
	sm, err = xgenny.RunWithValidation(tracer, g)
	if err != nil {
		return sm, err
	}
	return sm, finish(cacheStorage, opts.AppPath, s.modpath.RawPath)
}

// 如果提供的模塊實現了 IBC 模塊接口，isIBCModule 返回 true
// 我們天真地檢查 module_ibc.go 的存在以進行此檢查
func isIBCModule(appPath string, moduleName string) (bool, error) {
	absPath, err := filepath.Abs(filepath.Join(appPath, moduleDir, moduleName, ibcModuleImplementation))
	if err != nil {
		return false, err
	}

	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		// 不是 IBC 模塊
		return false, nil
	}

	return true, err
}

// 如果名稱被禁止作為數據包名稱，則 checkForbiddenPacketField 返回 true
func checkForbiddenPacketField(name string) error {
	mfName, err := multiformatname.NewName(name)
	if err != nil {
		return err
	}

	switch mfName.LowerCase {
	case
		"sender",
		"port",
		"channelid",
		datatype.TypeCustom:
		return fmt.Errorf("%s 由包腳手架使用", name)
	}

	return checkGoReservedWord(name)
}
