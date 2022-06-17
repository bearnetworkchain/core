package cosmosgen

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattn/go-zglob"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/bearnetworkchain/core/ignite/pkg/cosmosanalysis/module"
	"github.com/bearnetworkchain/core/ignite/pkg/protoc"
	protocgendart "github.com/bearnetworkchain/core/ignite/pkg/protoc-gen-dart"
)

var (
	dartOut = []string{
		"--dart_out=grpc:.",
	}
)

const (
	dartExportFileName = "export.dart"
	dartClientDirName  = "client"
)

type dartGenerator struct {
	g *generator
}

func newDartGenerator(g *generator) *dartGenerator {
	return &dartGenerator{
		g: g,
	}
}

func (g *generator) generateDart() error {
	return newDartGenerator(g).generateModules()
}

func (g *dartGenerator) generateModules() error {
	flag, cleanup, err := protocgendart.Flag()
	if err != nil {
		return err
	}
	defer cleanup()

	gg := &errgroup.Group{}

	add := func(sourcePath string, modules []module.Module) {
		for _, m := range modules {
			m := m
			gg.Go(func() error { return g.generateModule(g.g.ctx, flag, sourcePath, m) })
		}
	}

	add(g.g.appPath, g.g.appModules)

	if g.g.o.dartIncludeThirdParty {
		for sourcePath, modules := range g.g.thirdModules {
			add(sourcePath, modules)
		}
	}

	return gg.Wait()
}

func (g *dartGenerator) generateModule(ctx context.Context, plugin, appPath string, m module.Module) error {
	var (
		out       = g.g.o.dartOut(m)
		clientOut = filepath.Join(out, dartClientDirName)
		exportOut = filepath.Join(out, dartExportFileName)
	)

	includePaths, err := g.g.resolveInclude(appPath)
	if err != nil {
		return err
	}

	// 重置目標目錄。
	if err := os.RemoveAll(out); err != nil {
		return err
	}
	if err := os.MkdirAll(clientOut, 0766); err != nil {
		return err
	}

	// generate grpc 客戶端和 protobuf 類型。
	if err := protoc.Generate(
		ctx,
		clientOut,
		m.Pkg.Path,
		includePaths,
		dartOut,
		protoc.Plugin(plugin),
		protoc.GenerateDependencies(),
	); err != nil {
		return err
	}

	// generate 一個導出文件，通過單個入口點導出所有生成的代碼。
	generatedFiles, err := zglob.Glob(filepath.Join(clientOut, "**/*.dart"))
	if err != nil {
		return err
	}

	var exportContent bytes.Buffer
	for _, file := range generatedFiles {
		path, err := filepath.Rel(out, file)
		if err != nil {
			return err
		}
		exportContent.WriteString(fmt.Sprintf("export '%s';\n", path))
	}

	err = os.WriteFile(exportOut, exportContent.Bytes(), 0644)
	return errors.Wrap(err, "無法為模塊創建 Dart 導出文件")
}
