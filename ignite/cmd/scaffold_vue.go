package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui/clispinner"
	"github.com/ignite-hq/cli/ignite/services/scaffolder"
)

// NewScaffoldVue 為鏈搭建了一個 Vue.js 應用程序。
func NewScaffoldVue() *cobra.Command {
	c := &cobra.Command{
		Use:   "vue",
		Short: "Vue 3 網頁應用程序模板",
		Args:  cobra.NoArgs,
		RunE:  scaffoldVueHandler,
	}

	c.Flags().StringP(flagPath, "p", "./vue", "腳手架內容的路徑 Vue.js 應用程序")

	return c
}

func scaffoldVueHandler(cmd *cobra.Command, args []string) error {
	s := clispinner.New().SetText("創建中,請稍等一下...")
	defer s.Stop()

	path := flagGetPath(cmd)
	if err := scaffolder.Vue(path); err != nil {
		return err
	}

	s.Stop()
	fmt.Printf("\n🎉 搭建一個 Vue.js 應用程序.\n\n")

	return nil
}
