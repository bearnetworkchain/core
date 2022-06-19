package ignitecmd

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite-hq/cli/ignite/pkg/relayer"
)

// NewRelayerConnect 返回一個新的中繼器連接命令以鏈接所有或部分中繼器路徑並啟動
// 在兩者之間中繼 txs。
// 如果未指定路徑，則鏈接所有路徑。
func NewRelayerConnect() *cobra.Command {
	c := &cobra.Command{
		Use:   "connect [<path>,...]",
		Short: "與路徑關聯的鏈接鏈並開始在其間中繼 tx 數據包",
		RunE:  relayerConnectHandler,
	}

	c.Flags().AddFlagSet(flagSetKeyringBackend())

	return c
}

func relayerConnectHandler(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		err = handleRelayerAccountErr(err)
	}()

	session := cliui.New()
	defer session.Cleanup()

	ca, err := cosmosaccount.New(
		cosmosaccount.WithKeyringBackend(getKeyringBackend(cmd)),
	)
	if err != nil {
		return err
	}

	if err := ca.EnsureDefaultAccount(); err != nil {
		return err
	}

	var (
		use []string
		ids = args
		r   = relayer.New(ca)
	)

	all, err := r.ListPaths(cmd.Context())
	if err != nil {
		return err
	}

	// 如果沒有提供路徑 id，那麼我們連接所有路徑，否則，
	// 只連接指定的。
	if len(ids) == 0 {
		for _, path := range all {
			use = append(use, path.ID)

		}
	} else {
		for _, id := range ids {
			for _, path := range all {
				if id == path.ID {
					use = append(use, path.ID)
					break
				}

			}
		}
	}

	if len(use) == 0 {
		session.StopSpinner()
		session.Println("未找到可連接的鏈.")
		return nil
	}

	session.StartSpinner("在鏈之間創建鏈接...")

	if err := r.Link(cmd.Context(), use...); err != nil {
		return err
	}

	session.StopSpinner()

	if err := printSection(session, "Paths"); err != nil {
		return err
	}

	for _, id := range use {
		session.StartSpinner("正在加載...")

		path, err := r.GetPath(cmd.Context(), id)
		if err != nil {
			return err
		}

		session.StopSpinner()

		var buf bytes.Buffer
		w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', tabwriter.TabIndent)
		fmt.Fprintf(w, "%s:\n", path.ID)
		fmt.Fprintf(w, "   \t%s\t>\t(port: %s)\t(channel: %s)\n", path.Src.ChainID, path.Src.PortID, path.Src.ChannelID)
		fmt.Fprintf(w, "   \t%s\t>\t(port: %s)\t(channel: %s)\n", path.Dst.ChainID, path.Dst.PortID, path.Dst.ChannelID)
		fmt.Fprintln(w)
		w.Flush()
		session.Print(buf.String())
	}

	if err := printSection(session, "在鏈之間監聽及中繼數據包..."); err != nil {
		return err
	}

	return r.Start(cmd.Context(), use...)
}
