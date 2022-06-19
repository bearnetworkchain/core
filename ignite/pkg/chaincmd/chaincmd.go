package chaincmd

import (
	"fmt"

	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosver"
)

const (
	commandStart             = "start"
	commandInit              = "init"
	commandKeys              = "keys"
	commandAddGenesisAccount = "add-genesis-account"
	commandGentx             = "gentx"
	commandCollectGentxs     = "collect-gentxs"
	commandValidateGenesis   = "validate-genesis"
	commandShowNodeID        = "show-node-id"
	commandStatus            = "status"
	commandTx                = "tx"
	commandQuery             = "query"
	commandUnsafeReset       = "unsafe-reset-all"
	commandExport            = "export"
	commandTendermint        = "tendermint"

	optionHome                             = "--home"
	optionNode                             = "--node"
	optionKeyringBackend                   = "--keyring-backend"
	optionChainID                          = "--chain-id"
	optionOutput                           = "--output"
	optionRecover                          = "--recover"
	optionAddress                          = "--address"
	optionAmount                           = "--amount"
	optionValidatorMoniker                 = "--moniker"
	optionValidatorCommissionRate          = "--commission-rate"
	optionValidatorCommissionMaxRate       = "--commission-max-rate"
	optionValidatorCommissionMaxChangeRate = "--commission-max-change-rate"
	optionValidatorMinSelfDelegation       = "--min-self-delegation"
	optionValidatorGasPrices               = "--gas-prices"
	optionValidatorDetails                 = "--details"
	optionValidatorIdentity                = "--identity"
	optionValidatorWebsite                 = "--website"
	optionValidatorSecurityContact         = "--security-contact"
	optionYes                              = "--yes"
	optionHomeClient                       = "--home-client"
	optionCoinType                         = "--coin-type"
	optionVestingAmount                    = "--vesting-amount"
	optionVestingEndTime                   = "--vesting-end-time"
	optionBroadcastMode                    = "--broadcast-mode"

	constTendermint = "tendermint"
	constJSON       = "json"
	constSync       = "sync"
)

type KeyringBackend string

const (
	KeyringBackendUnspecified KeyringBackend = ""
	KeyringBackendOS          KeyringBackend = "os"
	KeyringBackendFile        KeyringBackend = "file"
	KeyringBackendPass        KeyringBackend = "pass"
	KeyringBackendTest        KeyringBackend = "test"
	KeyringBackendKwallet     KeyringBackend = "kwallet"
)

type ChainCmd struct {
	appCmd          string
	chainID         string
	homeDir         string
	keyringBackend  KeyringBackend
	keyringPassword string
	cliCmd          string
	cliHome         string
	nodeAddress     string
	legacySend      bool

	isAutoChainIDDetectionEnabled bool

	sdkVersion cosmosver.Version
}

// New 創建一個新的 ChainCmd 以使用鏈應用程序啟動命令
func New(appCmd string, options ...Option) ChainCmd {
	chainCmd := ChainCmd{
		appCmd:     appCmd,
		sdkVersion: cosmosver.Latest,
	}

	applyOptions(&chainCmd, options)

	return chainCmd
}

// Copy 通過使用給定選項覆蓋其選項來複製 ChainCmd。
func (c ChainCmd) Copy(options ...Option) ChainCmd {
	applyOptions(&c, options)

	return c
}

//選項配置 ChainCmd。
type Option func(*ChainCmd)

func applyOptions(c *ChainCmd, options []Option) {
	for _, applyOption := range options {
		applyOption(c)
	}
}

// WithVersion 設置區塊鏈的版本。
// 如果未提供，則假定為最新版本的 SDK。
func WithVersion(v cosmosver.Version) Option {
	return func(c *ChainCmd) {
		c.sdkVersion = v
	}
}

// WithHome 替換了鏈使用的默認主頁
func WithHome(home string) Option {
	return func(c *ChainCmd) {
		c.homeDir = home
	}
}

// WithChainID 為接受此選項的命令提供特定的鏈 ID
func WithChainID(chainID string) Option {
	return func(c *ChainCmd) {
		c.chainID = chainID
	}
}

