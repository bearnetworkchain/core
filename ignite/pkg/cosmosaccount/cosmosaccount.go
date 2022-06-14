package cosmosaccount

import (
	"errors"
	"fmt"
	"os"

	dkeyring "github.com/99designs/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/go-bip39"
)

const (
	// KeyringServiceName 用於 OS 後端的密鑰環名稱。
	KeyringServiceName = "bearnetwork"

	// DefaultAccount 是默認帳戶的名稱。
	DefaultAccount = "default"
)

// KeyringHome 用於存儲賬戶相關數據。
var KeyringHome = os.ExpandEnv("$HOME/.ignite/accounts")

var (
	ErrAccountExists = errors.New("賬戶已存在")
)

const (
	AccountPrefixCosmos = "bnkt"
)

// KeyringBackend 是存儲密鑰的後端.
type KeyringBackend string

const (
// KeyringTest 是測試密鑰環後端。使用此後端，您的密鑰將是
// 存儲在您應用的數據目錄下，
	KeyringTest KeyringBackend = "test"

// KeyringOS 是操作系統密鑰環後端。使用此後端，您的密鑰將是
// 存儲在操作系統的安全密鑰環中。
	KeyringOS KeyringBackend = "os"

// KeyringMemory 在內存中密鑰環後端，您的密鑰將存儲在應用程序內存中。
	KeyringMemory KeyringBackend = "memory"
)

// 帳戶註冊。
type Registry struct {
	homePath           string
	keyringServiceName string
	keyringBackend     KeyringBackend

	Keyring keyring.Keyring
}

//選項配置您的註冊表。
type Option func(*Registry)

func WithHome(path string) Option {
	return func(c *Registry) {
		c.homePath = path
	}
}

func WithKeyringServiceName(name string) Option {
	return func(c *Registry) {
		c.keyringServiceName = name
	}
}

func WithKeyringBackend(backend KeyringBackend) Option {
	return func(c *Registry) {
		c.keyringBackend = backend
	}
}

// New 創建一個新的註冊表來管理帳戶。
func New(options ...Option) (Registry, error) {
	r := Registry{
		keyringServiceName: sdktypes.KeyringServiceName(),
		keyringBackend:     KeyringTest,
		homePath:           KeyringHome,
	}

	for _, apply := range options {
		apply(&r)
	}

	var err error

	r.Keyring, err = keyring.New(r.keyringServiceName, string(r.keyringBackend), r.homePath, os.Stdin)
	if err != nil {
		return Registry{}, err
	}

	return r, nil
}

func NewStandalone(options ...Option) (Registry, error) {
	return New(
		append([]Option{
			WithKeyringServiceName(KeyringServiceName),
			WithHome(KeyringHome),
		}, options...)...,
	)
}

func NewInMemory(options ...Option) (Registry, error) {
	return New(
		append([]Option{
			WithKeyringBackend(KeyringMemory),
		}, options...)...,
	)
}

// Account 代表 Cosmos SDK 帳戶。
type Account struct {
	// 帳戶名稱。
	Name string

	// Info 包含有關該帳戶的附加信息。
	Info keyring.Info
}

// Address 從給定的前綴返回帳戶的地址。
func (a Account) Address(accPrefix string) string {
	if accPrefix == "" {
		accPrefix = AccountPrefixCosmos
	}

	return toBench32(accPrefix, a.Info.GetPubKey().Address())
}

// PubKey 返回帳戶的公鑰。
func (a Account) PubKey() string {
	return a.Info.GetPubKey().String()
}

func toBench32(prefix string, addr []byte) string {
	bech32Addr, err := bech32.ConvertAndEncode(prefix, addr)
	if err != nil {
		panic(err)
	}
	return bech32Addr
}

// EnsureDefaultAccount 確保默認帳戶存在。
func (r Registry) EnsureDefaultAccount() error {
	_, err := r.GetByName(DefaultAccount)

	var accErr *AccountDoesNotExistError
	if errors.As(err, &accErr) {
		_, _, err = r.Create(DefaultAccount)
		return err
	}

	return err
}

