package envtest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ignite-hq/cli/ignite/chainconfig"
	"github.com/ignite-hq/cli/ignite/pkg/availableport"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosfaucet"
	"github.com/ignite-hq/cli/ignite/pkg/gocmd"
	"github.com/ignite-hq/cli/ignite/pkg/httpstatuschecker"
	"github.com/ignite-hq/cli/ignite/pkg/xexec"
	"github.com/ignite-hq/cli/ignite/pkg/xurl"
)

const (
	ServeTimeout = time.Minute * 15
	IgniteApp    = "ignite"
	ConfigYML    = "config.yml"
)

var isCI, _ = strconv.ParseBool(os.Getenv("CI"))

// Env 提供了一個隔離的測試環境以及需要什麼
// 使它成為可能。
type Env struct {
	t   *testing.T
	ctx context.Context
}

//new 創建一個新的測試環境。
func New(t *testing.T) Env {
	ctx, cancel := context.WithCancel(context.Background())
	e := Env{
		t:   t,
		ctx: ctx,
	}
	t.Cleanup(cancel)

	if !xexec.IsCommandAvailable(IgniteApp) {
		t.Fatal("Ignite需要安裝")
	}

	return e
}

// SetCleanup 註冊一個函數，當測試（或子測試）及其所有
// 子測試完成。
func (e Env) SetCleanup(f func()) {
	e.t.Cleanup(f)
}

// Ctx 返回測試套件的父上下文以用於取消。
func (e Env) Ctx() context.Context {
	return e.ctx
}

type execOptions struct {
	ctx                    context.Context
	shouldErr, shouldRetry bool
	stdout, stderr         io.Writer
}

type ExecOption func(*execOptions)

// ExecShouldError 將命令執行的期望設置為以失敗結束。
func ExecShouldError() ExecOption {
	return func(o *execOptions) {
		o.shouldErr = true
	}
}

// ExecCtx 設置執行的取消上下文。
func ExecCtx(ctx context.Context) ExecOption {
	return func(o *execOptions) {
		o.ctx = ctx
	}
}

// ExecStdout 捕獲執行的標準輸出。
func ExecStdout(w io.Writer) ExecOption {
	return func(o *execOptions) {
		o.stdout = w
	}
}

// ExecStderr 捕獲執行的標準錯誤。
func ExecStderr(w io.Writer) ExecOption {
	return func(o *execOptions) {
		o.stderr = w
	}
}

// ExecRetry 在取消上下文之前重試命令直到成功。
func ExecRetry() ExecOption {
	return func(o *execOptions) {
		o.shouldRetry = true
	}
}

// Exec 執行帶有選項的命令步驟，其中 msg 描述了測試的期望。
// 除非使用 Must() 調用，否則 Exec() 不會在失敗時退出測試運行時。
func (e Env) Exec(msg string, steps step.Steps, options ...ExecOption) (ok bool) {
	opts := &execOptions{
		ctx:    e.ctx,
		stdout: io.Discard,
		stderr: io.Discard,
	}
	for _, o := range options {
		o(opts)
	}
	var (
		stdout = &bytes.Buffer{}
		stderr = &bytes.Buffer{}
	)
	copts := []cmdrunner.Option{
		cmdrunner.DefaultStdout(io.MultiWriter(stdout, opts.stdout)),
		cmdrunner.DefaultStderr(io.MultiWriter(stderr, opts.stderr)),
	}
	if isCI {
		copts = append(copts, cmdrunner.EndSignal(os.Kill))
	}
	err := cmdrunner.
		New(copts...).
		Run(opts.ctx, steps...)
	if err == context.Canceled {
		err = nil
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if opts.shouldRetry && opts.ctx.Err() == nil {
			time.Sleep(time.Second)
			return e.Exec(msg, steps, options...)
		}
	}
	if err != nil {
		msg = fmt.Sprintf("%s\n\nLogs:\n\n%s\n\nError Logs:\n\n%s\n",
			msg,
			stdout.String(),
			stderr.String())
	}
	if opts.shouldErr {
		return assert.Error(e.t, err, msg)
	}
	return assert.NoError(e.t, err, msg)
}

