package scaffolder

import (
	"context"
	"fmt"
	"strings"

	"github.com/gobuffalo/genny"

	"github.com/ignite-hq/cli/ignite/pkg/cache"
	"github.com/ignite-hq/cli/ignite/pkg/multiformatname"
	"github.com/ignite-hq/cli/ignite/pkg/placeholder"
	"github.com/ignite-hq/cli/ignite/pkg/xgenny"
	"github.com/ignite-hq/cli/ignite/templates/field"
	"github.com/ignite-hq/cli/ignite/templates/field/datatype"
	modulecreate "github.com/ignite-hq/cli/ignite/templates/module/create"
	"github.com/ignite-hq/cli/ignite/templates/typed"
	"github.com/ignite-hq/cli/ignite/templates/typed/dry"
	"github.com/ignite-hq/cli/ignite/templates/typed/list"
	maptype "github.com/ignite-hq/cli/ignite/templates/typed/map"
	"github.com/ignite-hq/cli/ignite/templates/typed/singleton"
)

// AddTypeOption 配置 AddType 的選項。
type AddTypeOption func(*addTypeOptions)

// AddTypeKind 為 AddType 配置類型種類選項。
type AddTypeKind func(*addTypeOptions)

type addTypeOptions struct {
	moduleName string
	fields     []string

	isList      bool
	isMap       bool
	isSingleton bool

	indexes []string

	withoutMessage    bool
	withoutSimulation bool
	signer            string
}

// newAddTypeOptions 返回帶有默認選項的 addTypeOptions
func newAddTypeOptions(moduleName string) addTypeOptions {
	return addTypeOptions{
		moduleName: moduleName,
		signer:     "creator",
	}
}

// ListType 使類型存儲在存儲中的列表約定中。
func ListType() AddTypeKind {
	return func(o *addTypeOptions) {
		o.isList = true
	}
}

// MapType 使類型存儲在具有自定義的存儲中的鍵值約定中
// 指數期權。
func MapType(indexes ...string) AddTypeKind {
	return func(o *addTypeOptions) {
		o.isMap = true
		o.indexes = indexes
	}
}

// SingletonType 使存儲在固定位置的類型作為存儲中的單個條目。
func SingletonType() AddTypeKind {
	return func(o *addTypeOptions) {
		o.isSingleton = true
	}
}

// DryType 僅創建具有基本定義的類型。
func DryType() AddTypeKind {
	return func(o *addTypeOptions) {}
}

// TypeWithModule 腳手架類型的模塊。
func TypeWithModule(name string) AddTypeOption {
	return func(o *addTypeOptions) {
		o.moduleName = name
	}
}

// TypeWithFields 將字段添加到要搭建的類型。
func TypeWithFields(fields ...string) AddTypeOption {
	return func(o *addTypeOptions) {
		o.fields = fields
	}
}

// TypeWithoutMessage 禁用生成 sdk 兼容消息和 tx 相關 API。
func TypeWithoutMessage() AddTypeOption {
	return func(o *addTypeOptions) {
		o.withoutMessage = true
	}
}

// TypeWithoutSimulation 禁用生成消息模擬。
func TypeWithoutSimulation() AddTypeOption {
	return func(o *addTypeOptions) {
		o.withoutSimulation = true
	}
}

// TypeWithSigner 為消息提供自定義簽名者名稱
func TypeWithSigner(signer string) AddTypeOption {
	return func(o *addTypeOptions) {
		o.signer = signer
	}
}

