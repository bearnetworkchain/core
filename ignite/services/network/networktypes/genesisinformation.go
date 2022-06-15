package networktypes

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	launchtypes "github.com/tendermint/spn/x/launch/types"
)

// GenesisInformation 代錶鍊構建創世的所有信息。
// 這個結構通過地址來索引賬戶和驗證者以獲得更好的性能
type GenesisInformation struct {
// 確保對以下內容使用切片，因為切片是有序的。
// 他們後來用來創建一個創世紀，所以，讓他們有序是很重要的
// 能夠產生確定性的創世紀。

	GenesisAccounts   []GenesisAccount
	VestingAccounts   []VestingAccount
	GenesisValidators []GenesisValidator
}

//GenesisAccount 代表一個用於鏈創世的具有初始硬幣分配的帳戶
type GenesisAccount struct {
	Address string
	Coins   string
}

// VestingAccount 表示具有初始硬幣分配和鏈創世歸屬選項的歸屬賬戶
// VestingAccount 目前僅支持延遲歸屬選項
type VestingAccount struct {
	Address      string
	TotalBalance string
	Vesting      string
	EndTime      int64
}

// GenesisValidator 代表了一個創世驗證器，它與創世鏈中的一個 gentx 相關聯
type GenesisValidator struct {
	Address        string
	Gentx          []byte
	Peer           launchtypes.Peer
	SelfDelegation sdk.Coin
}

// ToGenesisAccount 從 SPN 轉換創世賬戶
func ToGenesisAccount(acc launchtypes.GenesisAccount) GenesisAccount {
	return GenesisAccount{
		Address: acc.Address,
		Coins:   acc.Coins.String(),
	}
}

// ToVestingAccount 從 SPN 轉換歸屬賬戶
func ToVestingAccount(acc launchtypes.VestingAccount) (VestingAccount, error) {
	delayedVesting := acc.VestingOptions.GetDelayedVesting()
	if delayedVesting == nil {
		return VestingAccount{}, errors.New("僅支持延遲歸屬選項")
	}

	return VestingAccount{
		Address:      acc.Address,
		TotalBalance: delayedVesting.TotalBalance.String(),
		Vesting:      delayedVesting.Vesting.String(),
		EndTime:      delayedVesting.EndTime,
	}, nil
}

// ToGenesisValidator 從 SPN 轉換創世驗證器
func ToGenesisValidator(val launchtypes.GenesisValidator) GenesisValidator {
	return GenesisValidator{
		Address:        val.Address,
		Gentx:          val.GenTx,
		Peer:           val.Peer,
		SelfDelegation: val.SelfDelegation,
	}
}

// NewGenesisInformation 初始化一個新的GenesisInformation
func NewGenesisInformation(
	genAccs []GenesisAccount,
	vestingAccs []VestingAccount,
	genVals []GenesisValidator,
) (gi GenesisInformation) {
	return GenesisInformation{
		GenesisAccounts:   genAccs,
		VestingAccounts:   vestingAccs,
		GenesisValidators: genVals,
	}
}

func (gi GenesisInformation) ContainsGenesisAccount(address string) bool {
	for _, account := range gi.GenesisAccounts {
		if account.Address == address {
			return true
		}
	}
	return false
}
func (gi GenesisInformation) ContainsVestingAccount(address string) bool {
	for _, account := range gi.VestingAccounts {
		if account.Address == address {
			return true
		}
	}
	return false
}
func (gi GenesisInformation) ContainsGenesisValidator(address string) bool {
	for _, account := range gi.GenesisValidators {
		if account.Address == address {
			return true
		}
	}
	return false
}

func (gi *GenesisInformation) AddGenesisAccount(acc GenesisAccount) {
	gi.GenesisAccounts = append(gi.GenesisAccounts, acc)
}

func (gi *GenesisInformation) AddVestingAccount(acc VestingAccount) {
	gi.VestingAccounts = append(gi.VestingAccounts, acc)
}

