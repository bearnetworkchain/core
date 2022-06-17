package networkchain

import (
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	sperrors "github.com/bearnetworkchain/core/ignite/errors"
	"github.com/bearnetworkchain/core/ignite/pkg/cache"
	"github.com/bearnetworkchain/core/ignite/pkg/chaincmd"
	"github.com/bearnetworkchain/core/ignite/pkg/checksum"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosaccount"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosver"
	"github.com/bearnetworkchain/core/ignite/pkg/events"
	"github.com/bearnetworkchain/core/ignite/pkg/gitpod"
	"github.com/bearnetworkchain/core/ignite/services/chain"
	"github.com/bearnetworkchain/core/ignite/services/network/networktypes"
)

// Chain 代表一個網絡區塊鏈，並允許您與其源代碼和二進製文件進行交互。
type Chain struct {
	id       string
	launchID uint64

	path string
	home string

	url         string
	hash        string
	genesisURL  string
	genesisHash string
	launchTime  int64

	keyringBackend chaincmd.KeyringBackend

	isInitialized bool

	ref plumbing.ReferenceName

	chain *chain.Chain
	ev    events.Bus
	ar    cosmosaccount.Registry
}

//SourceOption 設置區塊鏈的來源。
type SourceOption func(*Chain)

//選項設置其他初始化選項。
type Option func(*Chain)

//SourceRemote 將遠程上的默認分支設置為區塊鏈的源。
func SourceRemote(url string) SourceOption {
	return func(c *Chain) {
		c.url = url
	}
}

// SourceRemoteBranch將遠程分支設置為區塊鏈的源。
func SourceRemoteBranch(url, branch string) SourceOption {
	return func(c *Chain) {
		c.url = url
		c.ref = plumbing.NewBranchReferenceName(branch)
	}
}

// SourceRemoteTag將遠程標籤設置為區塊鏈的源。
func SourceRemoteTag(url, tag string) SourceOption {
	return func(c *Chain) {
		c.url = url
		c.ref = plumbing.NewTagReferenceName(tag)
	}
}

// SourceRemoteHash使用遠程哈希作為區塊鏈的來源。
func SourceRemoteHash(url, hash string) SourceOption {
	return func(c *Chain) {
		c.url = url
		c.hash = hash
	}
}

// SourceLaunch返回用於從啟動初始化鏈的源選項
func SourceLaunch(launch networktypes.ChainLaunch) SourceOption {
	return func(c *Chain) {
		c.id = launch.ChainID
		c.launchID = launch.ID
		c.url = launch.SourceURL
		c.hash = launch.SourceHash
		c.genesisURL = launch.GenesisURL
		c.genesisHash = launch.GenesisHash
		c.home = ChainHome(launch.ID)
		c.launchTime = launch.LaunchTime
	}
}

// WithHome為初始化的區塊鏈提供特定的主路徑。
func WithHome(path string) Option {
	return func(c *Chain) {
		c.home = path
	}
}

// WithKeyringBackend提供用於初始化區塊鏈的密鑰環後端
func WithKeyringBackend(keyringBackend chaincmd.KeyringBackend) Option {
	return func(c *Chain) {
		c.keyringBackend = keyringBackend
	}
}

// WithGenesisFromURL為鏈區塊鏈的初始創世提供創世 URL
func WithGenesisFromURL(genesisURL string) Option {
	return func(c *Chain) {
		c.genesisURL = genesisURL
	}
}

// CollectEvents從鏈中收集事件。
func CollectEvents(ev events.Bus) Option {
	return func(c *Chain) {
		c.ev = ev
	}
}

// New initializes來自源和選項的網絡區塊鏈。
func New(ctx context.Context, ar cosmosaccount.Registry, source SourceOption, options ...Option) (*Chain, error) {
	c := &Chain{
		ar: ar,
	}
	source(c)
	for _, apply := range options {
		apply(c)
	}

	c.ev.Send(events.New(events.StatusOngoing, "獲取源代碼"))

	var err error
	if c.path, c.hash, err = fetchSource(ctx, c.url, c.ref, c.hash); err != nil {
		return nil, err
	}

	c.ev.Send(events.New(events.StatusDone, "已獲取源代碼"))
	c.ev.Send(events.New(events.StatusOngoing, "設置區塊鏈"))

	chainOption := []chain.Option{
		chain.ID(c.id),
		chain.HomePath(c.home),
		chain.LogLevel(chain.LogSilent),
	}

	// 在 Gitpod 上使用測試密鑰環後端，以防止提示輸入密鑰環密碼。這是因為 Gitpod 使用容器。

	if gitpod.IsOnGitpod() {
		c.keyringBackend = chaincmd.KeyringBackendTest
	}

	chainOption = append(chainOption, chain.KeyringBackend(c.keyringBackend))

	chain, err := chain.New(c.path, chainOption...)
	if err != nil {
		return nil, err
	}

	if !chain.Version.IsFamily(cosmosver.Stargate) {
		return nil, sperrors.ErrOnlyStargateSupported
	}

	c.chain = chain
	c.ev.Send(events.New(events.StatusDone, "區塊鏈設置"))

	return c, nil
}