// AddType 將新類型添加到腳手架應用程序。
// 如果未給出列表、映射或單例，則為沒有任何額外內容的干類型（如存儲層、模型、CLI 等）
// 將被腳手架。
// 如果沒有給出模塊，該類型將在應用程序的默認模塊中搭建。
func (s Scaffolder) AddType(
	ctx context.Context,
	cacheStorage cache.Storage,
	typeName string,
	tracer *placeholder.Tracer,
	kind AddTypeKind,
	options ...AddTypeOption,
) (sm xgenny.SourceModification, err error) {
// 應用選項。
	o := newAddTypeOptions(s.modpath.Package)
	for _, apply := range append(options, AddTypeOption(kind)) {
		apply(&o)
	}

	mfName, err := multiformatname.NewName(o.moduleName, multiformatname.NoNumber)
	if err != nil {
		return sm, err
	}
	moduleName := mfName.LowerCase

	name, err := multiformatname.NewName(typeName)
	if err != nil {
		return sm, err
	}

	if err := checkComponentValidity(s.path, moduleName, name, o.withoutMessage); err != nil {
		return sm, err
	}

	signer := ""
	if !o.withoutMessage {
		signer = o.signer
	}

// 檢查並解析提供的字段
	if err := checkCustomTypes(ctx, s.path, moduleName, o.fields); err != nil {
		return sm, err
	}
	tFields, err := field.ParseFields(o.fields, checkForbiddenTypeField, signer)
	if err != nil {
		return sm, err
	}

	mfSigner, err := multiformatname.NewName(o.signer)
	if err != nil {
		return sm, err
	}

	isIBC, err := isIBCModule(s.path, moduleName)
	if err != nil {
		return sm, err
	}

	var (
		g    *genny.Generator
		opts = &typed.Options{
			AppName:      s.modpath.Package,
			AppPath:      s.path,
			ModulePath:   s.modpath.RawPath,
			ModuleName:   moduleName,
			TypeName:     name,
			Fields:       tFields,
			NoMessage:    o.withoutMessage,
			NoSimulation: o.withoutSimulation,
			MsgSigner:    mfSigner,
			IsIBC:        isIBC,
		}
		gens []*genny.Generator
	)
// 檢查並支持 MsgServer 約定
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

	gens, err = supportGenesisTests(
		gens,
		opts.AppPath,
		opts.AppName,
		opts.ModulePath,
		opts.ModuleName,
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

// 根據模型創建類型生成器
	switch {
	case o.isList:
		g, err = list.NewStargate(tracer, opts)
	case o.isMap:
		g, err = mapGenerator(tracer, opts, o.indexes)
	case o.isSingleton:
		g, err = singleton.NewStargate(tracer, opts)
	default:
		g, err = dry.NewStargate(opts)
	}
	if err != nil {
		return sm, err
	}

// 運行生成
	gens = append(gens, g)
	sm, err = xgenny.RunWithValidation(tracer, gens...)
	if err != nil {
		return sm, err
	}

	return sm, finish(cacheStorage, opts.AppPath, s.modpath.RawPath)
}

// 如果名稱被禁止作為字段名稱，則 checkForbiddenTypeIndex 返回 true
func checkForbiddenTypeIndex(name string) error {
	fieldSplit := strings.Split(name, datatype.Separator)
	if len(fieldSplit) > 1 {
		name = fieldSplit[0]
		fieldType := datatype.Name(fieldSplit[1])
		if f, ok := datatype.SupportedTypes[fieldType]; !ok || f.NonIndex {
			return fmt.Errorf("invalid index type %s", fieldType)
		}
	}
	return checkForbiddenTypeField(name)
}

// 如果名稱被禁止作為字段名稱，則 checkForbiddenTypeField 返回 true
func checkForbiddenTypeField(name string) error {
	mfName, err := multiformatname.NewName(name)
	if err != nil {
		return err
	}

	switch mfName.LowerCase {
	case
		"id",
		"params",
		"appendedvalue",
		datatype.TypeCustom:
		return fmt.Errorf("%s 由類型腳手架使用", name)
	}

	return checkGoReservedWord(name)
}

//mapGenerator 返回地圖的模板生成器
func mapGenerator(replacer placeholder.Replacer, opts *typed.Options, indexes []string) (*genny.Generator, error) {
	//使用關聯類型解析索引
	parsedIndexes, err := field.ParseFields(indexes, checkForbiddenTypeIndex)
	if err != nil {
		return nil, err
	}

	// 索引和類型字段必須是不相交的
	exists := make(map[string]struct{})
	for _, name := range opts.Fields {
		exists[name.Name.LowerCamel] = struct{}{}
	}
	for _, index := range parsedIndexes {
		if _, ok := exists[index.Name.LowerCamel]; ok {
			return nil, fmt.Errorf("%s 不能同時是索引和字段", index.Name.Original)
		}
	}

	opts.Indexes = parsedIndexes
	return maptype.NewStargate(replacer, opts)
}
