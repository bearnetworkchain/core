package chain

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/ignite-hq/cli/ignite/chainconfig"
	"github.com/ignite-hq/cli/ignite/pkg/cache"
	chaincmdrunner "github.com/ignite-hq/cli/ignite/pkg/chaincmd/runner"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosfaucet"
	"github.com/ignite-hq/cli/ignite/pkg/dirchange"
	"github.com/ignite-hq/cli/ignite/pkg/localfs"
	"github.com/ignite-hq/cli/ignite/pkg/xexec"
	"github.com/ignite-hq/cli/ignite/pkg/xfilepath"
	"github.com/ignite-hq/cli/ignite/pkg/xhttp"
	"github.com/ignite-hq/cli/ignite/pkg/xurl"
)

const (
	// 導出創世紀是鏈的導出創世文件的名稱
	exportedGenesis = "exported_genesis.json"

	// sourceChecksumKey 是校驗和檢測源修改的緩存鍵
	sourceChecksumKey = "source_checksum"

	// binaryChecksumKey 是校驗和檢測二進制修改的緩存鍵
	binaryChecksumKey = "binary_checksum"

	// configChecksumKey 是包含校驗和以檢測配置修改的緩存鍵
	configChecksumKey = "config_checksum"

	// serveDirchangeCacheNamespace 是緩存命名空間的名稱，用於檢測目錄的變化
	serveDirchangeCacheNamespace = "serve.dirchange"
)

var (
	// ignoreExts 保存了一個被忽略的文件列表。
	ignoredExts = []string{"pb.go", "pb.gw.go"}

	// starportSavePath 是保存鏈導出創世的地方
	starportSavePath = xfilepath.Join(
		chainconfig.ConfigDirPath,
		xfilepath.Path("local-chains"),
	)
)

type serveOptions struct {
	forceReset bool
	resetOnce  bool
}

func newServeOption() serveOptions {
	return serveOptions{
		forceReset: false,
		resetOnce:  false,
	}
}

// ServeOption 為 serve 命令提供選項
type ServeOption func(*serveOptions)

// ServeForceReset 允許在服務鏈時以及每次源更改時強制重置狀態
func ServeForceReset() ServeOption {
	return func(c *serveOptions) {
		c.forceReset = true
	}
}

// ServeResetOnce 允許在鏈服務一次時重置狀態
func ServeResetOnce() ServeOption {
	return func(c *serveOptions) {
		c.resetOnce = true
	}
}

