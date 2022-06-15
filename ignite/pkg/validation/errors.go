package validation

// 錯誤必須由提供驗證信息的錯誤實現。
type Error interface {
	error
	ValidationInfo() string
}
