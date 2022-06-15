package xgenny

import (
	"bytes"
	"embed"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/packd"
)

// Walker 為 Go embed 的 fs.FS 實現了 packd.Walker。
type Walker struct {
	fs         embed.FS
	trimPrefix string
	path       string
}

// NewEmbedWalker 為 fs 返回一個新的 Walker。
// trimPrefix 用於從找到的文件的路徑中修剪父路徑。
func NewEmbedWalker(fs embed.FS, trimPrefix, path string) Walker {
	return Walker{fs: fs, trimPrefix: trimPrefix, path: path}
}

// Walk 實現 packd.Walker。
func (w Walker) Walk(wl packd.WalkFunc) error {
	return w.walkDir(wl, ".")
}

func (w Walker) walkDir(wl packd.WalkFunc, path string) error {
	entries, err := w.fs.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			w.walkDir(wl, filepath.Join(path, entry.Name()))
			continue
		}

		path := filepath.Join(path, entry.Name())

		data, err := w.fs.ReadFile(path)
		if err != nil {
			return err
		}

		ppath := strings.TrimPrefix(path, w.trimPrefix)
		ppath = filepath.Join(w.path, ppath)
		f, err := packd.NewFile(ppath, bytes.NewReader(data))
		if err != nil {
			return err
		}

		wl(ppath, f)
	}

	return nil
}