// WithAutoChainIDDetection 通過與運行的節點通信找出鏈 id。
func WithAutoChainIDDetection() Option {
	return func(c *ChainCmd) {
		c.isAutoChainIDDetectionEnabled = true
	}
}

// WithKeyringBackend 為接受此選項的命令提供特定的密鑰環後端
func WithKeyringBackend(keyringBackend KeyringBackend) Option {
	return func(c *ChainCmd) {
		c.keyringBackend = keyringBackend
	}
}

// WithKeyringPassword 提供了解鎖密鑰環的密碼
func WithKeyringPassword(password string) Option {
	return func(c *ChainCmd) {
		c.keyringPassword = password
	}
}

// WithNodeAddress 為需要生成的命令設置節點地址
// 向具有不同於默認節點地址的節點的 API 請求。
func WithNodeAddress(addr string) Option {
	return func(c *ChainCmd) {
		c.nodeAddress = addr
	}
}

// WithLaunchpadCLI 為區塊鏈提供 CLI 應用程序名稱
// 這對於 Launchpad 應用程序是必需的，因為它有兩個不同的二進製文件，但是
// Stargate 應用程序不需要
func WithLaunchpadCLI(cliCmd string) Option {
	return func(c *ChainCmd) {
		c.cliCmd = cliCmd
	}
}

// WithLaunchpadCLIHome 替換了 Launchpad 鏈 CLI 使用的默認主頁
func WithLaunchpadCLIHome(cliHome string) Option {
	return func(c *ChainCmd) {
		c.cliHome = cliHome
	}
}

// WithLegacySendCommand 將使命令使用來自啟動板的傳統 tx 發送語法
// 在星門鏈上。例如：CosmWasm
func WithLegacySendCommand() Option {
	return func(c *ChainCmd) {
		c.legacySend = true
	}
}

// StartCommand 返回啟動鏈的守護進程的命令
func (c ChainCmd) StartCommand(options ...string) step.Option {
	command := append([]string{
		commandStart,
	}, options...)
	return c.daemonCommand(command)
}

// InitCommand 返回初始化鏈的命令
func (c ChainCmd) InitCommand(moniker string) step.Option {
	command := []string{
		commandInit,
		moniker,
	}
	command = c.attachChainID(command)
	return c.daemonCommand(command)
}

// AddKeyCommand 返回在鏈密鑰環中添加新密鑰的命令
func (c ChainCmd) AddKeyCommand(accountName, coinType string) step.Option {
	command := []string{
		commandKeys,
		"add",
		accountName,
		optionOutput,
		constJSON,
	}
	if coinType != "" {
		command = append(command, optionCoinType, coinType)
	}
	command = c.attachKeyringBackend(command)

	return c.cliCommand(command)
}

// RecoverKeyCommand 返回從助記詞中將密鑰恢復到鏈密鑰環中的命令
func (c ChainCmd) RecoverKeyCommand(accountName, coinType string) step.Option {
	command := []string{
		commandKeys,
		"add",
		accountName,
		optionRecover,
	}
	if coinType != "" {
		command = append(command, optionCoinType, coinType)
	}
	command = c.attachKeyringBackend(command)

	return c.cliCommand(command)
}

// ImportKeyCommand 返回將密鑰從密鑰文件導入到鍊式密鑰環的命令
func (c ChainCmd) ImportKeyCommand(accountName, keyFile string) step.Option {
	command := []string{
		commandKeys,
		"import",
		accountName,
		keyFile,
	}
	command = c.attachKeyringBackend(command)

	return c.cliCommand(command)
}

// ShowKeyAddressCommand 返回命令以打印鏈密鑰環中密鑰的地址
func (c ChainCmd) ShowKeyAddressCommand(accountName string) step.Option {
	command := []string{
		commandKeys,
		"show",
		accountName,
		optionAddress,
	}
	command = c.attachKeyringBackend(command)

	return c.cliCommand(command)
}

// ListKeysCommand 返回命令以打印鏈密鑰環中的密鑰列表
func (c ChainCmd) ListKeysCommand() step.Option {
	command := []string{
		commandKeys,
		"list",
		optionOutput,
		constJSON,
	}
	command = c.attachKeyringBackend(command)

	return c.cliCommand(command)
}

