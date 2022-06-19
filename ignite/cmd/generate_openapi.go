package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui/clispinner"
	"github.com/ignite-hq/cli/ignite/services/chain"
)

func NewGenerateOpenAPI() *cobra.Command {
	return &cobra.Command{
		Use:   "openapi",
		Short: "從您的鏈中生成一個OpenAPI規範的config.yml",
		RunE:  generateOpenAPIHandler,
	}
}

func generateOpenAPIHandler(cmd *cobra.Command, args []string) error {
	s := clispinner.New().SetText("產生中...")
	defer s.Stop()

	c, err := newChainWithHomeFlags(cmd)
	if err != nil {
		return err
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	if err := c.Generate(cmd.Context(), cacheStorage, chain.GenerateOpenAPI()); err != nil {
		return err
	}

	s.Stop()
	fmt.Println("⛏️  生成OpenAPI規範.")

	return nil
}
