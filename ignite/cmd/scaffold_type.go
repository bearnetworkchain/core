package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/services/scaffolder"
)

// NewScaffoldType 返回一個新命令來構建類型。
func NewScaffoldType() *cobra.Command {
	c := &cobra.Command{
		Use:   "type NAME [field]...",
		Short: "腳手架只有一個類型定義",
		Args:  cobra.MinimumNArgs(1),
		RunE:  scaffoldTypeHandler,
	}

	flagSetPath(c)
	flagSetClearCache(c)
	c.Flags().AddFlagSet(flagSetScaffoldType())

	return c
}

func scaffoldTypeHandler(cmd *cobra.Command, args []string) error {
	return scaffoldType(cmd, args, scaffolder.DryType())
}