// AddGenesisAccountCommand 返回在鏈的創世文件中添加新賬戶的命令
func (c ChainCmd) AddGenesisAccountCommand(address, coins string) step.Option {
	command := []string{
		commandAddGenesisAccount,
		address,
		coins,
	}

	return c.daemonCommand(command)
}

// AddVestingAccountCommand 返回在鏈的創世文件中添加延遲歸屬賬戶的命令
func (c ChainCmd) AddVestingAccountCommand(address, originalCoins, vestingCoins string, vestingEndTime int64) step.Option {
	command := []string{
		commandAddGenesisAccount,
		address,
		originalCoins,
		optionVestingAmount,
		vestingCoins,
		optionVestingEndTime,
		fmt.Sprintf("%d", vestingEndTime),
	}

	return c.daemonCommand(command)
}

// GentxCommand 的 GentxOption
type GentxOption func([]string) []string

// GentxWithMoniker 為 gentx 命令提供 moniker 選項
func GentxWithMoniker(moniker string) GentxOption {
	return func(command []string) []string {
		if len(moniker) > 0 {
			return append(command, optionValidatorMoniker, moniker)
		}
		return command
	}
}

// GentxWithCommissionRate 為 gentx 命令提供佣金率選項
func GentxWithCommissionRate(commissionRate string) GentxOption {
	return func(command []string) []string {
		if len(commissionRate) > 0 {
			return append(command, optionValidatorCommissionRate, commissionRate)
		}
		return command
	}
}

// GentxWithCommissionMaxRate 為 gentx 命令提供佣金最高費率選項
func GentxWithCommissionMaxRate(commissionMaxRate string) GentxOption {
	return func(command []string) []string {
		if len(commissionMaxRate) > 0 {
			return append(command, optionValidatorCommissionMaxRate, commissionMaxRate)
		}
		return command
	}
}

// GentxWithCommissionMaxChangeRate 為 gentx 命令提供佣金最大變化率選項
func GentxWithCommissionMaxChangeRate(commissionMaxChangeRate string) GentxOption {
	return func(command []string) []string {
		if len(commissionMaxChangeRate) > 0 {
			return append(command, optionValidatorCommissionMaxChangeRate, commissionMaxChangeRate)
		}
		return command
	}
}

// GentxWithMinSelfDelegation 為 gentx 命令提供最小自我委託選項
func GentxWithMinSelfDelegation(minSelfDelegation string) GentxOption {
	return func(command []string) []string {
		if len(minSelfDelegation) > 0 {
			return append(command, optionValidatorMinSelfDelegation, minSelfDelegation)
		}
		return command
	}
}

// GentxWithGasPrices 為 gentx 命令提供 gas 價格選項
func GentxWithGasPrices(gasPrices string) GentxOption {
	return func(command []string) []string {
		if len(gasPrices) > 0 {
			return append(command, optionValidatorGasPrices, gasPrices)
		}
		return command
	}
}

// GentxWithDetails 為 gentx 命令提供驗證器詳細信息選項
func GentxWithDetails(details string) GentxOption {
	return func(command []string) []string {
		if len(details) > 0 {
			return append(command, optionValidatorDetails, details)
		}
		return command
	}
}

// GentxWithIdentity 為 gentx 命令提供驗證者身份選項
func GentxWithIdentity(identity string) GentxOption {
	return func(command []string) []string {
		if len(identity) > 0 {
			return append(command, optionValidatorIdentity, identity)
		}
		return command
	}
}

// GentxWithWebsite 為 gentx 命令提供驗證器網站選項
func GentxWithWebsite(website string) GentxOption {
	return func(command []string) []string {
		if len(website) > 0 {
			return append(command, optionValidatorWebsite, website)
		}
		return command
	}
}

// GentxWithSecurityContact 為 gentx 命令提供驗證器安全聯繫人選項
func GentxWithSecurityContact(securityContact string) GentxOption {
	return func(command []string) []string {
		if len(securityContact) > 0 {
			return append(command, optionValidatorSecurityContact, securityContact)
		}
		return command
	}
}

