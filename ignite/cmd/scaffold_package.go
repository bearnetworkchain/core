package ignitecmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/cliui/clispinner"
	"github.com/bearnetworkchain/core/ignite/pkg/placeholder"
	"github.com/bearnetworkchain/core/ignite/services/scaffolder"
)

const (
	flagAck = "ack"
)

// NewScaffoldPacket 在模塊中創建一個新數據包
func NewScaffoldPacket() *cobra.Command {
	c := &cobra.Command{
		Use:   "packet [packetName] [field1] [field2] ... --module [moduleName]",
		Short: "發送 IBC 數據包的消息",
		Long:  "在特定的啟用 IBC 的 Cosmos SDK 模塊中搭建 IBC 數據包",
		Args:  cobra.MinimumNArgs(1),
		RunE:  createPacketHandler,
	}

	flagSetPath(c)
	flagSetClearCache(c)
	c.Flags().StringSlice(flagAck, []string{}, "自定義確認類型 (field1(場地1),field2(場地2),...)")
	c.Flags().String(flagModule, "", "IBC 模塊將數據包添加到")
	c.Flags().String(flagSigner, "", "消息簽名者的標籤（默認：創建者）")
	c.Flags().Bool(flagNoMessage, false, "禁用發送消息腳手架")

	return c
}

func createPacketHandler(cmd *cobra.Command, args []string) error {
	s := clispinner.New().SetText("創建中,請耐心等候...")
	defer s.Stop()

	var (
		packet       = args[0]
		packetFields = args[1:]
		signer       = flagGetSigner(cmd)
		appPath      = flagGetPath(cmd)
	)

	module, err := cmd.Flags().GetString(flagModule)
	if err != nil {
		return err
	}
	if module == "" {
		return errors.New("請指定一個模塊來創建數據包: --module <模塊名稱>")
	}

	ackFields, err := cmd.Flags().GetStringSlice(flagAck)
	if err != nil {
		return err
	}

	noMessage, err := cmd.Flags().GetBool(flagNoMessage)
	if err != nil {
		return err
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	var options []scaffolder.PacketOption
	if noMessage {
		options = append(options, scaffolder.PacketWithoutMessage())
	} else if signer != "" {
		options = append(options, scaffolder.PacketWithSigner(signer))
	}

	sc, err := newApp(appPath)
	if err != nil {
		return err
	}

	sm, err := sc.AddPacket(cmd.Context(), cacheStorage, placeholder.New(), module, packet, packetFields, ackFields, options...)
	if err != nil {
		return err
	}

	s.Stop()

	modificationsStr, err := sourceModificationToString(sm)
	if err != nil {
		return err
	}

	fmt.Println(modificationsStr)
	fmt.Printf("\n🎉 創建了一個數據包 `%[1]v`.\n\n", args[0])

	return nil
}
