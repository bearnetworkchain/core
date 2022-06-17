package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/docs"
	"github.com/bearnetworkchain/core/ignite/pkg/localfs"
	"github.com/bearnetworkchain/core/ignite/pkg/markdownviewer"
)

func NewDocs() *cobra.Command {
	c := &cobra.Command{
		Use:   "docs",
		Short: "顯示 Ignite CLI 文檔",
		Args:  cobra.NoArgs,
		RunE:  docsHandler,
	}
	return c
}

func docsHandler(cmd *cobra.Command, args []string) error {
	path, cleanup, err := localfs.SaveTemp(docs.Docs)
	if err != nil {
		return err
	}
	defer cleanup()

	return markdownviewer.View(path)
}
