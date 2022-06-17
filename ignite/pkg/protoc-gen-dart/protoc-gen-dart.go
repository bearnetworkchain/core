package protocgendart

import (
	"fmt"

	"github.com/bearnetworkchain/core/ignite/pkg/localfs"
	"github.com/bearnetworkchain/core/ignite/pkg/protoc-gen-dart/data"
)

// 插件的名稱。
const Name = "protoc-gen-dart"

// BinaryPath 返回插件的二進制路徑。
func BinaryPath() (path string, cleanup func(), err error) {
	return localfs.SaveBytesTemp(data.Binary(), Name, 0755)
}

// 標誌返回二進制名稱-二進制路徑格式以傳遞給 protoc --plugin。
func Flag() (flag string, cleanup func(), err error) {
	path, cleanup, err := BinaryPath()
	flag = fmt.Sprintf("%s=%s", Name, path)
	return
}
