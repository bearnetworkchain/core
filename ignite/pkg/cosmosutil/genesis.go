package cosmosutil

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

const (
	FieldGenesisTime                 = "genesis_time"
	FieldChainID                     = "chain_id"
	FieldConsumerChainID             = "app_state.monitoringp.params.consumerChainID"
	FieldLastBlockHeight             = "app_state.monitoringp.params.lastBlockHeight"
	FieldConsensusTimestamp          = "app_state.monitoringp.params.consumerConsensusState.timestamp"
	FieldConsensusNextValidatorsHash = "app_state.monitoringp.params.consumerConsensusState.nextValidatorsHash"
	FieldConsensusRootHash           = "app_state.monitoringp.params.consumerConsensusState.root.hash"
	FieldConsumerUnbondingPeriod     = "app_state.monitoringp.params.consumerUnbondingPeriod"
	FieldConsumerRevisionHeight      = "app_state.monitoringp.params.consumerRevisionHeight"
)

type (
	// Genesis 代表更易讀的熊網鏈創世紀文件版本
	Genesis struct {
		Accounts   []string
		StakeDenom string
	}
	// ChainGenesis 代表熊網鏈創世文件
	ChainGenesis struct {
		ChainID  string `json:"chain_id"`
		AppState struct {
			Auth struct {
				Accounts []struct {
					Address string `json:"address"`
				} `json:"accounts"`
			} `json:"auth"`
			Staking struct {
				Params struct {
					BondDenom string `json:"bond_denom"`
				} `json:"params"`
			} `json:"staking"`
		} `json:"app_state"`
	}

	// 從創世紀更新的字段
	fields map[string]string
	// GenesisField 配置創世鍵值字段。
	GenesisField func(fields)
)

// HasAccount 檢查帳戶是否存在於創世帳戶中
func (g Genesis) HasAccount(address string) bool {
	for _, account := range g.Accounts {
		if account == address {
			return true
		}
	}
	return false
}

// WithKeyValue 將鍵和值字段設置為創世文件
func WithKeyValue(key, value string) GenesisField {
	return func(f fields) {
		f[key] = value
	}
}

// WithKeyValueInt 將 key 和 int64 值字段設置為創世紀文件
func WithKeyValueInt(key string, value int64) GenesisField {
	return func(f fields) {
		f[key] = strconv.FormatInt(value, 10)
	}
}

// WithKeyValueUint 將鍵和 uint64 值字段設置為創世文件
func WithKeyValueUint(key string, value uint64) GenesisField {
	return func(f fields) {
		f[key] = strconv.FormatUint(value, 10)
	}
}

// WithKeyValueTimestamp 將鍵和時間戳值字段設置為創世文件
func WithKeyValueTimestamp(key string, value int64) GenesisField {
	return func(f fields) {
		f[key] = time.Unix(value, 0).UTC().Format(time.RFC3339Nano)
	}
}

// WithKeyValueBoolean 將鍵和布爾值字段設置為創世文件
func WithKeyValueBoolean(key string, value bool) GenesisField {
	return func(f fields) {
		if value {
			f[key] = "true"
		} else {
			f[key] = "false"
		}
	}
}

func UpdateGenesis(genesisPath string, options ...GenesisField) error {
	f := fields{}
	for _, applyField := range options {
		applyField(f)
	}

	genesisBytes, err := os.ReadFile(genesisPath)
	if err != nil {
		return err
	}

	for key, value := range f {
		genesisBytes, err = jsonparser.Set(
			genesisBytes,
			[]byte(fmt.Sprintf(`"%s"`, value)),
			strings.Split(key, ".")...,
		)
		if err != nil {
			return err
		}
	}
	return os.WriteFile(genesisPath, genesisBytes, 0644)
}

// ParseGenesisFromPath 從創世文件中解析鏈創世紀對象
func ParseGenesisFromPath(genesisPath string) (Genesis, error) {
	genesisFile, err := os.ReadFile(genesisPath)
	if err != nil {
		return Genesis{}, errors.Wrap(err, "無法打開創世文件")
	}
	return ParseGenesis(genesisFile)
}

// ParseChainGenesis 從字節片中解析鏈創世紀對象
func ParseChainGenesis(genesisFile []byte) (chainGenesis ChainGenesis, err error) {
	if err := json.Unmarshal(genesisFile, &chainGenesis); err != nil {
		return chainGenesis, errors.New("無法解組鏈創世文件: " + err.Error())
	}
	return chainGenesis, err
}

// ParseGenesis 將 ChainGenesis 對像從字節切片解析為創世紀對象
func ParseGenesis(genesisFile []byte) (Genesis, error) {
	chainGenesis, err := ParseChainGenesis(genesisFile)
	if err != nil {
		return Genesis{}, errors.New("無法解組創世文件: " + err.Error())
	}
	genesis := Genesis{StakeDenom: chainGenesis.AppState.Staking.Params.BondDenom}
	for _, acc := range chainGenesis.AppState.Auth.Accounts {
		genesis.Accounts = append(genesis.Accounts, acc.Address)
	}
	return genesis, nil
}

// CheckGenesisContainsAddress 如果地址存在於創世文件中，則返回 true
func CheckGenesisContainsAddress(genesisPath, addr string) (bool, error) {
	_, err := os.Stat(genesisPath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	genesis, err := ParseGenesisFromPath(genesisPath)
	if err != nil {
		return false, err
	}
	return genesis.HasAccount(addr), nil
}

// GenesisAndHashFromURL 從給定的 url 獲取創世紀並返回其內容以及 sha256 哈希值。
func GenesisAndHashFromURL(ctx context.Context, url string) (genesis []byte, hash string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	genesis, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	h := sha256.New()
	if _, err := io.Copy(h, bytes.NewReader(genesis)); err != nil {
		return nil, "", err
	}

	hexHash := hex.EncodeToString(h.Sum(nil))

	return genesis, hexHash, nil
}
