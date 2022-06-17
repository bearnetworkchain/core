package networkchain

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"

	"github.com/bearnetworkchain/core/ignite/pkg/availableport"
	"github.com/bearnetworkchain/core/ignite/pkg/cache"
	"github.com/bearnetworkchain/core/ignite/pkg/events"
	"github.com/bearnetworkchain/core/ignite/pkg/httpstatuschecker"
	"github.com/bearnetworkchain/core/ignite/pkg/xurl"
	"github.com/bearnetworkchain/core/ignite/services/network/networktypes"
)

const (
	ListeningTimeout            = time.Minute * 1
	ValidatorSetNilErrorMessage = "驗證器集在創世中為零，在 InitChain 之後仍然為空"
)

// SimulateRequests 根據提供的請求模擬創世創建和網絡啟動
func (c Chain) SimulateRequests(
	ctx context.Context,
	cacheStorage cache.Storage,
	gi networktypes.GenesisInformation,
	reqs []networktypes.Request,
) (err error) {
	c.ev.Send(events.New(events.StatusOngoing, "驗證請求格式"))
	for _, req := range reqs {
		//請求的靜態驗證
		if err := networktypes.VerifyRequest(req); err != nil {
			return err
		}

		//將請求應用於創世信息
		gi, err = gi.ApplyRequest(req)
		if err != nil {
			return err
		}
	}
	c.ev.Send(events.New(events.StatusDone, "請求格式已驗證"))

	// 用請求準備鏈
	if err := c.Prepare(
		ctx,
		cacheStorage,
		gi,
		networktypes.Reward{RevisionHeight: 1},
		networktypes.SPNChainID,
		1,
		2,
	); err != nil {
		return err
	}

	c.ev.Send(events.New(events.StatusOngoing, "嘗試使用請求啟動網絡"))
	if err := c.simulateChainStart(ctx); err != nil {
		return err
	}
	c.ev.Send(events.New(events.StatusDone, "網絡可以啟動"))

	return nil
}

// SimulateChainStart 通過使用模擬配置啟動它來模擬和驗證鏈開始
// 並檢查 gentxs 執行是否成功
func (c Chain) simulateChainStart(ctx context.Context) error {
	cmd, err := c.chain.Commands(ctx)
	if err != nil {
		return err
	}

	//使用隨機端口設置配置以測試啟動命令
	addressAPI, err := c.setSimulationConfig()
	if err != nil {
		return err
	}

	//驗證該鍊是否可以以有效的創世紀啟動
	ctx, cancel := context.WithTimeout(ctx, ListeningTimeout)
	exit := make(chan error)

	//檢查應用程序是否正在收聽的例行程序
	go func() {
		defer cancel()
		exit <- isChainListening(ctx, addressAPI)
	}()

	// 常規鏈啟動
	go func() {
		// 如果錯誤是驗證器設置為 nil，這意味著在應用請求後創世沒有被破壞
		// 創世紀已正確生成，但目前還沒有 gentxs
		// 所以我們不認為這是一個錯誤，使請求驗證無效
		err := cmd.Start(ctx)
		if err != nil && strings.Contains(err.Error(), ValidatorSetNilErrorMessage) {
			err = nil
		}
		exit <- errors.Wrap(err, "熊網鏈無法啟動")
	}()

	return <-exit
}

// setSimulationConfig 在配置中設置隨機可用端口，以允許檢查鍊網絡是否可以啟動
func (c Chain) setSimulationConfig() (string, error) {
	// 生成隨機服務器端口和服務器列表
	ports, err := availableport.Find(5)
	if err != nil {
		return "", err
	}
	genAddr := func(port int) string {
		return fmt.Sprintf("localhost:%d", port)
	}

	//更新應用程序 toml
	appPath, err := c.AppTOMLPath()
	if err != nil {
		return "", err
	}
	config, err := toml.LoadFile(appPath)
	if err != nil {
		return "", err
	}

	apiAddr, err := xurl.TCP(genAddr(ports[0]))
	if err != nil {
		return "", err
	}

	config.Set("api.enable", true)
	config.Set("api.enabled-unsafe-cors", true)
	config.Set("rpc.cors_allowed_origins", []string{"*"})
	config.Set("api.address", apiAddr)
	config.Set("grpc.address", genAddr(ports[1]))

	file, err := os.OpenFile(appPath, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := config.WriteTo(file); err != nil {
		return "", err
	}

	// updating config toml
	configPath, err := c.ConfigTOMLPath()
	if err != nil {
		return "", err
	}
	config, err = toml.LoadFile(configPath)
	if err != nil {
		return "", err
	}

	rpcAddr, err := xurl.TCP(genAddr(ports[2]))
	if err != nil {
		return "", err
	}

	p2pAddr, err := xurl.TCP(genAddr(ports[3]))
	if err != nil {
		return "", err
	}

	config.Set("rpc.cors_allowed_origins", []string{"*"})
	config.Set("consensus.timeout_commit", "1s")
	config.Set("consensus.timeout_propose", "1s")
	config.Set("rpc.laddr", rpcAddr)
	config.Set("p2p.laddr", p2pAddr)
	config.Set("rpc.pprof_laddr", genAddr(ports[4]))

	file, err = os.OpenFile(configPath, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = config.WriteTo(file)

	return genAddr(ports[0]), err
}

// isChainListening 檢查鍊是否正在偵聽指定地址上的 API 查詢
func isChainListening(ctx context.Context, addressAPI string) error {
	checkAlive := func() error {
		addr, err := xurl.HTTP(addressAPI)
		if err != nil {
			return fmt.Errorf("api地址格式無效 %s: %w", addressAPI, err)
		}

		ok, err := httpstatuschecker.Check(ctx, fmt.Sprintf("%s/node_info", addr))
		if err == nil && !ok {
			err = errors.New("應用不在線")
		}
		return err
	}
	return backoff.Retry(checkAlive, backoff.WithContext(backoff.NewConstantBackOff(time.Second), ctx))
}