// Create 使用名稱創建一個新帳戶。
func (r Registry) Create(name string) (acc Account, mnemonic string, err error) {
	acc, err = r.GetByName(name)
	if err == nil {
		return Account{}, "", ErrAccountExists
	}
	var accErr *AccountDoesNotExistError
	if !errors.As(err, &accErr) {
		return Account{}, "", err
	}

	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		return Account{}, "", err
	}
	mnemonic, err = bip39.NewMnemonic(entropySeed)
	if err != nil {
		return Account{}, "", err
	}

	algo, err := r.algo()
	if err != nil {
		return Account{}, "", err
	}
	info, err := r.Keyring.NewAccount(name, mnemonic, "", r.hdPath(), algo)
	if err != nil {
		return Account{}, "", err
	}

	acc = Account{
		Name: name,
		Info: info,
	}

	return acc, mnemonic, nil
}

// Import 導入具有名稱、密碼和秘密的現有帳戶，其中秘密可以是助記符或私鑰。

func (r Registry) Import(name, secret, passphrase string) (Account, error) {
	_, err := r.GetByName(name)
	if err == nil {
		return Account{}, ErrAccountExists
	}
	var accErr *AccountDoesNotExistError
	if !errors.As(err, &accErr) {
		return Account{}, err
	}

	if bip39.IsMnemonicValid(secret) {
		algo, err := r.algo()
		if err != nil {
			return Account{}, err
		}
		_, err = r.Keyring.NewAccount(name, secret, passphrase, r.hdPath(), algo)
		if err != nil {
			return Account{}, err
		}
	} else if err := r.Keyring.ImportPrivKey(name, secret, passphrase); err != nil {
		return Account{}, err
	}

	return r.GetByName(name)
}

// Export 將帳戶導出為私鑰。

func (r Registry) Export(name, passphrase string) (key string, err error) {
	if _, err = r.GetByName(name); err != nil {
		return "", err
	}

	return r.Keyring.ExportPrivKeyArmor(name, passphrase)

}

// ExportHex 將帳戶導出為十六進制的私鑰。

func (r Registry) ExportHex(name, passphrase string) (hex string, err error) {
	if _, err = r.GetByName(name); err != nil {
		return "", err
	}

	return keyring.NewUnsafe(r.Keyring).UnsafeExportPrivKeyHex(name)
}

// GetByName 按其名稱返回一個帳戶。

func (r Registry) GetByName(name string) (Account, error) {
	info, err := r.Keyring.Key(name)
	if errors.Is(err, dkeyring.ErrKeyNotFound) || errors.Is(err, sdkerrors.ErrKeyNotFound) {
		return Account{}, &AccountDoesNotExistError{name}
	}
	if err != nil {
		return Account{}, nil
	}

	acc := Account{
		Name: name,
		Info: info,
	}

	return acc, nil
}

// List 列出所有帳戶。

func (r Registry) List() ([]Account, error) {
	info, err := r.Keyring.List()
	if err != nil {
		return nil, err
	}

	var accounts []Account

	for _, accinfo := range info {
		accounts = append(accounts, Account{
			Name: accinfo.GetName(),
			Info: accinfo,
		})
	}

	return accounts, nil
}

// DeleteByName 按名稱刪除帳戶。

func (r Registry) DeleteByName(name string) error {
	err := r.Keyring.Delete(name)
	if err == dkeyring.ErrKeyNotFound {
		return &AccountDoesNotExistError{name}
	}
	return err
}

func (r Registry) hdPath() string {
	return hd.CreateHDPath(sdktypes.GetConfig().GetCoinType(), 0, 0).String()
}

func (r Registry) algo() (keyring.SignatureAlgo, error) {
	algos, _ := r.Keyring.SupportedAlgorithms()
	return keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), algos)
}

type AccountDoesNotExistError struct {
	Name string
}

func (e *AccountDoesNotExistError) Error() string {
	return fmt.Sprintf("帳戶 %q 不存在", e.Name)
}
