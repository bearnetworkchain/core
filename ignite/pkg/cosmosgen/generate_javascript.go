package cosmosgen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/sync/errgroup"

	"github.com/bearnetworkchain/core/ignite/pkg/cache"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosanalysis/module"
	"github.com/bearnetworkchain/core/ignite/pkg/dirchange"
	"github.com/bearnetworkchain/core/ignite/pkg/gomodulepath"
	"github.com/bearnetworkchain/core/ignite/pkg/localfs"
	"github.com/bearnetworkchain/core/ignite/pkg/nodetime/programs/sta"
	tsproto "github.com/bearnetworkchain/core/ignite/pkg/nodetime/programs/ts-proto"
	"github.com/bearnetworkchain/core/ignite/pkg/protoc"
	"github.com/bearnetworkchain/core/ignite/pkg/xstrings"
)

var (
	tsOut = []string{
		"--ts_proto_out=.",
	}

	jsOpenAPIOut = []string{
		"--openapiv2_out=logtostderr=true,allow_merge=true,json_names_for_fields=false,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:.",
	}
)

const (
	vuexRootMarker          = "vuex-root"
	dirchangeCacheNamespace = "generate.javascript.dirchange"
)

type jsGenerator struct {
	g *generator
}

func newJSGenerator(g *generator) *jsGenerator {
	return &jsGenerator{
		g: g,
	}
}

func (g *generator) generateJS() error {
	jsg := newJSGenerator(g)

	if err := jsg.generateModules(); err != nil {
		return err
	}

	return jsg.generateVuexModuleLoader()
}

func (g *jsGenerator) generateModules() error {
	tsprotoPluginPath, cleanup, err := tsproto.BinaryPath()
	if err != nil {
		return err
	}
	defer cleanup()

	gg := &errgroup.Group{}

	dirCache := cache.New[[]byte](g.g.cacheStorage, dirchangeCacheNamespace)
	add := func(sourcePath string, modules []module.Module) {
		for _, m := range modules {
			m := m
			gg.Go(func() error {
				cacheKey := m.Pkg.Path
				paths := append([]string{m.Pkg.Path, g.g.o.jsOut(m)}, g.g.o.includeDirs...)
				changed, err := dirchange.HasDirChecksumChanged(dirCache, cacheKey, sourcePath, paths...)
				if err != nil {
					return err
				}

				if !changed {
					return nil
				}

				if err := g.generateModule(g.g.ctx, tsprotoPluginPath, sourcePath, m); err != nil {
					return err
				}

				return dirchange.SaveDirChecksum(dirCache, cacheKey, sourcePath, paths...)
			})
		}
	}

	add(g.g.appPath, g.g.appModules)

	if g.g.o.jsIncludeThirdParty {
		for sourcePath, modules := range g.g.thirdModules {
			add(sourcePath, modules)
		}
	}

	return gg.Wait()
}

// generateModule 產生為模塊生成 JS 代碼。
func (g *jsGenerator) generateModule(ctx context.Context, tsprotoPluginPath, appPath string, m module.Module) error {
	var (
		out          = g.g.o.jsOut(m)
		storeDirPath = filepath.Dir(out)
		typesOut     = filepath.Join(out, "types")
	)

	includePaths, err := g.g.resolveInclude(appPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(typesOut, 0766); err != nil {
		return err
	}

	// generate ts-first 類型。
	err = protoc.Generate(
		g.g.ctx,
		typesOut,
		m.Pkg.Path,
		includePaths,
		tsOut,
		protoc.Plugin(tsprotoPluginPath, "--ts_proto_opt=snakeToCamel=false"),
		protoc.Env("NODE_OPTIONS="), // 取消設置 nodejs 選項以避免 vercel "pkg" 出現意外問題
	)
	if err != nil {
		return err
	}

	// 生成 OpenAPI 規範。
	oaitemp, err := os.MkdirTemp("", "gen-js-openapi-module-spec")
	if err != nil {
		return err
	}
	defer os.RemoveAll(oaitemp)

	err = protoc.Generate(
		ctx,
		oaitemp,
		m.Pkg.Path,
		includePaths,
		jsOpenAPIOut,
	)
	if err != nil {
		return err
	}

	// 從 OpenAPI 規範生成 REST 客戶端。
	var (
		srcspec = filepath.Join(oaitemp, "apidocs.swagger.json")
		outREST = filepath.Join(out, "rest.ts")
	)

	if err := sta.Generate(g.g.ctx, outREST, srcspec, "-1"); err != nil { // -1 removes the route namespace.
		return err
	}

	// 生成 js 客戶端包裝器。
	pp := filepath.Join(appPath, g.g.protoDir)
	if err := templateJSClient.Write(out, pp, struct{ Module module.Module }{m}); err != nil {
		return err
	}

	// 如果啟用，則生成 Vuex。
	if g.g.o.vuexStoreRootPath != "" {
		err = templateVuexStore.Write(storeDirPath, pp, struct{ Module module.Module }{m})
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *jsGenerator) generateVuexModuleLoader() error {
	modulePaths, err := localfs.Search(g.g.o.vuexStoreRootPath, vuexRootMarker)
	if err != nil {
		return err
	}

	chainPath, _, err := gomodulepath.Find(g.g.appPath)
	if err != nil {
		return err
	}

	appModulePath := gomodulepath.ExtractAppPath(chainPath.RawPath)

	type module struct {
		Name     string
		Path     string
		FullName string
		FullPath string
	}

	data := struct {
		Modules     []module
		PackageName string
	}{
		PackageName: fmt.Sprintf("%s-js", strings.ReplaceAll(appModulePath, "/", "-")),
	}

	for _, path := range modulePaths {
		pathrel, err := filepath.Rel(g.g.o.vuexStoreRootPath, path)
		if err != nil {
			return err
		}

		var (
			fullPath = filepath.Dir(pathrel)
			fullName = xstrings.FormatUsername(strcase.ToCamel(strings.ReplaceAll(fullPath, "/", "_")))
			path     = filepath.Base(fullPath)
			name     = strcase.ToCamel(path)
		)
		data.Modules = append(data.Modules, module{
			Name:     name,
			Path:     path,
			FullName: fullName,
			FullPath: fullPath,
		})
	}

	if err := templateVuexRoot.Write(g.g.o.vuexStoreRootPath, "", data); err != nil {
		return err
	}

	return nil
}
