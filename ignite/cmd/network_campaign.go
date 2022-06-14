package ignitecmd

import (
	"github.com/spf13/cobra"
)

// NewNetworkCampaign 創建一個新的活動命令，該命令包含其他
// 與為活動啟動網絡相關的子命令。
func NewNetworkCampaign() *cobra.Command {
	c := &cobra.Command{
		Use:   "campaign",
		Short: "處理活動",
	}
	c.AddCommand(
		NewNetworkCampaignPublish(),
		NewNetworkCampaignList(),
		NewNetworkCampaignShow(),
		NewNetworkCampaignUpdate(),
		NewNetworkCampaignAccount(),
	)
	return c
}