func (c ChainCmd) IsAutoChainIDDetectionEnabled() bool {
	return c.isAutoChainIDDetectionEnabled
}

func (c ChainCmd) SDKVersion() cosmosver.Version {
	return c.sdkVersion
}

// GentxCommand 返回為鏈生成 gentx 的命令
func (c ChainCmd) GentxCommand(
	validatorName string,
	selfDelegation string,
	options ...GentxOption,
) step.Option {
	command := []string{
		commandGentx,
	}

	switch {
	case c.sdkVersion.LT(cosmosver.StargateFortyVersion):
		command = append(command,
			validatorName,
			optionAmount,
			selfDelegation,
		)
	case c.sdkVersion.GTE(cosmosver.StargateFortyVersion):
		command = append(command,
			validatorName,
			selfDelegation,
		)
	case c.sdkVersion.LTE(cosmosver.MaxLaunchpadVersion):
		command = append(command,
			optionName,
			validatorName,
			optionAmount,
			selfDelegation,
		)

		// 附加HOME客戶端選項
		if c.cliHome != "" {
			command = append(command, []string{optionHomeClient, c.cliHome}...)
		}
	}

	// 應用用戶提供的選項
	for _, applyOption := range options {
		command = applyOption(command)
	}

	// 添加必要的 flags
	if c.sdkVersion.IsFamily(cosmosver.Stargate) {
		command = c.attachChainID(command)
	}

	command = c.attachKeyringBackend(command)

	return c.daemonCommand(command)
}

// CollectGentxsCommand 返回命令將 /gentx 目錄中的 gentxs 收集到鏈的創世紀文件中
func (c ChainCmd) CollectGentxsCommand() step.Option {
	command := []string{
		commandCollectGentxs,
	}
	return c.daemonCommand(command)
}

// ValidateGenesisCommand 返回檢查鏈創世有效性的命令
func (c ChainCmd) ValidateGenesisCommand() step.Option {
	command := []string{
		commandValidateGenesis,
	}
	return c.daemonCommand(command)
}

// ShowNodeIDCommand返回命令以打印鏈的節點的節點 ID
func (c ChainCmd) ShowNodeIDCommand() step.Option {
	command := []string{
		constTendermint,
		commandShowNodeID,
	}
	return c.daemonCommand(command)
}

// UnsafeResetCommand返回重置區塊鏈數據庫的命令
func (c ChainCmd) UnsafeResetCommand() step.Option {
	var command []string

	if c.sdkVersion.GTE(cosmosver.StargateFortyFiveThreeVersion) {
		command = append(command, commandTendermint)
	}

	command = append(command, commandUnsafeReset)

	return c.daemonCommand(command)
}

// ExportCommand 返回將區塊鏈狀態導出到創世文件的命令
func (c ChainCmd) ExportCommand() step.Option {
	command := []string{
		commandExport,
	}
	return c.daemonCommand(command)
}

// BankSendCommand 返回用於傳輸令牌的命令。
func (c ChainCmd) BankSendCommand(fromAddress, toAddress, amount string) step.Option {
	command := []string{
		commandTx,
	}

	if c.sdkVersion.IsFamily(cosmosver.Stargate) && !c.legacySend {
		command = append(command,
			"bank",
		)
	}

	command = append(command,
		"send",
		fromAddress,
		toAddress,
		amount,
		optionBroadcastMode,
		constSync,
		optionYes,
	)

	command = c.attachChainID(command)
	command = c.attachKeyringBackend(command)
	command = c.attachNode(command)

	if c.sdkVersion.IsFamily(cosmosver.Launchpad) {
		command = append(command, optionOutput, constJSON)
	}

	return c.cliCommand(command)
}

// QueryTxCommand 返回查詢tx的命令
func (c ChainCmd) QueryTxCommand(txHash string) step.Option {
	command := []string{
		commandQuery,
		"tx",
		txHash,
	}

	command = c.attachNode(command)
	return c.cliCommand(command)
}

