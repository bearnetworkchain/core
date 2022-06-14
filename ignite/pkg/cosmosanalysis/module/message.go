package module

// 消息實現是 sdk.Msg 實現所需的方法列表
// 全部（低優先級）：從底層sdk的源代碼中動態獲取這些。
var messageImplementation = []string{
	"Route",
	"Type",
	"GetSigners",
	"GetSignBytes",
	"ValidateBasic",
}