const (
	Stargate = "stargate"
)

// Scaffold 將一個應用程序腳手架到一個唯一的 appPath 並返回它。
func (e Env) Scaffold(name string, flags ...string) (appPath string) {
	root := e.TmpDir()
	e.Exec("scaffold an app",
		step.NewSteps(step.New(
			step.Exec(
				IgniteApp,
				append([]string{
					"scaffold",
					"chain",
					name,
				}, flags...)...,
			),
			step.Workdir(root),
		)),
	)

	appDir := path.Base(name)

	// 清理應用的主目錄和緩存
	e.t.Cleanup(func() {
		os.RemoveAll(filepath.Join(e.Home(), fmt.Sprintf(".%s", appDir)))
	})

	return filepath.Join(root, appDir)
}

// Serve 服務於路徑下的應用程序，其中帶有 msg 描述的選項
// 從服務動作執行。
// 除非使用 Must() 調用，否則 Serve() 不會在失敗時退出測試運行時。
func (e Env) Serve(msg, path, home, configPath string, options ...ExecOption) (ok bool) {
	serveCommand := []string{
		"chain",
		"serve",
		"-v",
	}

	if home != "" {
		serveCommand = append(serveCommand, "--home", home)
	}
	if configPath != "" {
		serveCommand = append(serveCommand, "--config", configPath)
	}

	return e.Exec(msg,
		step.NewSteps(step.New(
			step.Exec(IgniteApp, serveCommand...),
			step.Workdir(path),
		)),
		options...,
	)
}

// Simulate 為應用程序運行模擬測試
func (e Env) Simulate(appPath string, numBlocks, blockSize int) {
	e.Exec("running the simulation tests",
		step.NewSteps(step.New(
			step.Exec(
				IgniteApp,
				"chain",
				"simulate",
				"--numBlocks",
				strconv.Itoa(numBlocks),
				"--blockSize",
				strconv.Itoa(blockSize),
			),
			step.Workdir(appPath),
		)),
	)
}

// EnsureAppIsSteady 確保位於該路徑的應用程序可以編譯及其測試
// 正在通過。
func (e Env) EnsureAppIsSteady(appPath string) {
	_, statErr := os.Stat(filepath.Join(appPath, ConfigYML))
	require.False(e.t, os.IsNotExist(statErr), "config.yml找不到")

	e.Exec("確保應用程序穩定",
		step.NewSteps(step.New(
			step.Exec(gocmd.Name(), "test", "./..."),
			step.Workdir(appPath),
		)),
	)
}

// IsAppServed 檢查應用程序是否正確服務並且服務器是否開始監聽
// 在 ctx 取消之前。
func (e Env) IsAppServed(ctx context.Context, host chainconfig.Host) error {
	checkAlive := func() error {
		addr, err := xurl.HTTP(host.API)
		if err != nil {
			return err
		}

		ok, err := httpstatuschecker.Check(ctx, fmt.Sprintf("%s/node_info", addr))
		if err == nil && !ok {
			err = errors.New("應用不在線")
		}
		return err
	}
	return backoff.Retry(checkAlive, backoff.WithContext(backoff.NewConstantBackOff(time.Second), ctx))
}

// IsFaucetServed 檢查應用程序的水龍頭是否正確提供
func (e Env) IsFaucetServed(ctx context.Context, faucetClient cosmosfaucet.HTTPClient) error {
	checkAlive := func() error {
		_, err := faucetClient.FaucetInfo(ctx)
		return err
	}
	return backoff.Retry(checkAlive, backoff.WithContext(backoff.NewConstantBackOff(time.Second), ctx))
}

// TmpDir創建一個新的臨時目錄。
func (e Env) TmpDir() (path string) {
	path, err := os.MkdirTemp("", "integration")
	require.NoError(e.t, err, "create a tmp dir")
	e.t.Cleanup(func() { os.RemoveAll(path) })
	return path
}

