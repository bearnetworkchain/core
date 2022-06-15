package chain

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/gookit/color"
	"github.com/tendermint/spn/pkg/chainid"

	"github.com/ignite-hq/cli/ignite/chainconfig"
	sperrors "github.com/ignite-hq/cli/ignite/errors"
	"github.com/ignite-hq/cli/ignite/pkg/chaincmd"
	chaincmdrunner "github.com/ignite-hq/cli/ignite/pkg/chaincmd/runner"
	"github.com/ignite-hq/cli/ignite/pkg/confile"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosver"
	"github.com/ignite-hq/cli/ignite/pkg/repoversion"
	"github.com/ignite-hq/cli/ignite/pkg/xurl"
)

var (
	appBackendSourceWatchPaths = []string{
		"app",
		"cmd",
		"x",
		"proto",
		"third_party",
	}

	errorColor = color.Red.Render
	infoColor  = color.Yellow.Render
)

type version struct {
	tag  string
	hash string
}

type LogLvl int

const (
	LogSilent LogLvl = iota
	LogRegular
	LogVerbose
)

//Chain 為 Cosmos SDK 區塊鏈提供編程訪問和工具。
type Chain struct {
	// 應用程序保存有關區塊鏈應用程序的信息。
	app App

	options chainOptions

	Version cosmosver.Version

	plugin         Plugin
	sourceVersion  version
	logLevel       LogLvl
	serveCancel    context.CancelFunc
	serveRefresher chan struct{}
	served         bool

	// protoBuiltAtLeastOnce 表示應用程序的原型生成至少生成一次。
	protoBuiltAtLeastOnce bool

	stdout, stderr io.Writer
}

// chainOptions 包含覆蓋鏈默認值的用戶給定選項。
type chainOptions struct {
	// chainID is鏈的 ID。
	chainID string

	// homePath鏈的配置目錄。
	homePath string

	// 如果未在配置中指定，則命令使用的密鑰環後端
	keyringBackend chaincmd.KeyringBackend

// isThirdPartyModuleCodegen 指示是否應該生成 proto 代碼
// 對於第 3 方模塊。 SDK 模塊也被視為第 3 方。
	isThirdPartyModuleCodegenEnabled bool

	// 自定義配置文件的路徑
	ConfigFile string
}

// 選項配置鏈。
type Option func(*Chain)

//LogLevel 設置日誌記錄級別。
func LogLevel(level LogLvl) Option {
	return func(c *Chain) {
		c.logLevel = level
	}
}

//ID 用給定的 id 替換鏈的 id。
func ID(id string) Option {
	return func(c *Chain) {
		c.options.chainID = id
	}
}

// HomePath用給定路徑替換鏈的配置主路徑。
func HomePath(path string) Option {
	return func(c *Chain) {
		c.options.homePath = path
	}
}

// KeyringBackend指定用於鏈命令的密鑰環後端
func KeyringBackend(keyringBackend chaincmd.KeyringBackend) Option {
	return func(c *Chain) {
		c.options.keyringBackend = keyringBackend
	}
}

// ConfigFile指定要使用的自定義配置文件
func ConfigFile(configFile string) Option {
	return func(c *Chain) {
		c.options.ConfigFile = configFile
	}
}

// EnableThirdPartyModuleCodegen 啟用第三方模塊的代碼生成，
// 包括 SDK。
func EnableThirdPartyModuleCodegen() Option {
	return func(c *Chain) {
		c.options.isThirdPartyModuleCodegenEnabled = true
	}
}

