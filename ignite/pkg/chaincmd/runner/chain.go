package chaincmdrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"

	"github.com/ignite-hq/cli/ignite/pkg/chaincmd"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosver"
)

// Start 啟動區塊鏈。
func (r Runner) Start(ctx context.Context, args ...string) error {
	return r.run(
		ctx,
		runOptions{wrappedStdErrMaxLen: 50000},
		r.chainCmd.StartCommand(args...),
	)
}

// LaunchpadStartRestServer 重啟服務器。
func (r Runner) LaunchpadStartRestServer(ctx context.Context, apiAddress, rpcAddress string) error {
	return r.run(
		ctx,
		runOptions{wrappedStdErrMaxLen: 50000},
		r.chainCmd.LaunchpadRestServerCommand(apiAddress, rpcAddress),
	)
}

// Init 初始化區塊鏈。
func (r Runner) Init(ctx context.Context, moniker string) error {
	return r.run(ctx, runOptions{}, r.chainCmd.InitCommand(moniker))
}

// KV 持有一個鍵值對。
type KV struct {
	key   string
	value string
}

// NewKV 返回一個新的鍵值對。
func NewKV(key, value string) KV {
	return KV{key, value}
}

// LaunchpadSetConfigs 更新啟動板應用程序的配置。
func (r Runner) LaunchpadSetConfigs(ctx context.Context, kvs ...KV) error {
	for _, kv := range kvs {
		if err := r.run(
			ctx,
			runOptions{},
			r.chainCmd.LaunchpadSetConfigCommand(kv.key, kv.value),
		); err != nil {
			return err
		}
	}
	return nil
}

var gentxRe = regexp.MustCompile(`(?m)"(.+?)"`)

// Gentx 生成一個帶有自我委託的創世紀 tx。
func (r Runner) Gentx(
	ctx context.Context,
	validatorName,
	selfDelegation string,
	options ...chaincmd.GentxOption,
) (gentxPath string, err error) {
	b := &bytes.Buffer{}

	if err := r.run(ctx, runOptions{
		stdout: b,
		stderr: b,
		stdin:  os.Stdin,
	}, r.chainCmd.GentxCommand(validatorName, selfDelegation, options...)); err != nil {
		return "", err
	}

	return gentxRe.FindStringSubmatch(b.String())[1], nil
}

// CollectGentxs 收集gentxs。
func (r Runner) CollectGentxs(ctx context.Context) error {
	return r.run(ctx, runOptions{}, r.chainCmd.CollectGentxsCommand())
}

// ValidateGenesis 驗證創世紀.
func (r Runner) ValidateGenesis(ctx context.Context) error {
	return r.run(ctx, runOptions{}, r.chainCmd.ValidateGenesisCommand())
}

// UnsafeReset 重置區塊鏈數據庫.
func (r Runner) UnsafeReset(ctx context.Context) error {
	return r.run(ctx, runOptions{}, r.chainCmd.UnsafeResetCommand())
}

// ShowNodeID 顯示節點ID.
func (r Runner) ShowNodeID(ctx context.Context) (nodeID string, err error) {
	b := &bytes.Buffer{}
	err = r.run(ctx, runOptions{stdout: b}, r.chainCmd.ShowNodeIDCommand())
	nodeID = strings.TrimSpace(b.String())
	return
}

// NodeStatus 保存有關節點狀態的信息.
type NodeStatus struct {
	ChainID string
}

// 狀態返回節點的狀態.
func (r Runner) Status(ctx context.Context) (NodeStatus, error) {
	b := newBuffer()

	if err := r.run(ctx, runOptions{stdout: b, stderr: b}, r.chainCmd.StatusCommand()); err != nil {
		return NodeStatus{}, err
	}

	var chainID string

	data, err := b.JSONEnsuredBytes()
	if err != nil {
		return NodeStatus{}, err
	}

	version := r.chainCmd.SDKVersion()
	switch {
	case version.GTE(cosmosver.StargateFortyVersion):
		out := struct {
			NodeInfo struct {
				Network string `json:"network"`
			} `json:"NodeInfo"`
		}{}

		if err := json.Unmarshal(data, &out); err != nil {
			return NodeStatus{}, err
		}

		chainID = out.NodeInfo.Network
	default:
		out := struct {
			NodeInfo struct {
				Network string `json:"network"`
			} `json:"node_info"`
		}{}

		if err := json.Unmarshal(data, &out); err != nil {
			return NodeStatus{}, err
		}

		chainID = out.NodeInfo.Network
	}

	return NodeStatus{
		ChainID: chainID,
	}, nil
}