// 服務提供應用程序。
func (c *Chain) Serve(ctx context.Context, cacheStorage cache.Storage, options ...ServeOption) error {
	serveOptions := newServeOption()

	// 應用選項
	for _, apply := range options {
		apply(&serveOptions)
	}

	// 初始檢查和設置。
	if err := c.setup(); err != nil {
		return err
	}

	// 確保 config.yml 存在
	if c.options.ConfigFile != "" {
		if _, err := os.Stat(c.options.ConfigFile); err != nil {
			return err
		}
	} else if _, err := chainconfig.LocateDefault(c.app.Path); err != nil {
		return err
	}

	// 開始服務組件。
	g, ctx := errgroup.WithContext(ctx)

	//區塊鏈節點例程
	g.Go(func() error {
		c.refreshServe()

		for {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			select {
			case <-ctx.Done():
				return ctx.Err()

			case <-c.serveRefresher:
				commands, err := c.Commands(ctx)
				if err != nil {
					return err
				}

				var (
					serveCtx context.Context
					buildErr *CannotBuildAppError
					startErr *CannotStartAppError
				)
				serveCtx, c.serveCancel = context.WithCancel(ctx)

				// 確定鍊是否應該重置狀態
				shouldReset := serveOptions.forceReset || serveOptions.resetOnce

				// 為應用程序服務。
				err = c.serve(serveCtx, cacheStorage, shouldReset)
				serveOptions.resetOnce = false

				switch {
				case err == nil:
				case errors.Is(err, context.Canceled):
					//如果應用程序已經被服務，我們保存創世狀態
					if c.served {
						c.served = false

						fmt.Fprintln(c.stdLog().out, "💿 保存熊網鏈創世狀態...")

						// 如果服務已停止，則保存創世狀態
						if err := c.saveChainState(context.TODO(), commands); err != nil {
							fmt.Fprint(c.stdLog().err, err.Error())
							return err
						}

						genesisPath, err := c.exportedGenesisPath()
						if err != nil {
							fmt.Fprintln(c.stdLog().err, err.Error())
							return err
						}
						fmt.Fprintf(c.stdLog().out, "💿 熊網鏈創世狀態保存在 %s\n", genesisPath)
					}
				case errors.As(err, &buildErr):
					fmt.Fprintf(c.stdLog().err, "%s\n", errorColor(err.Error()))

					var validationErr *chainconfig.ValidationError
					if errors.As(err, &validationErr) {
						fmt.Fprintln(c.stdLog().out, "請查看: https://github.com/ignite-hq/cli#configure")
					}

					fmt.Fprintf(c.stdLog().out, "%s\n", infoColor("在重試之前等待修復..."))

				case errors.As(err, &startErr):
					// 解析返回的錯誤日誌
					parsedErr := startErr.ParseStartError()

					// 如果為空，我們無法識別錯誤
					// 因此，該錯誤可能是由於與舊應用狀態不兼容的新邏輯引起的
					// 我們建議用戶最終重置應用狀態
					if parsedErr == "" {
						fmt.Fprintf(c.stdLog().out, "%s %s\n", infoColor(`區塊鏈無法啟動。如果新代碼不再與保存的狀態兼容，您可以通過啟動來重置數據庫:`), "ignite chain serve --reset-once")

						return fmt.Errorf("不能啟動 %s", startErr.AppName)
					}

					// 返回明確的解析錯誤
					return errors.New(parsedErr)
				default:
					return err
				}
			}
		}
	})

	// 日常看後端
	g.Go(func() error {
		return c.watchAppBackend(ctx)
	})

	return g.Wait()
}

func (c *Chain) setup() error {
	fmt.Fprintf(c.stdLog().out, "熊網鏈版本是: %s\n\n", infoColor(c.Version))

	return c.checkSystem()
}

// checkSystem 檢查開發人員的工作環境是否符合必須有
// 依賴關係和前提條件。
func (c *Chain) checkSystem() error {
	// 檢查 Go 是否已安裝。
	if !xexec.IsCommandAvailable("go") {
		return errors.New("請檢查是否正確安裝了Go語言 $PATH. See https://golang.org/doc/install")
	}
	return nil
}

func (c *Chain) refreshServe() {
	if c.serveCancel != nil {
		c.serveCancel()
	}
	c.serveRefresher <- struct{}{}
}

func (c *Chain) watchAppBackend(ctx context.Context) error {
	watchPaths := appBackendSourceWatchPaths
	if c.ConfigPath() != "" {
		watchPaths = append(watchPaths, c.ConfigPath())
	}

	return localfs.Watch(
		ctx,
		watchPaths,
		localfs.WatcherWorkdir(c.app.Path),
		localfs.WatcherOnChange(c.refreshServe),
		localfs.WatcherIgnoreHidden(),
		localfs.WatcherIgnoreFolders(),
		localfs.WatcherIgnoreExt(ignoredExts...),
	)
}

