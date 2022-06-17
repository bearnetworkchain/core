package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/services/scaffolder"
)

// NewScaffoldList 返回一個新命令來構建列表。
func NewScaffoldList() *cobra.Command {
	c := &cobra.Command{
		Use:   "list NAME [field]...",
		Short: "CRUD 用於存儲為數組的數據",
		Args:  cobra.MinimumNArgs(1),
		RunE:  scaffoldListHandler,
	}

	flagSetPath(c)
	flagSetClearCache(c)
	c.Flags().AddFlagSet(flagSetScaffoldType())

	return c
}

func scaffoldListHandler(cmd *cobra.Command, args []string) error {
	return scaffoldType(cmd, args, scaffolder.ListType())
}
