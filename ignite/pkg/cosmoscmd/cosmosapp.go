package cosmoscmd

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CosmosApp 實現了基於 Cosmos SDK 的應用程序的常用方法
// 特定的區塊鏈。
type CosmosApp interface {
	// 應用程序的指定名稱.
	Name() string

// 應用程序類型編解碼器。
// 注意：這應該在退回之前密封。
	LegacyAmino() *codec.LegacyAmino

	// 應用程序更新每個開始塊。
	BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock

	// 應用程序更新每個結束塊。
	EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock

	// 鏈（即應用程序）初始化時的應用程序更新.
	InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain

	// 在給定高度加載應用程序.
	LoadHeight(height int64) error

	// 導出創世文件的應用程序狀態。
	ExportAppStateAndValidators(
		forZeroHeight bool, jailAllowedAddrs []string,
	) (types.ExportedApp, error)

	// 所有註冊的模塊賬號地址.
	ModuleAccountAddrs() map[string]bool
}
