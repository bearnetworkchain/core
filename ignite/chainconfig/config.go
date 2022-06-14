package chainconfig

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/imdario/mergo"

	"github.com/ignite-hq/cli/ignite/pkg/xfilepath"
)

var (
	// ConfigDirPath 返回 Ignite 的配置目錄路徑。
	ConfigDirPath = xfilepath.JoinFromHome(xfilepath.Path(".ignite"))

	// ConfigFileNames 是 Ignite 配置文件的公認名稱列表。
	ConfigFileNames = []string{"config.yml", "config.yaml"}
)

var (
	// ErrCouldntLocateConfig 在源代碼中找不到 config.yml 時返回。
	ErrCouldntLocateConfig = errors.New(
		"找不到config.yml在你的鏈條中。請點擊鏈接" +
			"how-to: https://github.com/ignite-hq/cli/blob/develop/docs/configure/index.md")
)

// DefaultConf 保存默認配置。
var DefaultConf = Config{
	Host: Host{
		// 在 MacOS 上的 Docker 中時，它僅適用於 192.168.1.188.
		RPC:     "192.168.1.188:26657",
		P2P:     "192.168.1.188:26656",
		Prof:    "192.168.1.188:6060",
		GRPC:    "192.168.1.188:9090",
		GRPCWeb: "192.168.1.188:9091",
		API:     "192.168.1.188:1317",
	},
	Build: Build{
		Proto: Proto{
			Path: "proto",
			ThirdPartyPaths: []string{
				"third_party/proto",
				"proto_vendor",
			},
		},
	},
	Faucet: Faucet{
		Host: "192.168.1.188:4500",
	},
}

//Config 是用戶給定的配置以進行額外的設置
// 發球期間。
type Config struct {
	Accounts  []Account              `yaml:"accounts"`
	Validator Validator              `yaml:"validator"`
	Faucet    Faucet                 `yaml:"faucet"`
	Client    Client                 `yaml:"client"`
	Build     Build                  `yaml:"build"`
	Init      Init                   `yaml:"init"`
	Genesis   map[string]interface{} `yaml:"genesis"`
	Host      Host                   `yaml:"host"`
}

// AccountByName 按名稱查找帳戶。
func (c Config) AccountByName(name string) (acc Account, found bool) {
	for _, acc := range c.Accounts {
		if acc.Name == name {
			return acc, true
		}
	}
	return Account{}, false
}

// Account 擁有與設置 Cosmos 錢包相關的選項。
type Account struct {
	Name     string   `yaml:"name"`
	Coins    []string `yaml:"coins,omitempty"`
	Mnemonic string   `yaml:"mnemonic,omitempty"`
	Address  string   `yaml:"address,omitempty"`
	CoinType string   `yaml:"cointype,omitempty"`

	// 發行帳戶的鏈外的 RPCAddress。
	RPCAddress string `yaml:"rpc_address,omitempty"`
}

//驗證器保存與驗證器設置相關的信息。
type Validator struct {
	Name   string `yaml:"name"`
	Staked string `yaml:"staked"`
}

//Build 包含構建配置。
type Build struct {
	Main    string   `yaml:"main"`
	Binary  string   `yaml:"binary"`
	LDFlags []string `yaml:"ldflags"`
	Proto   Proto    `yaml:"proto"`
}

// Proto 包含 proto 構建配置。
type Proto struct {
	// Path 是應用程序的 proto 文件所在的相對路徑。
	Path string `yaml:"path"`

	// ThirdPartyPath 是第三方 proto 文件所在的相對路徑
	// 位於應用程序使用的位置。
	ThirdPartyPaths []string `yaml:"third_party_paths"`
}

// 客戶端為客戶端配置代碼生成。
type Client struct {
	// Vuex 為 Vuex 配置代碼生成。
	Vuex Vuex `yaml:"vuex"`

	// Dart 為 Dart 配置客戶端代碼生成。
	Dart Dart `yaml:"dart"`

	// OpenAPI 為 API 配置 OpenAPI 規範生成。
	OpenAPI OpenAPI `yaml:"openapi"`
}

