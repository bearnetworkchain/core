package ignitecmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/cliui"
	"github.com/bearnetworkchain/core/ignite/pkg/cliui/colors"
	"github.com/bearnetworkchain/core/ignite/pkg/cliui/icons"
	"github.com/bearnetworkchain/core/ignite/pkg/goenv"
	"github.com/bearnetworkchain/core/ignite/services/network"
	"github.com/bearnetworkchain/core/ignite/services/network/networkchain"
)

const (
	flagForce = "force"
)

// NewNetworkChainPrepare 返回一個新命令以準備啟動熊網鏈
func NewNetworkChainPrepare() *cobra.Command {
	c := &cobra.Command{
		Use:   "prepare [launch-id]",
		Short: "準備啟動鏈",
		Args:  cobra.ExactArgs(1),
		RunE:  networkChainPrepareHandler,
	}

	flagSetClearCache(c)
	c.Flags().BoolP(flagForce, "f", false, "即使鏈沒有啟動，也強制命令運行")
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	c.Flags().AddFlagSet(flagSetHome())

	return c
}

func networkChainPrepareHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	force, _ := cmd.Flags().GetBool(flagForce)

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	nb, err := newNetworkBuilder(cmd, CollectEvents(session.EventBus()))
	if err != nil {
		return err
	}

	// 解析啟動 ID
	launchID, err := network.ParseID(args[0])
	if err != nil {
		return err
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	// 獲取鏈信息
	chainLaunch, err := n.ChainLaunch(cmd.Context(), launchID)
	if err != nil {
		return err
	}

	if !force && !chainLaunch.LaunchTriggered {
		return fmt.Errorf("熊網鏈 %d 啟動尚未觸發. 指令加入 --force 無論如何都要啟動", launchID)
	}

	c, err := nb.Chain(networkchain.SourceLaunch(chainLaunch))
	if err != nil {
		return err
	}

	// 獲取信息以構建創世紀
	genesisInformation, err := n.GenesisInformation(cmd.Context(), launchID)
	if err != nil {
		return err
	}

	rewardsInfo, lastBlockHeight, unboundingTime, err := n.RewardsInfo(
		cmd.Context(),
		launchID,
		chainLaunch.ConsumerRevisionHeight,
	)
	if err != nil {
		return err
	}

	spnChainID, err := n.ChainID(cmd.Context())
	if err != nil {
		return err
	}

	if err := c.Prepare(
		cmd.Context(),
		cacheStorage,
		genesisInformation,
		rewardsInfo,
		spnChainID,
		lastBlockHeight,
		unboundingTime,
	); err != nil {
		return err
	}

	chainHome, err := c.Home()
	if err != nil {
		return err
	}
	binaryName, err := c.BinaryName()
	if err != nil {
		return err
	}
	binaryDir := filepath.Dir(filepath.Join(goenv.Bin(), binaryName))

	session.StopSpinner()
	session.Printf("%s 鏈準備啟動\n", icons.OK)
	session.Println("\n您可以通過運行以下命令來啟動節點:")
	commandStr := fmt.Sprintf("%s start --home %s", binaryName, chainHome)
	session.Printf("\t%s/%s\n", binaryDir, colors.Info(commandStr))

	return nil
}
