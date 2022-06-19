package cosmosgen

import (
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
	"github.com/pkg/errors"

	"github.com/ignite-hq/cli/ignite/pkg/protoanalysis"
	"github.com/ignite-hq/cli/ignite/pkg/protoc"
)

var (
	goOuts = []string{
		"--gocosmos_out=plugins=interfacetype+grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:.",
		"--grpc-gateway_out=logtostderr=true:.",
	}
)

func (g *generator) generateGo() error {
	includePaths, err := g.resolveInclude(g.appPath)
	if err != nil {
		return err
	}

	// 創建了一個臨時目錄來定位生成的代碼，稍後只有其中一些會被移動到
	// 應用程序的源代碼。這也可以防止在應用程序的源代碼或其父目錄中存在剩余文件 -when
	// 直接在那裡執行的命令——在中斷的情況下。
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	// 在應用程序中發現 proto 包。
	pp := filepath.Join(g.appPath, g.protoDir)
	pkgs, err := protoanalysis.Parse(g.ctx, nil, pp)
	if err != nil {
		return err
	}

	// 為每個模塊生成代碼。
	for _, pkg := range pkgs {
		if err := protoc.Generate(g.ctx, tmp, pkg.Path, includePaths, goOuts); err != nil {
			return err
		}
	}

	// 將應用程序生成的代碼移動到其源代碼中的相對位置下。
	generatedPath := filepath.Join(tmp, g.o.gomodPath)

	_, err = os.Stat(generatedPath)
	if err == nil {
		err = copy.Copy(generatedPath, g.appPath)
		if err != nil {
			return errors.Wrap(err, "無法複製路徑")
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}
