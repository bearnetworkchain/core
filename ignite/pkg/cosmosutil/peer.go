package cosmosutil

import (
	"strings"

	launchtypes "github.com/tendermint/spn/x/launch/types"

	"github.com/bearnetworkchain/core/ignite/pkg/xurl"
)

// VerifyPeerFormat檢查對等地址格式是否有效
func VerifyPeerFormat(peer launchtypes.Peer) bool {
	// 檢查對等體的格式
	switch conn := peer.Connection.(type) {
	case *launchtypes.Peer_TcpAddress:
		nodeHost := strings.Split(conn.TcpAddress, ":")
		if len(nodeHost) != 2 ||
			len(nodeHost[0]) == 0 ||
			len(nodeHost[1]) == 0 {
			return false
		}
		return true
	case *launchtypes.Peer_HttpTunnel:
		return xurl.IsHTTP(conn.HttpTunnel.Address)
	default:
		return false
	}
}