// serve 執行為區塊鏈服務的操作：構建、初始化和啟動
// 如果鏈已經初始化並且文件沒有改變，則直接啟動應用程序
// 如果文件改變了，狀態被導入
func (c *Chain) serve(ctx context.Context, cacheStorage cache.Storage, forceReset bool) error {
	conf, err := c.Config()
	if err != nil {
		return &CannotBuildAppError{err}
	}

	commands, err := c.Commands(ctx)
	if err != nil {
		return err
	}

	// isInit 判斷應用是否初始化
	var isInit bool

	dirCache := cache.New[[]byte](cacheStorage, serveDirchangeCacheNamespace)

// 確定應用程序是否必須重置狀態
// 如果必須重置狀態，那麼我們認為鏈沒有初始化
	isInit, err = c.IsInitialized()
	if err != nil {
		return err
	}
	if isInit {
		configModified := false
		if c.ConfigPath() != "" {
			configModified, err = dirchange.HasDirChecksumChanged(dirCache, configChecksumKey, c.app.Path, c.ConfigPath())
			if err != nil {
				return err
			}
		}

		if forceReset || configModified {
			// 如果設置了 forceReset，我們認為應用程序沒有初始化
			fmt.Fprintln(c.stdLog().out, "🔄 重置熊網鏈應用狀態...")
			isInit = false
		}
	}

	// 檢查自上次服務以來是否已修改源
	// 如果狀態不能被重置但源已經改變，我們重建鏈並導入導出的狀態
	sourceModified, err := dirchange.HasDirChecksumChanged(dirCache, sourceChecksumKey, c.app.Path, appBackendSourceWatchPaths...)
	if err != nil {
		return err
	}

	// 我們還考慮了校驗和中的二進製文件，以確保二進製文件未被第三方更改
	var binaryModified bool
	binaryName, err := c.Binary()
	if err != nil {
		return err
	}
	binaryPath, err := exec.LookPath(binaryName)
	if err != nil {
		if !errors.Is(err, exec.ErrNotFound) {
			return err
		}
		binaryModified = true
	} else {
		binaryModified, err = dirchange.HasDirChecksumChanged(dirCache, binaryChecksumKey, "", binaryPath)
		if err != nil {
			return err
		}
	}

	appModified := sourceModified || binaryModified

	// 檢查導出的創世紀是否存在
	exportGenesisExists := true
	exportedGenesisPath, err := c.exportedGenesisPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(exportedGenesisPath); os.IsNotExist(err) {
		exportGenesisExists = false
	} else if err != nil {
		return err
	}

	// 構建階段
	if !isInit || appModified {
		// 構建區塊鏈應用程序
		if err := c.build(ctx, cacheStorage, ""); err != nil {
			return err
		}
	}

	// 初始階段
	// 不願意：gocritic
	if !isInit || (appModified && !exportGenesisExists) {
		fmt.Fprintln(c.stdLog().out, "💿 初始化熊網鏈應用程序...")

		if err := c.Init(ctx, true); err != nil {
			return err
		}
	} else if appModified {
		// 如果鏈已經初始化但源已被修改
		// 我們重置鏈數據庫並導入創世狀態
		fmt.Fprintln(c.stdLog().out, "💿 檢測到存在的創世起源，正在恢復數據庫...")

		if err := commands.UnsafeReset(ctx); err != nil {
			return err
		}

		if err := c.importChainState(); err != nil {
			return err
		}
	} else {
		fmt.Fprintln(c.stdLog().out, "▶️  重啟熊網鏈現有應用...")
	}

	// 保存校驗和
	if c.ConfigPath() != "" {
		if err := dirchange.SaveDirChecksum(dirCache, configChecksumKey, c.app.Path, c.ConfigPath()); err != nil {
			return err
		}
	}
	if err := dirchange.SaveDirChecksum(dirCache, sourceChecksumKey, c.app.Path, appBackendSourceWatchPaths...); err != nil {
		return err
	}
	binaryPath, err = exec.LookPath(binaryName)
	if err != nil {
		return err
	}
	if err := dirchange.SaveDirChecksum(dirCache, binaryChecksumKey, "", binaryPath); err != nil {
		return err
	}

	// 啟動區塊鏈
	return c.start(ctx, conf)
}