func (gi *GenesisInformation) AddGenesisValidator(val GenesisValidator) {
	gi.GenesisValidators = append(gi.GenesisValidators, val)
}

func (gi *GenesisInformation) RemoveGenesisAccount(address string) {
	for i, account := range gi.GenesisAccounts {
		if account.Address == address {
			gi.GenesisAccounts = append(gi.GenesisAccounts[:i], gi.GenesisAccounts[i+1:]...)
		}
	}
}

func (gi *GenesisInformation) RemoveVestingAccount(address string) {
	for i, account := range gi.VestingAccounts {
		if account.Address == address {
			gi.VestingAccounts = append(gi.VestingAccounts[:i], gi.VestingAccounts[i+1:]...)
		}
	}
}

func (gi *GenesisInformation) RemoveGenesisValidator(address string) {
	for i, account := range gi.GenesisValidators {
		if account.Address == address {
			gi.GenesisValidators = append(gi.GenesisValidators[:i], gi.GenesisValidators[i+1:]...)
		}
	}
}

// ApplyRequest 將批准請求所隱含的更改應用於 genesisInformation
func (gi GenesisInformation) ApplyRequest(request Request) (GenesisInformation, error) {
	switch requestContent := request.Content.Content.(type) {
	case *launchtypes.RequestContent_GenesisAccount:
		// 創世紀中的新創世賬戶
		ga := ToGenesisAccount(*requestContent.GenesisAccount)
		genExist := gi.ContainsGenesisAccount(ga.Address)
		vestingExist := gi.ContainsVestingAccount(ga.Address)
		if genExist || vestingExist {
			return gi, NewWrappedErrInvalidRequest(request.RequestID, "創世賬戶已經在創世文件中")
		}
		gi.AddGenesisAccount(ga)

	case *launchtypes.RequestContent_VestingAccount:
		// 創世紀中的新歸屬賬戶
		va, err := ToVestingAccount(*requestContent.VestingAccount)
		if err != nil {
// 我們不將此錯誤視為 errInvalidRequests
// 因為如果我們不支持這種歸屬賬戶格式，就會發生這種情況
// 但請求仍然正確
			return gi, err
		}

		genExist := gi.ContainsGenesisAccount(va.Address)
		vestingExist := gi.ContainsVestingAccount(va.Address)
		if genExist || vestingExist {
			return gi, NewWrappedErrInvalidRequest(request.RequestID, "歸屬賬戶已經在創世中")
		}
		gi.AddVestingAccount(va)

	case *launchtypes.RequestContent_AccountRemoval:
	// 賬戶從創世中移除
		ar := requestContent.AccountRemoval
		genExist := gi.ContainsGenesisAccount(ar.Address)
		vestingExist := gi.ContainsVestingAccount(ar.Address)
		if !genExist && !vestingExist {
			return gi, NewWrappedErrInvalidRequest(request.RequestID, "無法刪除帳戶，因為它不存在")
		}
		gi.RemoveGenesisAccount(ar.Address)
		gi.RemoveVestingAccount(ar.Address)

	case *launchtypes.RequestContent_GenesisValidator:
		// new genesis validator in the genesis
		gv := ToGenesisValidator(*requestContent.GenesisValidator)
		if gi.ContainsGenesisValidator(gv.Address) {
			return gi, NewWrappedErrInvalidRequest(request.RequestID, "創世驗證器已經在創世文件中")
		}
		gi.AddGenesisValidator(gv)

	case *launchtypes.RequestContent_ValidatorRemoval:
		//驗證者從創世中移除
		vr := requestContent.ValidatorRemoval
		if !gi.ContainsGenesisValidator(vr.ValAddress) {
			return gi, NewWrappedErrInvalidRequest(request.RequestID, "創世紀驗證器無法刪除，因為它不存在")
		}
	}

	return gi, nil
}
