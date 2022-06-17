package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/services/scaffolder"
)

const (
	FlagIndexes = "index"
)

// NewScaffoldMap 返回一個新命令來構建地圖。
func NewScaffoldMap() *cobra.Command {
	c := &cobra.Command{
		Use:   "map NAME [field]...",
		Short: "以鍵值對形式存儲的數據的CRUD",
		Args:  cobra.MinimumNArgs(1),
		RunE:  scaffoldMapHandler,
	}

	flagSetPath(c)
	flagSetClearCache(c)
	c.Flags().AddFlagSet(flagSetScaffoldType())
	c.Flags().StringSlice(FlagIndexes, []string{"index"}, "索引值的字段")

	return c
}

func scaffoldMapHandler(cmd *cobra.Command, args []string) error {
	indexes, err := cmd.Flags().GetStringSlice(FlagIndexes)
	if err != nil {
		return err
	}

	return scaffoldType(cmd, args, scaffolder.MapType(indexes...))
}
