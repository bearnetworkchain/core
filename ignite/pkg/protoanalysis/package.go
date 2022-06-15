package protoanalysis

import (
	"strings"

	"github.com/pkg/errors"
)

type Packages []Package

func (p Packages) Files() Files {
	var files []File
	for _, pkg := range p {
		files = append(files, pkg.Files...)
	}
	return files
}

// 包代表一個proto pkg。
type Package struct {
	// 原型 pkg 的名稱。
	Name string

	// fs 中包的路徑。
	Path string

	// Files 是包中 .proto 文件的列表。
	Files Files

	// GoImportName 是 proto 包的 go 包名。
	GoImportName string

	// Messages 是包中定義的原始消息列表。
	Messages []Message

	// Services 是 RPC 服務的列表。
	Services []Service
}

type Files []File

type File struct {
	//文件的路徑。
	Path string

	//Dependencies 是此包中導入的 .proto 文件的列表。
	Dependencies []string
}

func (f Files) Paths() []string {
	var paths []string
	for _, ff := range f {
		paths = append(paths, ff.Path)
	}
	return paths
}

// MessageByName 在 Package 中按其名稱查找消息。
func (p Package) MessageByName(name string) (Message, error) {
	for _, message := range p.Messages {
		if message.Name == name {
			return message, nil
		}
	}
	return Message{}, errors.New("未找到消息")
}

// GoImportPath 檢索 Go 導入路徑。
func (p Package) GoImportPath() string {
	return strings.Split(p.GoImportName, ";")[0]
}

// Message 表示原始消息。
type Message struct {
	// 消息的名稱。
	Name string

	//定義消息的文件的路徑。
	Path string

// HighestFieldNumber 是消息字段中最高的字段號
// 這允許在寫入 proto 消息時確定新的字段編號
	HighestFieldNumber int
}

//服務是一個 RPC 服務。
type Service struct {
	//服務的名稱。
	Name string

	//RPC 是服務的 RPC 函數列表。
	RPCFuncs []RPCFunc
}

//RPCFunc 是一個 RPC 函數。
type RPCFunc struct {
	// RPC 函數的名稱。
	Name string

	// RequestType 是 RPC 函數的請求類型。
	RequestType string

	// ReturnsType 是 RPC 函數的響應類型。
	ReturnsType string

// HTTPRules 保存有關 RPC 函數的 http 規則的信息。
// 規格：
//   https://github.com/googleapis/googleapis/blob/master/google/api/http.proto.
	HTTPRules []HTTPRule
}

//HTTPRule 保存有關 RPC 函數的已配置 http 規則的信息。
type HTTPRule struct {
	// Params 是在 http 端點本身中定義的參數列表。
	Params []string

	//HasQuery 指示是否有請求查詢。
	HasQuery bool

	// HasBody 指示是否存在請求有效負載。
	HasBody bool
}
