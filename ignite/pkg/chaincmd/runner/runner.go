// chaincmdrunner 提供對區塊鏈命令的高級訪問。
package chaincmdrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	"github.com/ignite-hq/cli/ignite/pkg/chaincmd"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/lineprefixer"
	"github.com/ignite-hq/cli/ignite/pkg/truncatedbuffer"
)

// Runner 提供對區塊鏈命令的高級訪問。
type Runner struct {
	chainCmd                      chaincmd.ChainCmd
	stdout, stderr                io.Writer
	daemonLogPrefix, cliLogPrefix string
}

// 選項配置 Runner。
type Option func(r *Runner)

// 標準輸出為執行的命令設置標準輸出。
func Stdout(w io.Writer) Option {
	return func(runner *Runner) {
		runner.stdout = w
	}
}

// DaemonLogPrefix 是添加到應用程序的守護程序日誌的前綴。
func DaemonLogPrefix(prefix string) Option {
	return func(runner *Runner) {
		runner.daemonLogPrefix = prefix
	}
}

// CLILogPrefix 是添加到應用程序 cli 日誌的前綴。
func CLILogPrefix(prefix string) Option {
	return func(runner *Runner) {
		runner.cliLogPrefix = prefix
	}
}

// Stderr 為執行的命令設置標準錯誤。
func Stderr(w io.Writer) Option {
	return func(runner *Runner) {
		runner.stderr = w
	}
}

// New 使用 cc 和 options 創建一個新的 Runner.
func New(ctx context.Context, chainCmd chaincmd.ChainCmd, options ...Option) (Runner, error) {
	runner := Runner{
		chainCmd: chainCmd,
		stdout:   io.Discard,
		stderr:   io.Discard,
	}

	applyOptions(&runner, options)

// 自動檢測鏈 id 並將其應用於 chaincmd 如果是 auto
// 啟用檢測。
	if chainCmd.IsAutoChainIDDetectionEnabled() {
		status, err := runner.Status(ctx)
		if err != nil {
			return Runner{}, err
		}

		runner.chainCmd = runner.chainCmd.Copy(chaincmd.WithChainID(status.ChainID))
	}

	return runner, nil
}

func applyOptions(r *Runner, options []Option) {
	for _, apply := range options {
		apply(r)
	}
}

// Copy 通過使用給定選項覆蓋其選項來複製 runner。
func (r Runner) Copy(options ...Option) Runner {
	applyOptions(&r, options)

	return r
}

// Cmd 返回底層鏈 cmd。
func (r Runner) Cmd() chaincmd.ChainCmd {
	return r.chainCmd
}

type runOptions struct {
// WrappedStdErrMaxLen 確定打包錯誤日誌的最大長度
// 此選項用於長時間運行的命令，以防止包含 stderr 的緩衝區變得太大
// 0 可以用於沒有最大長度
	wrappedStdErrMaxLen int

	// stdout 和 stderr 用於收集命令輸出的副本。
	stdout, stderr io.Writer

	// stdin 定義命令的輸入
	stdin io.Reader
}

// run 執行命令。
func (r Runner) run(ctx context.Context, runOptions runOptions, stepOptions ...step.Option) error {
	var (
// 我們使用截斷的緩衝區來防止內存洩漏
// 這是因為 Stargate 應用當前正在向 StdErr 發送日誌
// 因此，如果應用程序成功啟動，寫入的日誌會變得很長
		errb = truncatedbuffer.NewTruncatedBuffer(runOptions.wrappedStdErrMaxLen)

		// add optional prefixes to output streams.
		stdout io.Writer = lineprefixer.NewWriter(r.stdout,
			func() string { return r.daemonLogPrefix },
		)
		stderr io.Writer = lineprefixer.NewWriter(r.stderr,
			func() string { return r.cliLogPrefix },
		)
	)

	// 設置標準輸出
	if runOptions.stdout != nil {
		stdout = io.MultiWriter(stdout, runOptions.stdout)
	}
	if runOptions.stderr != nil {
		stderr = io.MultiWriter(stderr, runOptions.stderr)
	}

	stderr = io.MultiWriter(stderr, errb)

	runnerOptions := []cmdrunner.Option{
		cmdrunner.DefaultStdout(stdout),
		cmdrunner.DefaultStderr(stderr),
	}

	// 如果已定義，則設置標準輸入
	if runOptions.stdin != nil {
		runnerOptions = append(runnerOptions, cmdrunner.DefaultStdin(runOptions.stdin))
	}

	err := cmdrunner.
		New(runnerOptions...).
		Run(ctx, step.New(stepOptions...))

	return errors.Wrap(err, errb.GetBuffer().String())
}

func newBuffer() *buffer {
	return &buffer{
		Buffer: new(bytes.Buffer),
	}
}

//buffer 是一個帶有附加功能的 bytes.Buffer。
type buffer struct {
	*bytes.Buffer
}

// JSONEnsuredBytes 確保返回字節的編碼格式始終是
// JSON 即使寫入的數據最初是用 YAML 編碼的。
func (b *buffer) JSONEnsuredBytes() ([]byte, error) {
	bytes := b.Buffer.Bytes()

	var out interface{}

	if err := yaml.Unmarshal(bytes, &out); err == nil {
		return yaml.YAMLToJSON(bytes)
	}

	return bytes, nil
}

type txResult struct {
	Code   int    `json:"code"`
	RawLog string `json:"raw_log"`
	TxHash string `json:"txhash"`
}

func decodeTxResult(b *buffer) (txResult, error) {
	var r txResult

	data, err := b.JSONEnsuredBytes()
	if err != nil {
		return r, err
	}

	return r, json.Unmarshal(data, &r)
}