// RandomizeServerPorts 在路徑中隨機化應用程序的服務器端口，更新
// 它的 config.yml 並返回新值。
func (e Env) RandomizeServerPorts(path string, configFile string) chainconfig.Host {
	if configFile == "" {
		configFile = ConfigYML
	}

	// 生成隨機服務器端口和服務器列表。
	ports, err := availableport.Find(6)
	require.NoError(e.t, err)

	genAddr := func(port int) string {
		return fmt.Sprintf("localhost:%d", port)
	}

	servers := chainconfig.Host{
		RPC:     genAddr(ports[0]),
		P2P:     genAddr(ports[1]),
		Prof:    genAddr(ports[2]),
		GRPC:    genAddr(ports[3]),
		GRPCWeb: genAddr(ports[4]),
		API:     genAddr(ports[5]),
	}

	// 使用生成的服務器列表更新 config.yml。
	configyml, err := os.OpenFile(filepath.Join(path, configFile), os.O_RDWR|os.O_CREATE, 0755)
	require.NoError(e.t, err)
	defer configyml.Close()

	var conf chainconfig.Config
	require.NoError(e.t, yaml.NewDecoder(configyml).Decode(&conf))

	conf.Host = servers
	require.NoError(e.t, configyml.Truncate(0))
	_, err = configyml.Seek(0, 0)
	require.NoError(e.t, err)
	require.NoError(e.t, yaml.NewEncoder(configyml).Encode(conf))

	return servers
}

//ConfigureFaucet 為應用程序水龍頭找到一個隨機端口並使用此端口更新 config.yml 並提供硬幣選項
func (e Env) ConfigureFaucet(path string, configFile string, coins, coinsMax []string) string {
	if configFile == "" {
		configFile = ConfigYML
	}

	// 找到一個隨機的可用端口
	port, err := availableport.Find(1)
	require.NoError(e.t, err)

	configyml, err := os.OpenFile(filepath.Join(path, configFile), os.O_RDWR|os.O_CREATE, 0755)
	require.NoError(e.t, err)
	defer configyml.Close()

	var conf chainconfig.Config
	require.NoError(e.t, yaml.NewDecoder(configyml).Decode(&conf))

	conf.Faucet.Port = port[0]
	conf.Faucet.Coins = coins
	conf.Faucet.CoinsMax = coinsMax
	require.NoError(e.t, configyml.Truncate(0))
	_, err = configyml.Seek(0, 0)
	require.NoError(e.t, err)
	require.NoError(e.t, yaml.NewEncoder(configyml).Encode(conf))

	addr, err := xurl.HTTP(fmt.Sprintf("0.0.0.0:%d", port[0]))
	require.NoError(e.t, err)

	return addr
}

// SetRandomHomeConfig 在區塊鏈配置文件中設置為主目錄生成臨時目錄
func (e Env) SetRandomHomeConfig(path string, configFile string) {
	if configFile == "" {
		configFile = ConfigYML
	}

	// 使用生成的臨時目錄更新 config.yml
	configyml, err := os.OpenFile(filepath.Join(path, configFile), os.O_RDWR|os.O_CREATE, 0755)
	require.NoError(e.t, err)
	defer configyml.Close()

	var conf chainconfig.Config
	require.NoError(e.t, yaml.NewDecoder(configyml).Decode(&conf))

	conf.Init.Home = e.TmpDir()
	require.NoError(e.t, configyml.Truncate(0))
	_, err = configyml.Seek(0, 0)
	require.NoError(e.t, err)
	require.NoError(e.t, yaml.NewEncoder(configyml).Encode(conf))
}

// 如果不正常，必須立即失敗。
// 在運行 Must() 之前，需要為失敗的測試調用 t.Fail()。
func (e Env) Must(ok bool) {
	if !ok {
		e.t.FailNow()
	}
}

// Home 返回用戶的主目錄。
func (e Env) Home() string {
	home, err := os.UserHomeDir()
	require.NoError(e.t, err)
	return home
}

//AppdHome 返回 appd 的主目錄。
func (e Env) AppdHome(name string) string {
	return filepath.Join(e.Home(), fmt.Sprintf(".%s", name))
}
