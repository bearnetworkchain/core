package ignitecmd

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/yaml"
	"github.com/ignite-hq/cli/ignite/services/network"
)

// NewNetworkRequestShow 創建一個新的請求顯示命令來顯示
// 請求鏈的詳細信息
func NewNetworkRequestShow() *cobra.Command {
	c := &cobra.Command{
		Use:   "show [launch-id] [request-id]",
		Short: "顯示待處理的請求詳細信息",
		RunE:  networkRequestShowHandler,
		Args:  cobra.ExactArgs(2),
	}
	return c
}

func networkRequestShowHandler(cmd *cobra.Command, args []string) error {
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

	// 解析請求ID
	requestID, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return errors.Wrap(err, "error parsing requestID")
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	request, err := n.Request(cmd.Context(), launchID, requestID)
	if err != nil {
		return err
	}

	// 將請求對象轉換為 YAML 以提高可讀性
	// 並將字節數組字段轉換為字符串。
	requestYaml, err := yaml.Marshal(cmd.Context(), request,
		"$.Content.content.genesisValidator.genTx",
		"$.Content.content.genesisValidator.consPubKey",
	)
	if err != nil {
		return err
	}

	session.StopSpinner()

	return session.Println(requestYaml)
}