// 銀行發送金額從 from Account 發送到 to Account.
func (r Runner) BankSend(ctx context.Context, fromAccount, toAccount, amount string) (string, error) {
	b := newBuffer()
	opt := []step.Option{
		r.chainCmd.BankSendCommand(fromAccount, toAccount, amount),
	}

	if r.chainCmd.KeyringPassword() != "" {
		input := &bytes.Buffer{}
		fmt.Fprintln(input, r.chainCmd.KeyringPassword())
		fmt.Fprintln(input, r.chainCmd.KeyringPassword())
		fmt.Fprintln(input, r.chainCmd.KeyringPassword())
		opt = append(opt, step.Write(input.Bytes()))
	}

	if err := r.run(ctx, runOptions{stdout: b}, opt...); err != nil {
		if strings.Contains(err.Error(), "未找到密鑰") || // 星門
			strings.Contains(err.Error(), "未知地址") || // launchpad
			strings.Contains(b.String(), "找不到項目") { // launchpad
			return "", errors.New("帳戶沒有任何餘額")
		}

		return "", err
	}

	txResult, err := decodeTxResult(b)
	if err != nil {
		return "", err
	}

	if txResult.Code > 0 {
		return "", fmt.Errorf("無法發送令牌 (開發工具包代碼 %d): %s", txResult.Code, txResult.RawLog)
	}

	return txResult.TxHash, nil
}

// WaitTx 等到一個 tx 成功添加到一個塊並且可以被查詢
func (r Runner) WaitTx(ctx context.Context, txHash string, retryDelay time.Duration, maxRetry int) error {
	retry := 0

	// 重試查詢請求
	checkTx := func() error {
		b := newBuffer()
		if err := r.run(ctx, runOptions{stdout: b}, r.chainCmd.QueryTxCommand(txHash)); err != nil {
			// 過濾未找到錯誤並檢查最大重試次數
			if !strings.Contains(err.Error(), "未找到") {
				return backoff.Permanent(err)
			}
			retry++
			if retry == maxRetry {
				return backoff.Permanent(fmt.Errorf("無法檢索 tx %s", txHash))
			}
			return err
		}

		// parse tx and check code
		txResult, err := decodeTxResult(b)
		if err != nil {
			return backoff.Permanent(err)
		}
		if txResult.Code != 0 {
			return backoff.Permanent(fmt.Errorf("tx %s 失敗的: %s", txHash, txResult.RawLog))
		}

		return nil
	}
	return backoff.Retry(checkTx, backoff.WithContext(backoff.NewConstantBackOff(retryDelay), ctx))
}

// Export 將鏈的狀態導出到指定的文件中
func (r Runner) Export(ctx context.Context, exportedFile string) error {
	// 確保路徑存在
	dir := filepath.Dir(exportedFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	if err := r.run(ctx, runOptions{stdout: stdout, stderr: stderr}, r.chainCmd.ExportCommand()); err != nil {
		return err
	}

	// 導出的創世紀是從 Cosmos-SDK v0.44.0 寫在 stderr 上的
	var exportedState []byte
	if stdout.Len() > 0 {
		exportedState = stdout.Bytes()
	} else {
		exportedState = stderr.Bytes()
	}

	// 保存新狀態
	return os.WriteFile(exportedFile, exportedState, 0644)
}

// EventSelector 用於查詢事件。
type EventSelector struct {
	typ   string
	attr  string
	value string
}

// NewEventSelector 創建一個新的事件選擇器。
func NewEventSelector(typ, addr, value string) EventSelector {
	return EventSelector{typ, addr, value}
}

// Event 表示一個 TX 事件。
type Event struct {
	Type       string
	Attributes []EventAttribute
	Time       time.Time
}

// EventAttribute 保存事件的屬性。
type EventAttribute struct {
	Key   string
	Value string
}

//QueryTxEvents 通過事件選擇器查詢 tx 事件。
func (r Runner) QueryTxEvents(
	ctx context.Context,
	selector EventSelector,
	moreSelectors ...EventSelector,
) ([]Event, error) {
	// 準備選擇器。
	var list []string

	eventsSelectors := append([]EventSelector{selector}, moreSelectors...)

	for _, event := range eventsSelectors {
		list = append(list, fmt.Sprintf("%s.%s=%s", event.typ, event.attr, event.value))
	}

	query := strings.Join(list, "&")

	// 執行命令並解析輸出。
	b := newBuffer()

	if err := r.run(ctx, runOptions{stdout: b}, r.chainCmd.QueryTxEventsCommand(query)); err != nil {
		return nil, err
	}

	out := struct {
		Txs []struct {
			Logs []struct {
				Events []struct {
					Type  string `json:"type"`
					Attrs []struct {
						Key   string `json:"key"`
						Value string `json:"value"`
					} `json:"attributes"`
				} `json:"events"`
			} `json:"logs"`
			TimeStamp string `json:"timestamp"`
		} `json:"txs"`
	}{}

	data, err := b.JSONEnsuredBytes()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	var events []Event

	for _, tx := range out.Txs {
		for _, log := range tx.Logs {
			for _, e := range log.Events {
				var attrs []EventAttribute
				for _, attr := range e.Attrs {
					attrs = append(attrs, EventAttribute{
						Key:   attr.Key,
						Value: attr.Value,
					})
				}

				txTime, err := time.Parse(time.RFC3339, tx.TimeStamp)
				if err != nil {
					return nil, err
				}

				events = append(events, Event{
					Type:       e.Type,
					Attributes: attrs,
					Time:       txTime,
				})
			}
		}
	}

	return events, nil
}
