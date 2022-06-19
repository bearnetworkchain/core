package ignitecmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cache"
	"github.com/ignite-hq/cli/ignite/pkg/chaincmd"
	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/icons"
	"github.com/ignite-hq/cli/ignite/pkg/numbers"
	"github.com/ignite-hq/cli/ignite/services/network"
	"github.com/ignite-hq/cli/ignite/services/network/networkchain"
)

// NewNetworkRequestVerify 驗證請求並模擬鏈。
func NewNetworkRequestVerify() *cobra.Command {
	c := &cobra.Command{
		Use:   "verify [launch-id] [number<,...>]",
		Short: "驗證請求並從它們模擬鏈創世",
		RunE:  networkRequestVerifyHandler,
		Args:  cobra.ExactArgs(2),
	}

	flagSetClearCache(c)
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetHome())
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	return c
}

func networkRequestVerifyHandler(cmd *cobra.Command, args []string) error {
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

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	// 驗證請求
	if err := verifyRequest(cmd.Context(), cacheStorage, nb, launchID, ids...); err != nil {
		session.Printf("%s 要求 %s 無效\n", icons.NotOK, numbers.List(ids, "#"))
		return err
	}

	return session.Printf("%s 要求 %s 已驗證\n", icons.OK, numbers.List(ids, "#"))
}

// verifyRequest 從臨時目錄中的啟動 ID 初始化鏈
// 並使用請求 ID 模擬從 genesis 啟動鏈
func verifyRequest(
	ctx context.Context,
	cacheStorage cache.Storage,
	nb NetworkBuilder,
	launchID uint64,
	requestIDs ...uint64,
) error {
	n, err := nb.Network()
	if err != nil {
		return err
	}

	// 使用臨時目錄初始化鏈
	chainLaunch, err := n.ChainLaunch(ctx, launchID)
	if err != nil {
		return err
	}

	homeDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(homeDir)

	c, err := nb.Chain(
		networkchain.SourceLaunch(chainLaunch),
		networkchain.WithHome(homeDir),
		networkchain.WithKeyringBackend(chaincmd.KeyringBackendTest),
	)
	if err != nil {
		return err
	}

	// 獲取當前的創世信息和對鏈的請求以進行模擬
	genesisInformation, err := n.GenesisInformation(ctx, launchID)
	if err != nil {
		return err
	}

	requests, err := n.RequestFromIDs(ctx, launchID, requestIDs...)
	if err != nil {
		return err
	}

	return c.SimulateRequests(
		ctx,
		cacheStorage,
		genesisInformation,
		requests,
	)
}
