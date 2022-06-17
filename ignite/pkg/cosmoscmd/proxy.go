package cosmoscmd

import (
	"path/filepath"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/cosmosutil"
	"github.com/bearnetworkchain/core/ignite/pkg/ctxticker"
	"github.com/bearnetworkchain/core/ignite/pkg/gitpod"
	"github.com/bearnetworkchain/core/ignite/pkg/xchisel"
	"github.com/bearnetworkchain/core/ignite/services/network/networkchain"
)

const TunnelRerunDelay = 5 * time.Second

// startProxyForTunneledPeers 掛鉤 `appd start` 命令以啟動 HTTP 代理服務器和 HTTP 代理客戶端
// 對於每個需要 HTTP 隧道的節點。
// 如果您的應用程序的 `$APP_HOME/config` 目錄有 `spn.yml` 文件，則 HTTP 隧道會被激活**僅**
// 並且只有當這個文件裡面有 `tunneled_peers` 字段和隧道對等/節點列表時。
//
// 如果您使用 SPN 作為協調器並且根本不想允許 HTTP 隧道功能，
// 您可以通過不批准驗證器請求來防止生成 `spn.yml` 文件
// 啟用了 HTTP 隧道而不是普通的 TCP 連接。
func startProxyForTunneledPeers(clientCtx client.Context, cmd *cobra.Command) {
	if cmd.Name() != "start" {
		return
	}
	serverCtx := server.GetServerContextFromCmd(cmd)
	ctx := cmd.Context()

	spnConfigPath := filepath.Join(clientCtx.HomeDir, cosmosutil.ChainConfigDir, networkchain.SPNConfigFile)
	spnConfig, err := networkchain.GetSPNConfig(spnConfigPath)
	if err != nil {
		serverCtx.Logger.Error("無法打開 spn 配置文件", "原因", err.Error())
		return
	}
	// 如果網絡中沒有隧道驗證器，則退出
	if len(spnConfig.TunneledPeers) == 0 {
		return
	}

	for _, peer := range spnConfig.TunneledPeers {
		if peer.Name == networkchain.HTTPTunnelChisel {
			peer := peer
			go func() {
				ctxticker.DoNow(ctx, TunnelRerunDelay, func() error {
					serverCtx.Logger.Info("啟動隧道客戶端", "隧道地址", peer.Address, "本地端口", peer.LocalPort)
					err := xchisel.StartClient(ctx, peer.Address, peer.LocalPort, "26656")
					if err != nil {
						serverCtx.Logger.Error("啟動隧道客戶端失敗",
							"tunnelAddress", peer.Address,
							"localPort", peer.LocalPort,
							"reason", err.Error(),
						)
					}
					return nil
				})
			}()
		}
	}

	if gitpod.IsOnGitpod() {
		go func() {
			ctxticker.DoNow(ctx, TunnelRerunDelay, func() error {
				serverCtx.Logger.Info("啟動隧道服務器", "port", xchisel.DefaultServerPort)
				err := xchisel.StartServer(ctx, xchisel.DefaultServerPort)
				if err != nil {
					serverCtx.Logger.Error(
						"無法啟動隧道服務器",
						"port", xchisel.DefaultServerPort,
						"reason", err.Error(),
					)
				}
				return nil
			})
		}()
	}
}
