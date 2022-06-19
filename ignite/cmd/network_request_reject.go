package ignitecmd

import (
	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/icons"
	"github.com/ignite-hq/cli/ignite/pkg/numbers"
	"github.com/ignite-hq/cli/ignite/services/network"
)

// NewNetworkRequestReject 創建一個新的請求拒絕
// 拒絕鏈請求的命令.
func NewNetworkRequestReject() *cobra.Command {
	c := &cobra.Command{
		Use:     "reject [launch-id] [number<,...>]",
		Aliases: []string{"accept"},
		Short:   "拒絕請求",
		RunE:    networkRequestRejectHandler,
		Args:    cobra.ExactArgs(2),
	}
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetHome())
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	return c
}

func networkRequestRejectHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	nb, err := newNetworkBuilder(cmd, CollectEvents(session.EventBus()))
	if err != nil {
		return err
	}

	//解析啟動 ID
	launchID, err := network.ParseID(args[0])
	if err != nil {
		return err
	}

	// 獲取請求ID列表
	ids, err := numbers.ParseList(args[1])
	if err != nil {
		return err
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	// 提交被拒絕的請求
	reviewals := make([]network.Reviewal, 0)
	for _, id := range ids {
		reviewals = append(reviewals, network.RejectRequest(id))
	}
	if err := n.SubmitRequest(launchID, reviewals...); err != nil {
		return err
	}

	session.StopSpinner()

	return session.Printf("%s 要求(s) %s 被拒絕\n", icons.OK, numbers.List(ids, "#"))
}
