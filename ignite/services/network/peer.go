package network

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	launchtypes "github.com/tendermint/spn/x/launch/types"
)

func PeerAddress(peer launchtypes.Peer) (string, error) {
	var peerAddr string
	switch conn := peer.Connection.(type) {
	case *launchtypes.Peer_TcpAddress:
		peerAddr = fmt.Sprintf("%s@%s", peer.Id, conn.TcpAddress)
	case *launchtypes.Peer_HttpTunnel:
		peerAddr = fmt.Sprintf("%s@%s", peer.Id, conn.HttpTunnel.Address)
	default:
		return peerAddr, fmt.Errorf("對等連接類型無效: %T", peer.Connection)
	}
	return peerAddr, nil
}

func ParsePeerAddress(addr string) (launchtypes.Peer, error) {
	sp := strings.Split(addr, "@")
	if len(sp) != 2 {
		return launchtypes.Peer{}, errors.New("無效的對等地址格式")
	}
	return launchtypes.NewPeerConn(sp[0], sp[1]), nil
}
