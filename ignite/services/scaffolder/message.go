package scaffolder

import (
	"context"
	"fmt"

	"github.com/gobuffalo/genny"

	"github.com/bearnetworkchain/core/ignite/pkg/cache"
	"github.com/bearnetworkchain/core/ignite/pkg/multiformatname"
	"github.com/bearnetworkchain/core/ignite/pkg/placeholder"
	"github.com/bearnetworkchain/core/ignite/pkg/xgenny"
	"github.com/bearnetworkchain/core/ignite/templates/field"
	"github.com/bearnetworkchain/core/ignite/templates/field/datatype"
	"github.com/bearnetworkchain/core/ignite/templates/message"
	modulecreate "github.com/bearnetworkchain/core/ignite/templates/module/create"
)

// messageOptions表示消息腳手架的配置
type messageOptions struct {
	description       string
	signer            string
	withoutSimulation bool
}

// newMessageOptions返回帶有默認選項的 messageOptions
func newMessageOptions(messageName string) messageOptions {
	return messageOptions{
		description: fmt.Sprintf("廣播消息 %s", messageName),
		signer:      "creator",
	}
}

// MessageOption 配置消息腳手架
type MessageOption func(*messageOptions)

// WithDescription 為消息 CLI 命令提供自定義描述
func WithDescription(desc string) MessageOption {
	return func(m *messageOptions) {
		m.description = desc
	}
}

// WithSigner 為消息提供自定義簽名者名稱
func WithSigner(signer string) MessageOption {
	return func(m *messageOptions) {
		m.signer = signer
	}
}

// WithoutSimulation 禁用生成消息模擬
func WithoutSimulation() MessageOption {
	return func(m *messageOptions) {
		m.withoutSimulation = true
	}
}

// AddMessage 向腳手架應用程序添加新消息
func (s Scaffolder) AddMessage(
	ctx context.Context,
	cacheStorage cache.Storage,
	tracer *placeholder.Tracer,
	moduleName,
	msgName string,
	fields,
	resFields []string,
	options ...MessageOption,
) (sm xgenny.SourceModification, err error) {
	// 創建選項
	scaffoldingOpts := newMessageOptions(msgName)
	for _, apply := range options {
		apply(&scaffoldingOpts)
	}

	// 如果沒有提供模塊，我們將類型添加到應用程序的模塊中
	if moduleName == "" {
		moduleName = s.modpath.Package
	}
	mfName, err := multiformatname.NewName(moduleName, multiformatname.NoNumber)
	if err != nil {
		return sm, err
	}
	moduleName = mfName.LowerCase

	name, err := multiformatname.NewName(msgName)
	if err != nil {
		return sm, err
	}

	if err := checkComponentValidity(s.path, moduleName, name, false); err != nil {
		return sm, err
	}

	// 檢查和解析提供的字段
	if err := checkCustomTypes(ctx, s.path, moduleName, fields); err != nil {
		return sm, err
	}
	parsedMsgFields, err := field.ParseFields(fields, checkForbiddenMessageField, scaffoldingOpts.signer)
	if err != nil {
		return sm, err
	}

	//檢查並解析提供的響應字段
	if err := checkCustomTypes(ctx, s.path, moduleName, resFields); err != nil {
		return sm, err
	}
	parsedResFields, err := field.ParseFields(resFields, checkGoReservedWord, scaffoldingOpts.signer)
	if err != nil {
		return sm, err
	}

	mfSigner, err := multiformatname.NewName(scaffoldingOpts.signer)
	if err != nil {
		return sm, err
	}

	var (
		g    *genny.Generator
		opts = &message.Options{
			AppName:      s.modpath.Package,
			AppPath:      s.path,
			ModulePath:   s.modpath.RawPath,
			ModuleName:   moduleName,
			MsgName:      name,
			Fields:       parsedMsgFields,
			ResFields:    parsedResFields,
			MsgDesc:      scaffoldingOpts.description,
			MsgSigner:    mfSigner,
			NoSimulation: scaffoldingOpts.withoutSimulation,
		}
	)

	//檢查並支持 MsgServer 約定
	var gens []*genny.Generator
	gens, err = supportMsgServer(
		gens,
		tracer,
		s.path,
		&modulecreate.MsgServerOptions{
			ModuleName: opts.ModuleName,
			ModulePath: opts.ModulePath,
			AppName:    opts.AppName,
			AppPath:    opts.AppPath,
		},
	)
	if err != nil {
		return sm, err
	}

	gens, err = supportSimulation(
		gens,
		opts.AppPath,
		opts.ModulePath,
		opts.ModuleName,
	)
	if err != nil {
		return sm, err
	}

	// 腳手架
	g, err = message.NewStargate(tracer, opts)
	if err != nil {
		return sm, err
	}
	gens = append(gens, g)
	sm, err = xgenny.RunWithValidation(tracer, gens...)
	if err != nil {
		return sm, err
	}
	return sm, finish(cacheStorage, opts.AppPath, s.modpath.RawPath)
}

// checkForbiddenMessageField如果名稱被禁止作為消息名稱，則返回 true
func checkForbiddenMessageField(name string) error {
	mfName, err := multiformatname.NewName(name)
	if err != nil {
		return err
	}

	if mfName.LowerCase == datatype.TypeCustom {
		return fmt.Errorf("%s 由消息腳手架使用", name)
	}

	return checkGoReservedWord(name)
}
