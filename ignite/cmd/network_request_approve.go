package ignitecmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/icons"
	"github.com/ignite-hq/cli/ignite/pkg/numbers"
	"github.com/ignite-hq/cli/ignite/services/network"
)

const (
	flagNoVerification = "no-verification"
)

// NewNetworkRequestApprove 創建一個新的請求批准
// 批准鏈請求的命令.
func NewNetworkRequestApprove() *cobra.Command {
	c := &cobra.Command{
		Use:     "approve [launch-id] [number<,...>]",
		Aliases: []string{"accept"},
		Short:   "批准請求",
		RunE:    networkRequestApproveHandler,
		Args:    cobra.ExactArgs(2),
	}

	flagSetClearCache(c)
	c.Flags().Bool(flagNoVerification, false, "批准請求而不驗證它們")
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetHome())
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	return c
}

func networkRequestApproveHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	nb, err := newNetworkBuilder(cmd, CollectEvents(session.EventBus()))
	if err != nil {
		return err
	}

	// 解析啟動 ID
	launchID, err := network.ParseID(args[0])
	if err != nil {
		return err
	}

	// 獲取請求ID列表
	ids, err := numbers.ParseList(args[1])
	if err != nil {
		return err
	}

	// 驗證請求是否有效
	noVerification, err := cmd.Flags().GetBool(flagNoVerification)
	if err != nil {
		return err
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	// 如果必須驗證請求，我們將在臨時目錄中使用請求模擬鏈
	if !noVerification {
		if err := verifyRequest(cmd.Context(), cacheStorage, nb, launchID, ids...); err != nil {
			return errors.Wrap(err, "request(s) not valid")
		}
		session.Printf("%s 要求(s) %s 已驗證\n", icons.OK, numbers.List(ids, "#"))
	}

	// 提交批准的請求
	reviewals := make([]network.Reviewal, 0)
	for _, id := range ids {
		reviewals = append(reviewals, network.ApproveRequest(id))
	}
	if err := n.SubmitRequest(launchID, reviewals...); err != nil {
		return err
	}

	session.StopSpinner()

	return session.Printf("%s 要求(s) %s 得到正式認可的\n", icons.OK, numbers.List(ids, "#"))
}
