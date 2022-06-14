package ignitecmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/icons"
	"github.com/ignite-hq/cli/ignite/services/network"
)

func newNetworkChainShowPeers() *cobra.Command {
	c := &cobra.Command{
		Use:   "peers [launch-id]",
		Short: "顯示鏈的對等列表",
		Args:  cobra.ExactArgs(1),
		RunE:  networkChainShowPeersHandler,
	}

	c.Flags().String(flagOut, "./peers.txt", "Path to output peers list")

	return c
}

func networkChainShowPeersHandler(cmd *cobra.Command, args []string) error {
	session := cliui.New()
	defer session.Cleanup()

	out, _ := cmd.Flags().GetString(flagOut)

	nb, launchID, err := networkChainLaunch(cmd, args, session)
	if err != nil {
		return err
	}
	n, err := nb.Network()
	if err != nil {
		return err
	}

	genVals, err := n.GenesisValidators(cmd.Context(), launchID)
	if err != nil {
		return err
	}

	peers := make([]string, 0)
	for _, acc := range genVals {
		peer, err := network.PeerAddress(acc.Peer)
		if err != nil {
			return err
		}
		peers = append(peers, peer)
	}

	if len(peers) == 0 {
		session.Printf("%s %s\n", icons.Info, "找不到對等")
		return nil

	}

	if err := os.MkdirAll(filepath.Dir(out), 0744); err != nil {
		return err
	}

	b := &bytes.Buffer{}
	peerList := strings.Join(peers, ",")
	fmt.Fprintln(b, peerList)
	if err := os.WriteFile(out, b.Bytes(), 0644); err != nil {
		return err
	}

	session.StopSpinner()

	return session.Printf("%s 已生成對等列表: %s\n", icons.Bullet, out)
}