// New 使用其源位於路徑的選項初始化一個新鏈。
func New(path string, options ...Option) (*Chain, error) {
	app, err := NewAppAt(path)
	if err != nil {
		return nil, err
	}

	c := &Chain{
		app:            app,
		logLevel:       LogSilent,
		serveRefresher: make(chan struct{}, 1),
		stdout:         io.Discard,
		stderr:         io.Discard,
	}

	// 應用選項
	for _, apply := range options {
		apply(c)
	}

	if c.logLevel == LogVerbose {
		c.stdout = os.Stdout
		c.stderr = os.Stderr
	}

	c.sourceVersion, err = c.appVersion()
	if err != nil && err != git.ErrRepositoryNotExists {
		return nil, err
	}

	c.Version, err = cosmosver.Detect(c.app.Path)
	if err != nil {
		return nil, err
	}

	if !c.Version.IsFamily(cosmosver.Stargate) {
		return nil, sperrors.ErrOnlyStargateSupported
	}

	// 根據鏈的版本初始化插件
	c.plugin = c.pickPlugin()

	return c, nil
}

func (c *Chain) appVersion() (v version, err error) {

	ver, err := repoversion.Determine(c.app.Path)
	if err != nil {
		return version{}, err
	}

	v.hash = ver.Hash
	v.tag = ver.Tag

	return v, nil
}

// RPCPublicAddress 指向 Tendermint RPC 的公共地址，這是由
// 中繼器相關操作的其他鏈。
func (c *Chain) RPCPublicAddress() (string, error) {
	rpcAddress := os.Getenv("RPC_ADDRESS")
	if rpcAddress == "" {
		conf, err := c.Config()
		if err != nil {
			return "", err
		}
		rpcAddress = conf.Host.RPC
	}
	return rpcAddress, nil
}

// ConfigPath 返回鏈的配置路徑
// 空字符串表示鏈沒有定義的配置
func (c *Chain) ConfigPath() string {
	if c.options.ConfigFile != "" {
		return c.options.ConfigFile
	}
	path, err := chainconfig.LocateDefault(c.app.Path)
	if err != nil {
		return ""
	}
	return path
}

// Config 返回鏈的配置
func (c *Chain) Config() (chainconfig.Config, error) {
	configPath := c.ConfigPath()
	if configPath == "" {
		return chainconfig.DefaultConf, nil
	}
	return chainconfig.ParseFile(configPath)
}

// ID 返回鏈的 id。
func (c *Chain) ID() (string, error) {
	// chainID in App has the most priority.
	if c.options.chainID != "" {
		return c.options.chainID, nil
	}

	// 否則定義的用途config.yml
	chainConfig, err := c.Config()
	if err != nil {
		return "", err
	}
	genid, ok := chainConfig.Genesis["chain_id"]
	if ok {
		return genid.(string), nil
	}

	// 默認使用應用名稱。
	return c.app.N(), nil
}

// ChainID 返回默認網絡鏈的 id。
func (c *Chain) ChainID() (string, error) {
	chainID, err := c.ID()
	if err != nil {
		return "", err
	}
	return chainid.NewGenesisChainID(chainID, 1), nil
}

// Name 返回鏈的名稱
func (c *Chain) Name() string {
	return c.app.N()
}

// Binary 返回應用程序的默認 (appd) 二進製文件的名稱。
func (c *Chain) Binary() (string, error) {
	conf, err := c.Config()
	if err != nil {
		return "", err
	}

	if conf.Build.Binary != "" {
		return conf.Build.Binary, nil
	}

	return c.app.D(), nil
}

// SetHome 設置鍊主目錄。
func (c *Chain) SetHome(home string) {
	c.options.homePath = home
}

// Home 返回區塊鏈節點的主目錄。
func (c *Chain) Home() (string, error) {
	// 檢查是否為應用明確定義了主頁
	home := c.options.homePath
	if home == "" {
		// 否則返回默認主頁
		var err error
		home, err = c.DefaultHome()
		if err != nil {
			return "", err
		}

	}

	// 擴展環境變量 home
	home = os.ExpandEnv(home)

	return home, nil
}

// DefaultHome應用程序中未指定時返回區塊鏈節點的默認主目錄
func (c *Chain) DefaultHome() (string, error) {
	// check if home is defined in config
	config, err := c.Config()
	if err != nil {
		return "", err
	}
	if config.Init.Home != "" {
		return config.Init.Home, nil
	}

	return c.plugin.Home(), nil
}

