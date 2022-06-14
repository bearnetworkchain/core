// Package cosmosanalysis 提供靜態分析 Cosmos SDK 的工具集
// 基於 Cosmos SDK 的源碼和區塊鏈源碼
package cosmosanalysis

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/mod/modfile"
)

const (
	cosmosModulePath     = "github.com/cosmos/cosmos-sdk"
	tendermintModulePath = "github.com/tendermint/tendermint"
	appFileName          = "app.go"
	defaultAppFilePath   = "app/" + appFileName
)

var appImplementation = []string{
	"Name",
	"BeginBlocker",
	"EndBlocker",
}

// implementation 跟踪給定結構的接口的實現
type implementation map[string]bool

// DeepFindImplementation 和查找實施一樣，但是遞歸遍歷文件夾結構
// 如果實現可能位於子文件夾中，則很有用
func DeepFindImplementation(modulePath string, interfaceList []string) (found []string, err error) {
	err = filepath.Walk(modulePath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return nil
			}

			currFound, err := FindImplementation(path, interfaceList)
			if err != nil {
				return err
			}

			found = append(found, currFound...)
			return nil
		})

	if err != nil {
		return nil, err
	}

	return found, nil
}

// FindImplementation 查找實現所提供接口的所有類型的名稱
func FindImplementation(modulePath string, interfaceList []string) (found []string, err error) {
	// 解析路徑下的包/文件
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, modulePath, nil, 0)
	if err != nil {
		return nil, err
	}
	for _, pkg := range pkgs {
		var files []*ast.File
		for _, f := range pkg.Files {
			files = append(files, f)
		}
		found = append(found, findImplementationInFiles(files, interfaceList)...)
	}

	return found, nil
}

func findImplementationInFiles(files []*ast.File, interfaceList []string) (found []string) {
	// 收集路徑下的所有結構，找出滿足實現的結構
	structImplementations := make(map[string]implementation)

	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			// 尋找結構方法。
			methodDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// 不是方法。
			if methodDecl.Recv == nil {
				return true
			}

			methodName := methodDecl.Name.Name

			// 找到該方法所屬的結構名稱。
			t := methodDecl.Recv.List[0].Type
			ident, ok := t.(*ast.Ident)
			if !ok {
				sexp, ok := t.(*ast.StarExpr)
				if !ok {
					return true
				}
				ident = sexp.X.(*ast.Ident)
			}
			structName := ident.Name

			// 標記此結構滿足的實現。
			if _, ok := structImplementations[structName]; !ok {
				structImplementations[structName] = newImplementation(interfaceList)
			}

			structImplementations[structName][methodName] = true

			return true
		})
	}

	for name, impl := range structImplementations {
		if checkImplementation(impl) {
			found = append(found, name)
		}
	}

	return found
}

// newImplementation 返回一個新對象來解析接口的實現
func newImplementation(interfaceList []string) implementation {
	impl := make(implementation)
	for _, m := range interfaceList {
		impl[m] = false
	}
	return impl
}

// checkImplementation 檢查整個實現是否滿足
func checkImplementation(r implementation) bool {
	for _, ok := range r {
		if !ok {
			return false
		}
	}
	return true
}

// ValidateGoMod 檢查cosmos-sdk和tendermint包是否被導入.
func ValidateGoMod(module *modfile.File) error {
	moduleCheck := map[string]bool{
		cosmosModulePath:     true,
		tendermintModulePath: true,
	}
	for _, r := range module.Require {
		delete(moduleCheck, r.Mod.Path)
	}
	for m := range moduleCheck {
		return fmt.Errorf("無效的GO模塊，丟失 %s 依賴包", m)
	}
	return nil
}

// FindAppFilePath 查找實現了應用實施中列出的接口的應用文件
func FindAppFilePath(chainRoot string) (path string, err error) {
	var found []string

	err = filepath.Walk(chainRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" {
			return nil
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		currFound := findImplementationInFiles([]*ast.File{f}, appImplementation)

		if len(currFound) > 0 {
			found = append(found, path)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	numFound := len(found)
	if numFound == 0 {
		return "", errors.New("app.go 找不到文件")
	}

	if numFound == 1 {
		return found[0], nil
	}

	appFilePath := ""
	for _, p := range found {
		if filepath.Base(p) == appFileName {
			if appFilePath != "" {
				// 找到多個 app.go，回退到 app/app.go
				return getDefaultAppFile(chainRoot)
			}

			appFilePath = p
		}
	}

	if appFilePath != "" {
		return appFilePath, nil
	}

	return getDefaultAppFile(chainRoot)
}

// getDefaultAppFile 返回鏈的默認 app.go 文件路徑。
func getDefaultAppFile(chainRoot string) (string, error) {
	path := filepath.Join(chainRoot, defaultAppFilePath)
	_, err := os.Stat(path)
	return path, errors.Wrap(err, "找不到您的app.go")
}
