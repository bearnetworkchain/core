package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/cliui/clispinner"
	"github.com/bearnetworkchain/core/ignite/pkg/placeholder"
)

func NewScaffoldWasm() *cobra.Command {
	c := &cobra.Command{
		Use:   "wasm",
		Short: "將 wasm 模塊導入您的應用程序",
		Long:  "為您的區塊鏈添加對 WebAssembly 智能合約的支持",
		Args:  cobra.NoArgs,
		RunE:  scaffoldWasmHandler,
	}

	flagSetPath(c)

	return c
}

func scaffoldWasmHandler(cmd *cobra.Command, args []string) error {
	appPath := flagGetPath(cmd)

	s := clispinner.New().SetText("創建中,請耐心等待....順便去抽根菸,喝口飲料...")
	defer s.Stop()

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	sc, err := newApp(appPath)
	if err != nil {
		return err
	}

	sm, err := sc.ImportModule(cacheStorage, placeholder.New(), "wasm")
	if err != nil {
		return err
	}

	s.Stop()

	modificationsStr, err := sourceModificationToString(sm)
	if err != nil {
		return err
	}

	fmt.Println(modificationsStr)
	fmt.Printf("\n🎉 匯入wasm.\n\n")

	return nil
}
