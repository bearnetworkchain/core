package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui/clispinner"
	"github.com/ignite-hq/cli/ignite/pkg/placeholder"
	"github.com/ignite-hq/cli/ignite/services/scaffolder"
)

const flagSigner = "signer"

// NewScaffoldMessage 返回腳手架消息的命令
func NewScaffoldMessage() *cobra.Command {
	c := &cobra.Command{
		Use:   "message [name] [field1] [field2] ...",
		Short: "在區塊鏈上執行狀態轉換的消息",
		Args:  cobra.MinimumNArgs(1),
		RunE:  messageHandler,
	}

	flagSetPath(c)
	flagSetClearCache(c)
	c.Flags().String(flagModule, "", "將消息添加到的模塊。默認值：應用程序的主模塊")
	c.Flags().StringSliceP(flagResponse, "r", []string{}, "響應字段")
	c.Flags().Bool(flagNoSimulation, false, "禁用 CRUD 模擬腳手架")
	c.Flags().StringP(flagDescription, "d", "", "命令說明")
	c.Flags().String(flagSigner, "", "消息簽名者的標籤（默認：創建者）")

	return c
}

func messageHandler(cmd *cobra.Command, args []string) error {
	var (
		module, _         = cmd.Flags().GetString(flagModule)
		resFields, _      = cmd.Flags().GetStringSlice(flagResponse)
		desc, _           = cmd.Flags().GetString(flagDescription)
		signer            = flagGetSigner(cmd)
		appPath           = flagGetPath(cmd)
		withoutSimulation = flagGetNoSimulation(cmd)
	)

	s := clispinner.New().SetText("創建中,請耐心等待...")
	defer s.Stop()

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	var options []scaffolder.MessageOption

	// 獲取描述
	if desc != "" {
		options = append(options, scaffolder.WithDescription(desc))
	}

	// 獲取簽名者
	if signer != "" {
		options = append(options, scaffolder.WithSigner(signer))
	}

	// 跳過腳手架模擬
	if withoutSimulation {
		options = append(options, scaffolder.WithoutSimulation())
	}

	sc, err := newApp(appPath)
	if err != nil {
		return err
	}

	sm, err := sc.AddMessage(cmd.Context(), cacheStorage, placeholder.New(), module, args[0], args[1:], resFields, options...)
	if err != nil {
		return err
	}

	s.Stop()

	modificationsStr, err := sourceModificationToString(sm)
	if err != nil {
		return err
	}

	fmt.Println(modificationsStr)
	fmt.Printf("\n🎉 創建了一條消息 `%[1]v`.\n\n", args[0])

	return nil
}
