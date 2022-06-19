package ignitecmd

import "github.com/spf13/cobra"

// NewNetworkRequest 創建一個新的批准請求命令，其中包含其他一些
// 與處理鏈請求相關的子命令。
func NewNetworkRequest() *cobra.Command {
	c := &cobra.Command{
		Use:   "request",
		Short: "處理請求",
	}

	c.AddCommand(
		NewNetworkRequestShow(),
		NewNetworkRequestList(),
		NewNetworkRequestApprove(),
		NewNetworkRequestReject(),
		NewNetworkRequestVerify(),
	)

	return c
}
