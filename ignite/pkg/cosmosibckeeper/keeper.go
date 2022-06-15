package cosmosibckeeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
)

// Keeper定義 IBC Keeper
type Keeper struct {
	portKey       []byte
	storeKey      sdk.StoreKey
	ChannelKeeper ChannelKeeper
	PortKeeper    PortKeeper
	ScopedKeeper  ScopedKeeper
}

// NewKeeper創建 IBC Keeper
func NewKeeper(
	portKey []byte,
	storeKey sdk.StoreKey,
	channelKeeper ChannelKeeper,
	portKeeper PortKeeper,
	scopedKeeper ScopedKeeper,
) *Keeper {
	return &Keeper{
		portKey:       portKey,
		storeKey:      storeKey,
		ChannelKeeper: channelKeeper,
		PortKeeper:    portKeeper,
		ScopedKeeper:  scopedKeeper,
	}
}

// ChanCloseInit 為通道 Keeper 的函數定義了一個包裝函數
func (k Keeper) ChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	capName := host.ChannelCapabilityPath(portID, channelID)
	chanCap, ok := k.ScopedKeeper.GetCapability(ctx, capName)
	if !ok {
		return sdkerrors.Wrapf(channeltypes.ErrChannelCapabilityNotFound, "不能檢索頻道: %s", capName)
	}
	return k.ChannelKeeper.ChanCloseInit(ctx, portID, channelID, chanCap)
}

// IsBound 檢查模塊是否已經綁定到所需的端口
func (k Keeper) IsBound(ctx sdk.Context, portID string) bool {
	_, ok := k.ScopedKeeper.GetCapability(ctx, host.PortPath(portID))
	return ok
}

// BindPort 為 ort Keeper 的函數定義了一個包裝函數
// 為了將它暴露給模塊的 InitGenesis 函數
func (k Keeper) BindPort(ctx sdk.Context, portID string) error {
	cap := k.PortKeeper.BindPort(ctx, portID)
	return k.ClaimCapability(ctx, cap, host.PortPath(portID))
}

// GetPort 返回模塊的端口 ID。用於 ExportGenesis
func (k Keeper) GetPort(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(k.portKey))
}

// SetPort 設置模塊的 portID。用於 InitGenesis
func (k Keeper) SetPort(ctx sdk.Context, portID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(k.portKey, []byte(portID))
}

// AuthenticateCapability 包裝 scopedKeeper 的 AuthenticateCapability 函數
func (k Keeper) AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool {
	return k.ScopedKeeper.AuthenticateCapability(ctx, cap, name)
}

// ClaimCapability 允許可以聲明 IBC 模塊傳遞給它的功能的模塊
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.ScopedKeeper.ClaimCapability(ctx, cap, name)
}
