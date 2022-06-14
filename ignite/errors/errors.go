// 包 sperrors 包含 starport 特定錯誤。
package sperrors

import "errors"

var (
	// ErrOnlyStargateSupported 當底層鏈不是星門鏈時返回。
	ErrOnlyStargateSupported = errors.New("不再支持此版本的 Cosmos SDK")
)
