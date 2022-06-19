package cosmosgen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/iancoleman/strcase"

	"github.com/ignite-hq/cli/ignite/pkg/cache"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosanalysis/module"
	"github.com/ignite-hq/cli/ignite/pkg/dirchange"
	swaggercombine "github.com/ignite-hq/cli/ignite/pkg/nodetime/programs/swagger-combine"
	"github.com/ignite-hq/cli/ignite/pkg/protoc"
)

var openAPIOut = []string{
	"--openapiv2_out=logtostderr=true,allow_merge=true,json_names_for_fields=false,fqn_for_openapi_name=true,simple_operation_ids=true,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:.",
}

const specCacheNamespace = "generate.openapi.spec"

func generateOpenAPISpec(g *generator) error {
	out := filepath.Join(g.appPath, g.o.specOut)

	var (
		specDirs []string
		conf     = swaggercombine.Config{
			Swagger: "2.0",
			Info: swaggercombine.Info{
				Title: "HTTP API Console",
			},
		}
	)

	defer func() {
		for _, dir := range specDirs {
			os.RemoveAll(dir)
		}
	}()

	specCache := cache.New[[]byte](g.cacheStorage, specCacheNamespace)

	var hasAnySpecChanged bool

	// gen 為其源代碼位於 src 的模塊生成規範。
	// 並為其添加所需的 swaggercombine 配置。
	gen := func(src string, m module.Module) (err error) {
		dir, err := os.MkdirTemp("", "gen-openapi-module-spec")
		if err != nil {
			return err
		}
		specPath := filepath.Join(dir, "apidocs.swagger.json")

		checksumPaths := append([]string{m.Pkg.Path}, g.o.includeDirs...)
		checksum, err := dirchange.ChecksumFromPaths(src, checksumPaths...)
		if err != nil {
			return err
		}
		cacheKey := fmt.Sprintf("%x", checksum)
		existingSpec, err := specCache.Get(cacheKey)
		if err != nil && err != cache.ErrorNotFound {
			return err
		}

		if err != cache.ErrorNotFound {
			if err := os.WriteFile(specPath, existingSpec, 0644); err != nil {
				return err
			}
		} else {
			hasAnySpecChanged = true
			include, err := g.resolveInclude(src)
			if err != nil {
				return err
			}

			err = protoc.Generate(
				g.ctx,
				dir,
				m.Pkg.Path,
				include,
				openAPIOut,
			)
			if err != nil {
				return err
			}

			f, err := os.ReadFile(specPath)
			if err != nil {
				return err
			}
			if err := specCache.Put(cacheKey, f); err != nil {
				return err
			}
		}

		specDirs = append(specDirs, dir)

		return conf.AddSpec(strcase.ToCamel(m.Pkg.Name), specPath)
	}

	// 為每個模塊生成規範並將它們保存在文件系統中
	// 在將它們的路徑和配置添加到 swaggercombine.Config 之後，我們可以將它們組合起來
	// 進入單個規範。

	add := func(src string, modules []module.Module) error {
		for _, m := range modules {
			m := m
			if err := gen(src, m); err != nil {
				return err
			}
		}
		return nil
	}

	// protoc openapi 生成器在並發運行時表現得很奇怪，所以不要在這裡使用 goroutines。
	if err := add(g.appPath, g.appModules); err != nil {
		return err
	}

	for src, modules := range g.thirdModules {
		if err := add(src, modules); err != nil {
			return err
		}
	}

	if !hasAnySpecChanged {
		// 如果生成的輸出已更改
		changed, err := dirchange.HasDirChecksumChanged(specCache, out, g.appPath, out)
		if err != nil {
			return err
		}

		if !changed {
			return nil
		}
	}

	sort.Slice(conf.APIs, func(a, b int) bool { return conf.APIs[a].ID < conf.APIs[b].ID })

	// 確保存在目錄。
	outDir := filepath.Dir(out)
	if err := os.MkdirAll(outDir, 0766); err != nil {
		return err
	}

	// 將規格合二為一併保存到外面。
	if err := swaggercombine.Combine(g.ctx, conf, out); err != nil {
		return err
	}

	return dirchange.SaveDirChecksum(specCache, out, g.appPath, out)
}
