package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui/clispinner"
	"github.com/ignite-hq/cli/ignite/pkg/placeholder"
)

const (
	flagPaginated = "paginated"
)

//NewScaffoldQuery 命令創建一個新的類型命令來構建查詢
func NewScaffoldQuery() *cobra.Command {
	c := &cobra.Command{
		Use:   "query [name] [request_field1] [request_field2] ...",
		Short: "查詢從區塊鏈獲取數據",
		Args:  cobra.MinimumNArgs(1),
		RunE:  queryHandler,
	}

	flagSetPath(c)
	flagSetClearCache(c)
	c.Flags().String(flagModule, "", "將查詢添加到的模塊。默認值：應用程序的主模塊")
	c.Flags().StringSliceP(flagResponse, "r", []string{}, "響應字段")
	c.Flags().StringP(flagDescription, "d", "", "命令說明")
	c.Flags().Bool(flagPaginated, false, "定義請求是否可以分頁")

	return c
}

func queryHandler(cmd *cobra.Command, args []string) error {
	appPath := flagGetPath(cmd)

	s := clispinner.New().SetText("創建中,請耐心等候...")
	defer s.Stop()

	// 獲取要添加類型的模塊
	module, err := cmd.Flags().GetString(flagModule)
	if err != nil {
		return err
	}

	// 獲取請求字段
	resFields, err := cmd.Flags().GetStringSlice(flagResponse)
	if err != nil {
		return err
	}

	// 獲取描述
	desc, err := cmd.Flags().GetString(flagDescription)
	if err != nil {
		return err
	}
	if desc == "" {
		// 使用默認描述
		desc = fmt.Sprintf("Query %s", args[0])
	}

	paginated, err := cmd.Flags().GetBool(flagPaginated)
	if err != nil {
		return err
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	sc, err := newApp(appPath)
	if err != nil {
		return err
	}

	sm, err := sc.AddQuery(cmd.Context(), cacheStorage, placeholder.New(), module, args[0], desc, args[1:], resFields, paginated)
	if err != nil {
		return err
	}

	s.Stop()

	modificationsStr, err := sourceModificationToString(sm)
	if err != nil {
		return err
	}

	fmt.Println(modificationsStr)
	fmt.Printf("\n🎉 創建了一個查詢 `%[1]v`.\n\n", args[0])

	return nil
}
