package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/services/scaffolder"
)

// NewScaffoldSingle 返回一個新命令來搭建單例。
func NewScaffoldSingle() *cobra.Command {
	c := &cobra.Command{
		Use:   "single NAME [field]...",
		Short: "CRUD 用於存儲在單個位置的數據",
		Args:  cobra.MinimumNArgs(1),
		RunE:  scaffoldSingleHandler,
	}

	flagSetPath(c)
	flagSetClearCache(c)
	c.Flags().AddFlagSet(flagSetScaffoldType())

	return c
}

func scaffoldSingleHandler(cmd *cobra.Command, args []string) error {
	return scaffoldType(cmd, args, scaffolder.SingletonType())
}