func (c Chain) ChainID() (string, error) {
	return c.chain.ChainID()
}

func (c Chain) ID() (string, error) {
	return c.chain.ID()
}

func (c Chain) Name() string {
	return c.chain.Name()
}

func (c Chain) SetHome(home string) {
	c.chain.SetHome(home)
}

func (c Chain) Home() (path string, err error) {
	return c.chain.Home()
}

func (c Chain) BinaryName() (name string, err error) {
	return c.chain.Binary()
}

func (c Chain) GenesisPath() (path string, err error) {
	return c.chain.GenesisPath()
}

func (c Chain) GentxsPath() (path string, err error) {
	return c.chain.GentxsPath()
}

func (c Chain) DefaultGentxPath() (path string, err error) {
	return c.chain.DefaultGentxPath()
}

func (c Chain) AppTOMLPath() (string, error) {
	return c.chain.AppTOMLPath()
}

func (c Chain) ConfigTOMLPath() (string, error) {
	return c.chain.ConfigTOMLPath()
}

func (c Chain) SourceURL() string {
	return c.url
}

func (c Chain) SourceHash() string {
	return c.hash
}

func (c Chain) IsHomeDirExist() (ok bool, err error) {
	home, err := c.chain.Home()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(home)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// NodeID returns the chain node id
func (c Chain) NodeID(ctx context.Context) (string, error) {
	chainCmd, err := c.chain.Commands(ctx)
	if err != nil {
		return "", err
	}

	nodeID, err := chainCmd.ShowNodeID(ctx)
	if err != nil {
		return "", err
	}
	return nodeID, nil
}

// Build 構建鏈源，還檢查源是否已經構建
func (c *Chain) Build(ctx context.Context, cacheStorage cache.Storage) (binaryName string, err error) {
	// 如果鏈已經發布並且有啟動 ID 檢查二進制緩存
	if c.launchID != 0 {
		if binaryName, err = c.chain.Binary(); err != nil {
			return "", err
		}
		binaryChecksum, err := checksum.Binary(binaryName)
		if err != nil && !errors.Is(err, exec.ErrNotFound) {
			return "", err
		}
		binaryMatch, err := checkBinaryCacheForLaunchID(c.launchID, binaryChecksum, c.hash)
		if err != nil {
			return "", err
		}
		if binaryMatch {
			return binaryName, nil
		}
	}

	c.ev.Send(events.New(events.StatusOngoing, "構建鏈的二進製文件"))

	// 構建二進制
	if binaryName, err = c.chain.Build(ctx, cacheStorage, ""); err != nil {
		return "", err
	}

	c.ev.Send(events.New(events.StatusDone, "鏈的二進制構建"))

	// 為啟動 ID 緩存構建的二進製文件
	if c.launchID != 0 {
		if err := c.CacheBinary(c.launchID); err != nil {
			return "", nil
		}
	}

	return binaryName, nil
}

// CacheBinary 緩存與啟動 ID 關聯的最後構建的鏈二進製文件
func (c *Chain) CacheBinary(launchID uint64) error {
	binaryName, err := c.chain.Binary()
	if err != nil {
		return err
	}
	binaryChecksum, err := checksum.Binary(binaryName)

	if err != nil {
		return err
	}
	return cacheBinaryForLaunchID(launchID, binaryChecksum, c.hash)
}

// fetchSource從 url 獲取鏈源並返回保存源的臨時路徑
func fetchSource(
	ctx context.Context,
	url string,
	ref plumbing.ReferenceName,
	customHash string,
) (path, hash string, err error) {
	var repo *git.Repository

	if path, err = os.MkdirTemp("", ""); err != nil {
		return "", "", err
	}

	// ensure鏈源路徑存在
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", "", err
	}

	// prepare克隆選項。
	gitoptions := &git.CloneOptions{
		URL: url,
	}

	// clone指定時的 ref，由鏈協調器在創建時使用。
	if ref != "" {
		gitoptions.ReferenceName = ref
		gitoptions.SingleBranch = true
	}
	if repo, err = git.PlainCloneContext(ctx, path, false, gitoptions); err != nil {
		return "", "", err
	}

	if customHash != "" {
		hash = customHash

		// 指定時結帳到某個哈希值。驗證器使用它來確保使用
		// 區塊鏈的鎖定版本。
		wt, err := repo.Worktree()
		if err != nil {
			return "", "", err
		}
		h, err := repo.ResolveRevision(plumbing.Revision(customHash))
		if err != nil {
			return "", "", err
		}
		githash := *h
		if err := wt.Checkout(&git.CheckoutOptions{
			Hash: githash,
		}); err != nil {
			return "", "", err
		}
	} else {
		// 當沒有提供特定的哈希值時。獲取 HEAD
		ref, err := repo.Head()
		if err != nil {
			return "", "", err
		}
		hash = ref.Hash().String()
	}

	return path, hash, nil
}