// Vuex 為 Vuex 配置代碼生成。
type Vuex struct {
	// Path 為生成的 Vuex 代碼配置出位置。
	Path string `yaml:"path"`
}

// Dart 為 Dart 配置客戶端代碼生成。
type Dart struct {
	// Path 配置生成的 Dart 代碼的位置。
	Path string `yaml:"path"`
}

// OpenAPI 為 API 配置 OpenAPI 規範生成。
type OpenAPI struct {
	Path string `yaml:"path"`
}

// 水龍頭配置。
type Faucet struct {
	// 名稱是水龍頭帳戶的名稱。
	Name *string `yaml:"name"`

	// 硬幣持有硬幣面額的類型和分發的數量。
	Coins []string `yaml:"coins"`

	// CoinsMax 持有鏈上的面額及其可以轉移的最大數量
	// 給單個用戶。
	CoinsMax []string `yaml:"coins_max"`

	// LimitRefreshTime 設置在結束時刷新限制的時間範圍
	RateLimitWindow string `yaml:"rate_limit_window"`

	// Host 是水龍頭服務器的主機
	Host string `yaml:"host"`

	// 水龍頭服務器要監聽的端口號。
	Port int `yaml:"port"`
}

// Init 用給定的值覆蓋 sdk 配置。
type Init struct {
	// 應用程序覆蓋 appd 的 config/app.toml 配置。
	App map[string]interface{} `yaml:"app"`

	// 客戶端覆蓋 appd 的 config/client.toml 配置.
	Client map[string]interface{} `yaml:"client"`

	// 配置覆蓋 appd 的 config/config.toml 配置。
	Config map[string]interface{} `yaml:"config"`

	// Home覆蓋默認值home應用程序使用的目錄
	Home string `yaml:"home"`

	// KeyringBackend 是用於區塊鏈初始化的默認密鑰環後端
	KeyringBackend string `yaml:"keyring-backend"`
}

// 主機保留與已啟動服務器相關的配置。
type Host struct {
	RPC     string `yaml:"rpc"`
	P2P     string `yaml:"p2p"`
	Prof    string `yaml:"prof"`
	GRPC    string `yaml:"grpc"`
	GRPCWeb string `yaml:"grpc-web"`
	API     string `yaml:"api"`
}

// Parse 將 config.yml 解析為 UserConfig。
func Parse(r io.Reader) (Config, error) {
	var conf Config
	if err := yaml.NewDecoder(r).Decode(&conf); err != nil {
		return conf, err
	}
	if err := mergo.Merge(&conf, DefaultConf); err != nil {
		return Config{}, err
	}
	return conf, validate(conf)
}

// ParseFile 從路徑中解析 config.yml。
func ParseFile(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, nil
	}
	defer file.Close()
	return Parse(file)
}

// validate 驗證用戶配置。
func validate(conf Config) error {
	if len(conf.Accounts) == 0 {
		return &ValidationError{"at least 1 account is needed"}
	}
	if conf.Validator.Name == "" {
		return &ValidationError{"validator is required"}
	}
	return nil
}

// 當配置無效時返回 ValidationError。
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("config is not valid: %s", e.Message)
}

// LocateDefault 定位配置文件的默認路徑，如果沒有找到文件返回 ErrCouldntLocateConfig。
func LocateDefault(root string) (path string, err error) {
	for _, name := range ConfigFileNames {
		path = filepath.Join(root, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	return "", ErrCouldntLocateConfig
}

// FaucetHost 返回要使用的水龍頭主機
func FaucetHost(conf Config) string {
	// 我們繼續支持端口選項以實現向後兼容性
	// TODO：以後放棄這個選項
	host := conf.Faucet.Host
	if conf.Faucet.Port != 0 {
		host = fmt.Sprintf(":%d", conf.Faucet.Port)
	}

	return host
}

// 如果尚未創建，則 CreateConfigDir 創建配置目錄。
func CreateConfigDir() error {
	confPath, err := ConfigDirPath()
	if err != nil {
		return err
	}

	return os.MkdirAll(confPath, 0755)
}
