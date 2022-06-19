package ignitecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui/clispinner"
	"github.com/ignite-hq/cli/ignite/services/scaffolder"
)

// NewScaffoldFlutter 為鏈構建了一個 Flutter 應用程序。
func NewScaffoldFlutter() *cobra.Command {
	c := &cobra.Command{
		Use:   "flutter",
		Short: "適用於您的鏈生態的 Flutter 應用",
		Args:  cobra.NoArgs,
		RunE:  scaffoldFlutterHandler,
	}

	c.Flags().StringP(flagPath, "p", "./flutter", "Flutter 應用的腳手架內容的路徑")

	return c
}

func scaffoldFlutterHandler(cmd *cobra.Command, args []string) error {
	s := clispinner.New().SetText("創建中,請耐心等待...")
	defer s.Stop()

	path := flagGetPath(cmd)
	if err := scaffolder.Flutter(path); err != nil {
		return err
	}

	s.Stop()
	fmt.Printf("\n🎉 搭建了一個 Flutter 應用程序.\n\n")

	return nil
}
