package chain

import (
	"path/filepath"
	"strings"

	"github.com/bearnetworkchain/core/ignite/pkg/gomodulepath"
)

// 應用保存有關鏈的信息。
type App struct {
	Name       string
	Path       string
	ImportPath string
}

// NewAppAt 從位於 path 的區塊鏈源代碼創建一個應用程序。
func NewAppAt(path string) (App, error) {
	p, appPath, err := gomodulepath.Find(path)
	if err != nil {
		return App{}, err
	}
	return App{
		Path:       appPath,
		Name:       p.Root,
		ImportPath: p.RawPath,
	}, nil
}

// N 返回不帶破折號的應用名稱。
func (a App) N() string {
	return strings.ReplaceAll(a.Name, "-", "")
}

// D 返回應用程序名稱。
func (a App) D() string {
	return a.Name + "d"
}

// ND 返回無破折號的應用程序名稱。
func (a App) ND() string {
	return a.N() + "d"
}

// Root 返回app的根路徑。
func (a App) Root() string {
	path, _ := filepath.Abs(a.Path)
	return path
}
