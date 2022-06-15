package localfs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	wt "github.com/radovskyb/watcher"
)

type watcher struct {
	wt            *wt.Watcher
	workdir       string
	ignoreHidden  bool
	ignoreFolders bool
	ignoreExts    []string
	onChange      func()
	interval      time.Duration
	ctx           context.Context
	done          *sync.WaitGroup
}

// WatcherOption 用於配置觀察者。
type WatcherOption func(*watcher)

// WatcherWorkdir需要注意設置為路徑的根.
func WatcherWorkdir(path string) WatcherOption {
	return func(w *watcher) {
		w.workdir = path
	}
}

// WatcherOnChange設置一個在文件系統上的每次更改時執行的鉤子。
func WatcherOnChange(hook func()) WatcherOption {
	return func(w *watcher) {
		w.onChange = hook
	}
}

// WatcherPollingInterval 覆蓋默認輪詢間隔以檢查文件系統更改。
func WatcherPollingInterval(d time.Duration) WatcherOption {
	return func(w *watcher) {
		w.interval = d
	}
}

// WatcherIgnoreHidden忽略隱藏（點）文件。
func WatcherIgnoreHidden() WatcherOption {
	return func(w *watcher) {
		w.ignoreHidden = true
	}
}

func WatcherIgnoreFolders() WatcherOption {
	return func(w *watcher) {
		w.ignoreFolders = true
	}
}

// WatcherIgnoreExt忽略具有匹配文件擴展名的文件。
func WatcherIgnoreExt(exts ...string) WatcherOption {
	return func(w *watcher) {
		w.ignoreExts = exts
	}
}

// Watch 開始觀察路徑上的變化。選項用於配置
// watch 操作的行為。
func Watch(ctx context.Context, paths []string, options ...WatcherOption) error {
	w := &watcher{
		wt:       wt.New(),
		onChange: func() {},
		interval: time.Millisecond * 300,
		done:     &sync.WaitGroup{},
		ctx:      ctx,
	}
	w.wt.SetMaxEvents(1)

	for _, o := range options {
		o(w)
	}

	w.wt.AddFilterHook(func(info os.FileInfo, fullPath string) error {
		if info.IsDir() && w.ignoreFolders {
			return wt.ErrSkip
		}
		if w.isFileIgnored(fullPath) {
			return wt.ErrSkip
		}

		return nil
	})

	// 忽略隱藏路徑。
	w.wt.IgnoreHiddenFiles(w.ignoreHidden)

	//添加要觀看的路徑
	if err := w.addPaths(paths...); err != nil {
		return err
	}

	// 開始觀看。
	w.done.Add(1)
	go w.listen()
	if err := w.wt.Start(w.interval); err != nil {
		return err
	}
	w.done.Wait()
	return nil
}

func (w *watcher) listen() {
	defer w.done.Done()
	for {
		select {
		case <-w.wt.Event:
			w.onChange()
		case <-w.wt.Closed:
			return
		case <-w.ctx.Done():
			w.wt.Close()
		}
	}
}

func (w *watcher) addPaths(paths ...string) error {
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(w.workdir, path)
		}

		// 忽略不存在的路徑
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err := w.wt.AddRecursive(path); err != nil {
			return err
		}
	}

	return nil
}

func (w *watcher) isFileIgnored(path string) bool {
	for _, ext := range w.ignoreExts {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