// QueryTxEventsCommand 返回查詢事件的命令。
func (c ChainCmd) QueryTxEventsCommand(query string) step.Option {
	command := []string{
		commandQuery,
		"txs",
		"--events",
		query,
		"--page", "1",
		"--limit", "1000",
	}

	if c.sdkVersion.IsFamily(cosmosver.Launchpad) {
		command = append(command,
			"--trust-node",
		)
	}

	command = c.attachNode(command)
	return c.cliCommand(command)
}

// LaunchpadSetConfigCommand 返回設置配置值的命令
func (c ChainCmd) LaunchpadSetConfigCommand(name, value string) step.Option {
	// Check version
	if c.isStargate() {
		panic("Stargate 的配置命令不存在")
	}
	return c.launchpadSetConfigCommand(name, value)
}

// LaunchpadRestServerCommand 返回啟動 CLI REST 服務器的命令
func (c ChainCmd) LaunchpadRestServerCommand(apiAddress, rpcAddress string) step.Option {
	// Check version
	if c.isStargate() {
		panic("rest-server Stargate 的命令不存在")
	}
	return c.launchpadRestServerCommand(apiAddress, rpcAddress)
}

// StatusCommand 返回獲取節點狀態的命令.
func (c ChainCmd) StatusCommand() step.Option {
	command := []string{
		commandStatus,
	}

	command = c.attachNode(command)
	return c.cliCommand(command)
}

// KeyringBackend 返回底層密鑰環後端.
func (c ChainCmd) KeyringBackend() KeyringBackend {
	return c.keyringBackend
}

// KeyringPassword 返回底層密鑰環密碼.
func (c ChainCmd) KeyringPassword() string {
	return c.keyringPassword
}

// attachChainID 將Chain ID 標誌附加到提供的命令
func (c ChainCmd) attachChainID(command []string) []string {
	if c.chainID != "" {
		command = append(command, []string{optionChainID, c.chainID}...)
	}
	return command
}

// attachKeyringBackend 將密鑰環後端標誌附加到提供的命令
func (c ChainCmd) attachKeyringBackend(command []string) []string {
	if c.keyringBackend != "" {
		command = append(command, []string{optionKeyringBackend, string(c.keyringBackend)}...)
	}
	return command
}

// attachHome 將 home 標誌附加到提供的命令
func (c ChainCmd) attachHome(command []string) []string {
	if c.homeDir != "" {
		command = append(command, []string{optionHome, c.homeDir}...)
	}
	return command
}

// attachNode 將節點標誌附加到提供的命令
func (c ChainCmd) attachNode(command []string) []string {
	if c.nodeAddress != "" {
		command = append(command, []string{optionNode, c.nodeAddress}...)
	}
	return command
}

// isStargate 檢查命令的版本是否為星門
func (c ChainCmd) isStargate() bool {
	return c.sdkVersion.Family == cosmosver.Stargate
}

// daemonCommand 從提供的命令返回守護程序命令
func (c ChainCmd) daemonCommand(command []string) step.Option {
	return step.Exec(c.appCmd, c.attachHome(command)...)
}

// cliCommand 從提供的命令返回 cli 命令
// cli 是星際之門的守護進程
func (c ChainCmd) cliCommand(command []string) step.Option {
	//檢查版本
	if c.isStargate() {
		return step.Exec(c.appCmd, c.attachHome(command)...)
	}
	return step.Exec(c.cliCmd, c.attachCLIHome(command)...)
}

// KeyringBackendFromString 從其字符串返回密鑰環後端
func KeyringBackendFromString(kb string) (KeyringBackend, error) {
	existingKeyringBackend := map[KeyringBackend]bool{
		KeyringBackendUnspecified: true,
		KeyringBackendOS:          true,
		KeyringBackendFile:        true,
		KeyringBackendPass:        true,
		KeyringBackendTest:        true,
		KeyringBackendKwallet:     true,
	}

	if _, ok := existingKeyringBackend[KeyringBackend(kb)]; ok {
		return KeyringBackend(kb), nil
	}
	return KeyringBackendUnspecified, fmt.Errorf("無法識別的密鑰環後端: %s", kb)
}
