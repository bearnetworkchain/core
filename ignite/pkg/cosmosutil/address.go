package cosmosutil

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// ChangeAddressPrefix 返回帶有另一個前綴的地址
func ChangeAddressPrefix(address, newPrefix string) (string, error) {
	if newPrefix == "" {
		return "", errors.New("空前綴")
	}
	_, pubKey, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return "", err
	}
	return bech32.ConvertAndEncode(newPrefix, pubKey)
}

// GetAddressPrefix 返回地址使用的 bech 32 前綴
func GetAddressPrefix(address string) (string, error) {
	prefix, _, err := bech32.DecodeAndConvert(address)
	return prefix, err
}
