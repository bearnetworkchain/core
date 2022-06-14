package ignitecmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/icons"
	"github.com/ignite-hq/cli/ignite/services/network/networkchain"
)

func newNetworkChainShowGenesis() *cobra.Command {
	c := &cobra.Command{
		Use:   "genesis [launch-id]",
		Short: "顯示熊網鏈創世文件",
		Args:  cobra.ExactArgs(1),
		RunE:  networkChainShowGenesisHandler,
	}

	flagSetClearCache(c)
	c.Flags().String(flagOut, "./genesis.json", "Path to output Genesis file")

	return c
}

func networkChainShowGenesisHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	out, _ := cmd.Flags().GetString(flagOut)

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	nb, launchID, err := networkChainLaunch(cmd, args, session)
	if err != nil {
		return err
	}
	n, err := nb.Network()
	if err != nil {
		return err
	}

	chainLaunch, err := n.ChainLaunch(cmd.Context(), launchID)
	if err != nil {
		return err
	}

	c, err := nb.Chain(networkchain.SourceLaunch(chainLaunch))
	if err != nil {
		return err
	}

	genesisPath, err := c.GenesisPath()
	if err != nil {
		return err
	}

	spnChainID, err := n.ChainID(cmd.Context())
	if err != nil {
		return err
	}

	// 檢查起源是否已經存在
	if _, err = os.Stat(genesisPath); os.IsNotExist(err) {
		// 獲取信息以構建創世紀
		genesisInformation, err := n.GenesisInformation(cmd.Context(), launchID)
		if err != nil {
			return err
		}

		// 在臨時目錄中創建鏈
		tmpHome, err := os.MkdirTemp("", "*-spn")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpHome)

		c.SetHome(tmpHome)

		rewardsInfo, lastBlockHeight, unboundingTime, err := n.RewardsInfo(
			cmd.Context(),
			launchID,
			chainLaunch.ConsumerRevisionHeight,
		)
		if err != nil {
			return err
		}

		if err = c.Prepare(
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

		// 獲得新的創世路徑
		genesisPath, err = c.GenesisPath()
		if err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Dir(out), 0744); err != nil {
		return err
	}

	if err := os.Rename(genesisPath, out); err != nil {
		return err
	}

	session.StopSpinner()

	return session.Printf("%s 創世紀生成: %s\n", icons.Bullet, out)
}
