package networktypes

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	campaigntypes "github.com/tendermint/spn/x/campaign/types"
)

// Campaign 代表 SPN 上某鏈的活動
type Campaign struct {
	ID                 uint64    `json:"ID"`
	Name               string    `json:"Name"`
	CoordinatorID      uint64    `json:"CoordinatorID"`
	MainnetID          uint64    `json:"MainnetID"`
	MainnetInitialized bool      `json:"MainnetInitialized"`
	TotalSupply        sdk.Coins `json:"TotalSupply"`
	AllocatedShares    string    `json:"AllocatedShares"`
	Metadata           string    `json:"Metadata"`
}

// ToCampaign 從 SPN 轉換活動數據並返回一個活動對象
func ToCampaign(campaign campaigntypes.Campaign) Campaign {
	return Campaign{
		ID:                 campaign.CampaignID,
		Name:               campaign.CampaignName,
		CoordinatorID:      campaign.CoordinatorID,
		MainnetID:          campaign.MainnetID,
		MainnetInitialized: campaign.MainnetInitialized,
		TotalSupply:        campaign.TotalSupply,
		AllocatedShares:    campaign.AllocatedShares.String(),
		Metadata:           string(campaign.Metadata),
	}
}

// MainnetAccount 代表 SPN 上某鏈的競選主網賬戶
type MainnetAccount struct {
	Address string               `json:"Address"`
	Shares  campaigntypes.Shares `json:"Shares"`
}

// ToMainnetAccount 從 SPN 轉換主網賬戶數據並返回 MainnetAccount 對象
func ToMainnetAccount(acc campaigntypes.MainnetAccount) MainnetAccount {
	return MainnetAccount{
		Address: acc.Address,
		Shares:  acc.Shares,
	}
}

// MainnetVestingAccount 代表 SPN 上鍊的活動主網歸屬賬戶
type MainnetVestingAccount struct {
	Address     string               `json:"Address"`
	TotalShares campaigntypes.Shares `json:"TotalShares"`
	Vesting     campaigntypes.Shares `json:"Vesting"`
	EndTime     int64                `json:"EndTime"`
}

// ToMainnetVestingAccount 從 SPN 轉換主網歸屬賬戶數據並返回 MainnetVestingAccount 對象
func ToMainnetVestingAccount(acc campaigntypes.MainnetVestingAccount) MainnetVestingAccount {
	delaydVesting := acc.VestingOptions.GetDelayedVesting()
	return MainnetVestingAccount{
		Address:     acc.Address,
		TotalShares: delaydVesting.TotalShares,
		Vesting:     delaydVesting.Vesting,
		EndTime:     delaydVesting.EndTime,
	}
}
