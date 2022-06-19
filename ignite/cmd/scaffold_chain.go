package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui/clispinner"
	"github.com/ignite-hq/cli/ignite/pkg/placeholder"
	"github.com/ignite-hq/cli/ignite/services/scaffolder"
)

const (
	flagNoDefaultModule = "no-module"
)

// NewScaffoldChain 創建新命令來構建基於 Comos-SDK 的區塊鏈。
func NewScaffoldChain() *cobra.Command {
	c := &cobra.Command{
		Use:   "chain [name]",
		Short: "功能齊全的 Cosmos SDK 區塊鏈",
		Long: `創建一個新的特定於應用程序的 Cosmos SDK 區塊鏈.

例如，以下命令將創建一個名為"hello"目錄:

  ignite scaffold chain hello

項目名稱可以是簡單名稱或 URL。該名稱將用作項目的 Go 模塊路徑。項目名稱示例:

  ignite scaffold chain foo
  ignite scaffold chain foo/bar
  ignite scaffold chain example.org/foo
  ignite scaffold chain github.com/username/foo
		
將在當前目錄中創建一個包含源代碼文件的新目錄。要使用不同的路徑，請使用 "--path" flag.

區塊鏈的大部分邏輯都是用自定義模塊編寫的。每個模塊都有效地封裝了一個獨立的功能。按照 Cosmos SDK 約定，自定義模塊存儲在“x/”目錄中。
默認情況下，Ignite 創建一個名稱與項目名稱匹配的模塊。要創建沒有默認模塊的區塊鏈，請使用“--no-module”標誌。
使用“ignite 腳手架模塊”創建項目後可以添加其他模塊"命令.

基於 Cosmos SDK 的區塊鏈上的賬戶地址具有字符串前綴。
例如,Cosmos Hub 區塊鏈使用默認"cosmos"前輟, 所以地址看起來像這樣: "cosmos12fjzdtqfrrve7zyg9sv8j25azw2ua6tvu07ypf". 
要使用自定義地址前綴，請使用 "--address-prefix" flag. 例如:

  ignite scaffold chain foo --address-prefix bar

默認情況下，在編譯區塊鏈的源代碼時，Ignite 會創建一個緩存以加快構建過程. 
要在構建區塊鏈時清除緩存，請使用 "--clear-cache" flag. 您不太可能需要使用它flag.

區塊鏈使用 Cosmos SDK 模塊化區塊鏈框架. 了解有關 Cosmos SDK 的更多信息 https://docs.cosmos.network`,
		Args: cobra.ExactArgs(1),
		RunE: scaffoldChainHandler,
	}

	flagSetClearCache(c)
	c.Flags().AddFlagSet(flagSetAccountPrefixes())
	c.Flags().StringP(flagPath, "p", ".", "在特定路徑中創建項目")
	c.Flags().Bool(flagNoDefaultModule, false, "創建一個沒有默認模塊的項目")

	return c
}

func scaffoldChainHandler(cmd *cobra.Command, args []string) error {
	s := clispinner.New().SetText("創建中,請耐心等待...")
	defer s.Stop()

	var (
		name               = args[0]
		addressPrefix      = getAddressPrefix(cmd)
		appPath            = flagGetPath(cmd)
		noDefaultModule, _ = cmd.Flags().GetBool(flagNoDefaultModule)
	)

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	appdir, err := scaffolder.Init(cacheStorage, placeholder.New(), appPath, name, addressPrefix, noDefaultModule)
	if err != nil {
		return err
	}

	s.Stop()

	path, err := relativePath(appdir)
	if err != nil {
		return err
	}

	message := `
⭐️ 成功創建新區塊鏈 '%[1]v'.
👉 開始使用以下命令:

 %% cd %[1]v
 %% ignite chain serve

文檔: https://docs.ignite.com
`
	fmt.Printf(message, path)

	return nil
}
