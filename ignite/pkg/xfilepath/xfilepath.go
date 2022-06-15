// Package xfilepath 定義函數來定義支持錯誤處理的路徑檢索器
package xfilepath

import (
	"os"
	"path/filepath"
)

// PathRetriever 是檢索包含路徑或錯誤的函數
type PathRetriever func() (path string, err error)

// PathsRetriever 是檢索包含的路徑列表或錯誤的函數
type PathsRetriever func() (path []string, err error)

// Path 從提供的路徑返迴路徑檢索器
func Path(path string) PathRetriever {
	return func() (string, error) { return path, nil }
}

// PathWithError 從提供的路徑和錯誤返迴路徑檢索器
func PathWithError(path string, err error) PathRetriever {
	return func() (string, error) { return path, err }
}

// Join 從提供的路徑檢索器的連接中返回一個路徑檢索器
// 返回的路徑檢索器最終從返回非零錯誤的第一個提供的路徑檢索器返回錯誤
func Join(paths ...PathRetriever) PathRetriever {
	var components []string
	var err error
	for _, path := range paths {
		var component string
		component, err = path()
		if err != nil {
			break
		}
		components = append(components, component)
	}
	path := filepath.Join(components...)

	return func() (string, error) {
		return path, err
	}
}

// JoinFromHome從用戶 home 和提供的路徑檢索器的連接返迴路徑檢索器
// 返回的路徑檢索器最終從返回非零錯誤的第一個提供的路徑檢索器返回錯誤
func JoinFromHome(paths ...PathRetriever) PathRetriever {
	return Join(append([]PathRetriever{os.UserHomeDir}, paths...)...)
}

// List 從路徑檢索器列表中返回一個路徑檢索器
// 返回的路徑檢索器最終從返回非零錯誤的第一個提供的路徑檢索器返回錯誤
func List(paths ...PathRetriever) PathsRetriever {
	var list []string
	var err error
	for _, path := range paths {
		var resolved string
		resolved, err = path()
		if err != nil {
			break
		}
		list = append(list, resolved)
	}

	return func() ([]string, error) {
		return list, err
	}
}
