package scaffolder

import (
	"context"
	"errors"

	"github.com/gobuffalo/genny"

	"github.com/ignite-hq/cli/ignite/pkg/cache"
	"github.com/ignite-hq/cli/ignite/pkg/multiformatname"
	"github.com/ignite-hq/cli/ignite/pkg/placeholder"
	"github.com/ignite-hq/cli/ignite/pkg/xgenny"
	"github.com/ignite-hq/cli/ignite/templates/field"
	"github.com/ignite-hq/cli/ignite/templates/query"
)

//AddQuery 將新查詢添加到腳手架應用程序
func (s Scaffolder) AddQuery(
	ctx context.Context,
	cacheStorage cache.Storage,
	tracer *placeholder.Tracer,
	moduleName,
	queryName,
	description string,
	reqFields,
	resFields []string,
	paginated bool,
) (sm xgenny.SourceModification, err error) {
	// 如果沒有提供模塊，我們將類型添加到應用程序的模塊中
	if moduleName == "" {
		moduleName = s.modpath.Package
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

	if err := checkComponentValidity(s.path, moduleName, name, true); err != nil {
		return sm, err
	}

	// 檢查並解析提供的請求字段
	if ok := containCustomTypes(reqFields); ok {
		return sm, errors.New("查詢請求參數不能包含自定義類型")
	}
	parsedReqFields, err := field.ParseFields(reqFields, checkGoReservedWord)
	if err != nil {
		return sm, err
	}

	// 檢查並解析提供的響應字段
	if err := checkCustomTypes(ctx, s.path, moduleName, resFields); err != nil {
		return sm, err
	}
	parsedResFields, err := field.ParseFields(resFields, checkGoReservedWord)
	if err != nil {
		return sm, err
	}

	var (
		g    *genny.Generator
		opts = &query.Options{
			AppName:     s.modpath.Package,
			AppPath:     s.path,
			ModulePath:  s.modpath.RawPath,
			ModuleName:  moduleName,
			QueryName:   name,
			ReqFields:   parsedReqFields,
			ResFields:   parsedResFields,
			Description: description,
			Paginated:   paginated,
		}
	)

	// 腳手架
	g, err = query.NewStargate(tracer, opts)
	if err != nil {
		return sm, err
	}
	sm, err = xgenny.RunWithValidation(tracer, g)
	if err != nil {
		return sm, err
	}
	return sm, finish(cacheStorage, opts.AppPath, s.modpath.RawPath)
}
