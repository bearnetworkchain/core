package scaffolder

import (
	"os"
	"path/filepath"

	"github.com/gobuffalo/genny"

	"github.com/ignite-hq/cli/ignite/pkg/placeholder"
	modulecreate "github.com/ignite-hq/cli/ignite/templates/module/create"
)

// supportSimulation 檢查 module_simulation.go 是否存在
// 如果沒有，則附加生成器以創建文件
func supportSimulation(
	gens []*genny.Generator,
	appPath,
	modulePath,
	moduleName string,
) ([]*genny.Generator, error) {
	simulation, err := modulecreate.AddSimulation(
		appPath,
		modulePath,
		moduleName,
	)
	if err != nil {
		return gens, err
	}
	gens = append(gens, simulation)
	return gens, nil
}

// supportGenesisTests 檢查 types/genesis_test.go 是否存在
// 如果沒有，則附加生成器以創建文件
func supportGenesisTests(
	gens []*genny.Generator,
	appPath,
	appName,
	modulePath,
	moduleName string,
) ([]*genny.Generator, error) {
	isIBC, err := isIBCModule(appPath, moduleName)
	if err != nil {
		return gens, err
	}
	genesisTest, err := modulecreate.AddGenesisTest(
		appPath,
		appName,
		modulePath,
		moduleName,
		isIBC,
	)
	if err != nil {
		return gens, err
	}
	gens = append(gens, genesisTest)
	return gens, nil
}

// supportMsgServer 檢查模塊是否支持 MsgServer 約定
// 如果不支持，則附加生成器以支持它
// https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-031-msg-service.md
func supportMsgServer(
	gens []*genny.Generator,
	replacer placeholder.Replacer,
	appPath string,
	opts *modulecreate.MsgServerOptions,
) ([]*genny.Generator, error) {
	// 檢查是否使用了約定
	msgServerDefined, err := isMsgServerDefined(appPath, opts.ModuleName)
	if err != nil {
		return nil, err
	}
	if !msgServerDefined {
		// 為模塊打補丁以支持約定
		g, err := modulecreate.AddMsgServerConventionToLegacyModule(replacer, opts)
		if err != nil {
			return nil, err
		}
		gens = append(gens, g)
	}
	return gens, nil
}

// isMsgServerDefined 檢查模塊是否對事務使用 MsgServer 約定
// 這是通過驗證 tx.proto 文件的存在來檢查的
func isMsgServerDefined(appPath, moduleName string) (bool, error) {
	txProto, err := filepath.Abs(filepath.Join(appPath, "proto", moduleName, "tx.proto"))
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(txProto); os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