// DefaultGentxPath返回應用程序的默認 gentx.json 路徑。
func (c *Chain) DefaultGentxPath() (string, error) {
	home, err := c.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config/gentx/gentx.json"), nil
}

// GenesisPath返回應用程序的 genesis.json 路徑。
func (c *Chain) GenesisPath() (string, error) {
	home, err := c.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config/genesis.json"), nil
}

// GentxsPath 返回為應用程序存儲 gentxs 的目錄。
func (c *Chain) GentxsPath() (string, error) {
	home, err := c.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config/gentx"), nil
}

// AppTOMLPath 返回應用程序的 app.toml 路徑。
func (c *Chain) AppTOMLPath() (string, error) {
	home, err := c.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config/app.toml"), nil
}

// ConfigTOMLPath 返回應用程序的 config.toml 路徑。
func (c *Chain) ConfigTOMLPath() (string, error) {
	home, err := c.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config/config.toml"), nil
}

// ClientTOMLPath 返回應用程序的 client.toml 路徑。
func (c *Chain) ClientTOMLPath() (string, error) {
	home, err := c.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config/client.toml"), nil
}

// KeyringBackend 返回為鏈選擇的密鑰環後端。
func (c *Chain) KeyringBackend() (chaincmd.KeyringBackend, error) {
	// 1st.
	if c.options.keyringBackend != "" {
		return c.options.keyringBackend, nil
	}

	config, err := c.Config()
	if err != nil {
		return "", err
	}

	// 2nd.
	if config.Init.KeyringBackend != "" {
		return chaincmd.KeyringBackendFromString(config.Init.KeyringBackend)
	}

	// 3rd.
	if config.Init.Client != nil {
		if backend, ok := config.Init.Client["keyring-backend"]; ok {
			if backendStr, ok := backend.(string); ok {
				return chaincmd.KeyringBackendFromString(backendStr)
			}
		}
	}

	// 4th.
	configTOMLPath, err := c.ClientTOMLPath()
	if err != nil {
		return "", err
	}
	cf := confile.New(confile.DefaultTOMLEncodingCreator, configTOMLPath)
	var conf struct {
		KeyringBackend string `toml:"keyring-backend"`
	}
	if err := cf.Load(&conf); err != nil {
		return "", err
	}
	if conf.KeyringBackend != "" {
		return chaincmd.KeyringBackendFromString(conf.KeyringBackend)
	}

	// 5th.
	return chaincmd.KeyringBackendTest, nil
}

// Commands 返回運行者在鏈的二進製文件上執行命令
func (c *Chain) Commands(ctx context.Context) (chaincmdrunner.Runner, error) {
	id, err := c.ID()
	if err != nil {
		return chaincmdrunner.Runner{}, err
	}

	home, err := c.Home()
	if err != nil {
		return chaincmdrunner.Runner{}, err
	}

	binary, err := c.Binary()
	if err != nil {
		return chaincmdrunner.Runner{}, err
	}

	backend, err := c.KeyringBackend()
	if err != nil {
		return chaincmdrunner.Runner{}, err
	}

	config, err := c.Config()
	if err != nil {
		return chaincmdrunner.Runner{}, err
	}

	nodeAddr, err := xurl.TCP(config.Host.RPC)
	if err != nil {
		return chaincmdrunner.Runner{}, err
	}

	chainCommandOptions := []chaincmd.Option{
		chaincmd.WithChainID(id),
		chaincmd.WithHome(home),
		chaincmd.WithVersion(c.Version),
		chaincmd.WithNodeAddress(nodeAddr),
		chaincmd.WithKeyringBackend(backend),
	}

	cc := chaincmd.New(binary, chainCommandOptions...)

	ccrOptions := make([]chaincmdrunner.Option, 0)
	if c.logLevel == LogVerbose {
		ccrOptions = append(ccrOptions,
			chaincmdrunner.Stdout(os.Stdout),
			chaincmdrunner.Stderr(os.Stderr),
			chaincmdrunner.DaemonLogPrefix(c.genPrefix(logAppd)),
		)
	}

	return chaincmdrunner.New(ctx, cc, ccrOptions...)
}
