package cosmosfaucet

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// faucetTimeout 用於在從水龍頭轉移硬幣時設置超時。
const faucetTimeout = time.Second * 20

// TryRetrieve 嘗試從水龍頭中檢索令牌。水龍頭地址在提供時使用。
// 否則，它將嘗試從鏈的 rpc 地址中猜測水龍頭地址。
// 如果無法確定水龍頭的地址或硬幣檢索不成功，則返回非零錯誤。
func TryRetrieve(
	ctx context.Context,
	chainID,
	rpcAddress,
	faucetAddress,
	accountAddress string,
) error {
	var faucetURL *url.URL
	var err error

	if faucetAddress != "" {
		// 如果有用戶給定水龍頭地址，則使用。
		faucetURL, err = url.Parse(faucetAddress)
	} else {
		// 找到水龍頭網址。可以是用戶給定的，否則是猜測的。
		faucetURL, err = discoverFaucetURL(ctx, chainID, rpcAddress)
	}
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, faucetTimeout)
	defer cancel()

	fc := NewClient(faucetURL.String())

	resp, err := fc.Transfer(ctx, TransferRequest{
		AccountAddress: accountAddress,
	})
	if err != nil {
		return errors.Wrap(err, "水龍頭不工作")
	}
	if resp.Error != "" {
		return fmt.Errorf("水龍頭不工作: %s", resp.Error)
	}

	return nil
}

func discoverFaucetURL(ctx context.Context, chainID, rpcAddress string) (*url.URL, error) {
	// 否則猜測水龍頭地址.
	guessedURLs, err := guessFaucetURLs(rpcAddress)
	if err != nil {
		return nil, err
	}

	for _, u := range guessedURLs {
		// check if the potential faucet server accepts connections.
		address := u.Host
		if u.Scheme == "https" {
			address += ":443"
		}
		if _, err := net.DialTimeout("tcp", address, time.Second); err != nil {
			continue
		}

		// 確保這是一個真正的水龍頭服務器。
		info, err := NewClient(u.String()).FaucetInfo(ctx)
		if err != nil || info.ChainID != chainID || !info.IsAFaucet {
			continue
		}

		return u, nil
	}

	return nil, errors.New("沒有可用的水龍頭, 請發送硬幣到地址")
}

// guess 嘗試猜測所有可能的水龍頭地址。
func guessFaucetURLs(rpcAddress string) ([]*url.URL, error) {
	u, err := url.Parse(rpcAddress)
	if err != nil {
		return nil, err
	}

	var guessedURLs []*url.URL

	possibilities := []struct {
		port         string
		subname      string
		nameSperator string
	}{
		{"4500", "", "."},
		{"", "faucet", "."},
		{"", "4500", "-"}, //Gitpod 使用端口號作為子域名。
	}

	// 通過基於 RPC 地址創建猜測地址。
	for _, poss := range possibilities {
		guess, _ := url.Parse(u.String())                  //複製原始網址。
		for _, scheme := range []string{"http", "https"} { //為這兩個方案做。
			guess, _ := url.Parse(guess.String()) // 複製猜測。
			guess.Scheme = scheme

			// 嘗試使用端口號。
			if poss.port != "" {
				guess.Host = fmt.Sprintf("%s:%s", u.Hostname(), "4500")
				guessedURLs = append(guessedURLs, guess)
				continue
			}

			// 嘗試使用子名稱。
			if poss.subname != "" {
				bases := []string{
					// 嘗試將子名稱附加到默認名稱。
					// e.g.: faucet.my.domain.
					u.Hostname(),
				}

				// 嘗試替換 1 級的子名。
				// e.g.: faucet.domain.
				sp := strings.SplitN(u.Hostname(), poss.nameSperator, 2)
				if len(sp) == 2 {
					bases = append(bases, sp[1])
				}
				for _, basename := range bases {
					guess, _ := url.Parse(guess.String()) // 複製猜測。
					guess.Host = fmt.Sprintf("%s%s%s", poss.subname, poss.nameSperator, basename)
					guessedURLs = append(guessedURLs, guess)
				}
			}
		}
	}

	return guessedURLs, nil
}
