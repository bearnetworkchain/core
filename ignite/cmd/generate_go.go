package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/cliui/clispinner"
	"github.com/bearnetworkchain/core/ignite/services/chain"
)

func NewGenerateGo() *cobra.Command {
	return &cobra.Command{
		Use:   "proto-go",
		Short: "生成應用程序源代碼所需的原型Golang代碼",
		RunE:  generateGoHandler,
	}
}

func generateGoHandler(cmd *cobra.Command, args []string) error {
	s := clispinner.New().SetText("生成中,抽根菸,喝口飲料,稍等一下...")
	defer s.Stop()

	c, err := newChainWithHomeFlags(cmd)
	if err != nil {
		return err
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	if err := c.Generate(cmd.Context(), cacheStorage, chain.GenerateGo()); err != nil {
		return err
	}

	s.Stop()
	fmt.Println("⛏️  生成Golang代碼.")

	return nil
}