func (c *Chain) start(ctx context.Context, config chainconfig.Config) error {
	commands, err := c.Commands(ctx)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	// 啟動區塊鏈。
	g.Go(func() error { return c.plugin.Start(ctx, commands, config) })

	// 如果啟用，請啟動水龍頭。
	faucet, err := c.Faucet(ctx)
	isFaucetEnabled := err != ErrFaucetIsNotEnabled

	if isFaucetEnabled {
		if err == ErrFaucetAccountDoesNotExist {
			return &CannotBuildAppError{errors.Wrap(err, "水龍頭帳戶不存在")}
		}
		if err != nil {
			return err
		}

		g.Go(func() (err error) {
			if err := c.runFaucetServer(ctx, faucet); err != nil {
				return &CannotBuildAppError{err}
			}
			return nil
		})
	}

	// 將應用設置為正在服務
	c.served = true

	// 注意：地址格式錯誤由錯誤組，因此可以在這里安全地忽略它們

	rpcAddr, _ := xurl.HTTP(config.Host.RPC)
	apiAddr, _ := xurl.HTTP(config.Host.API)

	// 打印服務器地址。
	fmt.Fprintf(c.stdLog().out, "🌍 熊網鏈節點: %s\n", rpcAddr)
	fmt.Fprintf(c.stdLog().out, "🌍 熊網鏈API: %s\n", apiAddr)

	if isFaucetEnabled {
		faucetAddr, _ := xurl.HTTP(chainconfig.FaucetHost(config))
		fmt.Fprintf(c.stdLog().out, "🌍 熊幣水龍頭: %s\n", faucetAddr)
	}

	return g.Wait()
}

func (c *Chain) runFaucetServer(ctx context.Context, faucet cosmosfaucet.Faucet) error {
	config, err := c.Config()
	if err != nil {
		return err
	}

	return xhttp.Serve(ctx, &http.Server{
		Addr:    chainconfig.FaucetHost(config),
		Handler: faucet,
	})
}

// saveChainState 運行鏈的導出命令並將導出的創世紀存儲在鏈保存的配置中
func (c *Chain) saveChainState(ctx context.Context, commands chaincmdrunner.Runner) error {
	genesisPath, err := c.exportedGenesisPath()
	if err != nil {
		return err
	}

	return commands.Export(ctx, genesisPath)
}

// importChainState 導入鏈配置中保存的創世紀以將其用作創世紀
func (c *Chain) importChainState() error {
	exportGenesisPath, err := c.exportedGenesisPath()
	if err != nil {
		return err
	}
	genesisPath, err := c.GenesisPath()
	if err != nil {
		return err
	}

	return copy.Copy(exportGenesisPath, genesisPath)
}

// chainSavePath 返回鏈狀態保存的路徑
// 如果路徑不存在則創建路徑
func (c *Chain) chainSavePath() (string, error) {
	savePath, err := starportSavePath()
	if err != nil {
		return "", err
	}

	chainID, err := c.ID()
	if err != nil {
		return "", err
	}
	chainSavePath := filepath.Join(savePath, chainID)

	// 確保路徑存在
	if err := os.MkdirAll(savePath, 0700); err != nil && !os.IsExist(err) {
		return "", err
	}

	return chainSavePath, nil
}

// exportGenesisPath 返回導出的創世文件的路徑
func (c *Chain) exportedGenesisPath() (string, error) {
	savePath, err := c.chainSavePath()
	if err != nil {
		return "", err
	}

	return filepath.Join(savePath, exportedGenesis), nil
}

type CannotBuildAppError struct {
	Err error
}

func (e *CannotBuildAppError) Error() string {
	return fmt.Sprintf("無法構建熊網鏈應用程序:\n\n\t%s", e.Err)
}

func (e *CannotBuildAppError) Unwrap() error {
	return e.Err
}

type CannotStartAppError struct {
	AppName string
	Err     error
}

func (e *CannotStartAppError) Error() string {
	return fmt.Sprintf("不能啟動 %sd 開始:\n%s", e.AppName, errors.Unwrap(e.Err))
}

func (e *CannotStartAppError) Unwrap() error {
	return e.Err
}

// ParseStartError 將錯誤解析為明確的錯誤字符串
// Cosmos SDK 應用程序的錯誤日誌太長，無法直接打印
// 如果錯誤沒有被識別，返回一個空字符串
func (e *CannotStartAppError) ParseStartError() string {
	errorLogs := errors.Unwrap(e.Err).Error()
	switch {
	case strings.Contains(errorLogs, "綁定：地址已經在使用中"):
		r := regexp.MustCompile(`listen .* 綁定：地址已經在使用中`)
		return r.FindString(errorLogs)
	case strings.Contains(errorLogs, "驗證器集在創世中為零"):
		return "錯誤：握手期間出錯：重放時出錯：驗證器集在創世中為零，並且在 InitChain 之後仍然為空"
	default:
		return ""
	}
}
