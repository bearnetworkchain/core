package app

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/bearnetworkchain/core/ignite/pkg/cosmosanalysis"
	"github.com/bearnetworkchain/core/ignite/pkg/goanalysis"
)

var appImplementation = []string{
	"RegisterAPIRoutes",
	"RegisterTxService",
	"RegisterTendermintService",
}

// CheckKeeper 在應用程序結構中檢查是否存在具有提供的名稱的門將
func CheckKeeper(path, keeperName string) error {
	// find app type
	appImpl, err := cosmosanalysis.FindImplementation(path, appImplementation)
	if err != nil {
		return err
	}
	if len(appImpl) != 1 {
		return errors.New("app.go 應該包含一個應用程序")
	}
	appTypeName := appImpl[0]

	// 檢查模塊的 app 結構
	var found bool
	fileSet := token.NewFileSet()
	pkgs, err := parser.ParseDir(fileSet, path, nil, 0)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				// 尋找結構方法。
				appType, ok := n.(*ast.TypeSpec)
				if !ok || appType.Name.Name != appTypeName {
					return true
				}

				appStruct, ok := appType.Type.(*ast.StructType)
				if !ok {
					return true
				}

				// 搜索 keeper 特定字段
				for _, field := range appStruct.Fields.List {
					for _, fieldName := range field.Names {
						if fieldName.Name == keeperName {
							found = true
							return false
						}
					}
				}

				return false
			})
		}
	}

	if !found {
		return fmt.Errorf("應用程序不包含 %s", keeperName)
	}
	return nil
}

// FindRegisteredModules 查找 App 中所有已註冊的模塊
// 它通過檢查導入的模塊是否已在應用程序中註冊並檢查其查詢客戶端是否已註冊來查找激活的模塊
// 它通過以下方式實現：
// 1. 映射出所有導入和命名導入
// 2. 尋找對 module.NewBasicManager 的調用並找到在那裡註冊的模塊
// 3.尋找RegisterAPIRoutes的實現，找到調用其RegisterGRPCGatewayRoutes的模塊
func FindRegisteredModules(chainRoot string) ([]string, error) {
	appFilePath, err := cosmosanalysis.FindAppFilePath(chainRoot)
	if err != nil {
		return nil, err
	}

	fileSet := token.NewFileSet()
	f, err := parser.ParseFile(fileSet, appFilePath, nil, 0)
	if err != nil {
		return []string{}, err
	}

	packages, err := goanalysis.FindImportedPackages(appFilePath)
	if err != nil {
		return nil, err
	}

	basicManagerModule, err := findBasicManagerModule(packages)
	if err != nil {
		return nil, err
	}

	var basicModules []string
	ast.Inspect(f, func(n ast.Node) bool {
		if pkgsReg := findBasicManagerRegistrations(n, basicManagerModule); pkgsReg != nil {
			for _, rp := range pkgsReg {
				importModule := packages[rp]
				basicModules = append(basicModules, importModule)
			}

			return false
		}

		if pkgsReg := findRegisterAPIRoutersRegistrations(n); pkgsReg != nil {
			for _, rp := range pkgsReg {
				importModule := packages[rp]
				if importModule == "" {
					continue
				}
				basicModules = append(basicModules, importModule)
			}

			return false
		}

		return true
	})

	return basicModules, nil
}

func findBasicManagerRegistrations(n ast.Node, basicManagerModule string) []string {
	callExprType, ok := n.(*ast.CallExpr)
	if !ok {
		return nil
	}

	selectorExprType, ok := callExprType.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	identExprType, ok := selectorExprType.X.(*ast.Ident)
	if !ok || identExprType.Name != basicManagerModule || selectorExprType.Sel.Name != "NewBasicManager" {
		return nil
	}

	packagesRegistered := make([]string, len(callExprType.Args))
	for i, arg := range callExprType.Args {
		argAsCompositeLitType, ok := arg.(*ast.CompositeLit)
		if ok {
			compositeTypeSelectorExpr, ok := argAsCompositeLitType.Type.(*ast.SelectorExpr)
			if !ok {
				continue
			}

			compositeTypeX, ok := compositeTypeSelectorExpr.X.(*ast.Ident)
			if ok {
				packagesRegistered[i] = compositeTypeX.Name
				continue
			}
		}

		argAsCallType, ok := arg.(*ast.CallExpr)
		if ok {
			argAsFunctionType, ok := argAsCallType.Fun.(*ast.SelectorExpr)
			if !ok {
				continue
			}

			argX, ok := argAsFunctionType.X.(*ast.Ident)
			if ok {
				packagesRegistered[i] = argX.Name
			}
		}
	}

	return packagesRegistered
}

func findBasicManagerModule(pkgs map[string]string) (string, error) {
	for mod, pkg := range pkgs {
		if pkg == "github.com/cosmos/cosmos-sdk/types/module" {
			return mod, nil
		}
	}

	return "", errors.New("沒有找到 Basic Manager 的模塊")
}

func findRegisterAPIRoutersRegistrations(n ast.Node) []string {
	funcLitType, ok := n.(*ast.FuncDecl)
	if !ok {
		return nil
	}

	if funcLitType.Name.Name != "RegisterAPIRoutes" {
		return nil
	}

	var packagesRegistered []string
	for _, stmt := range funcLitType.Body.List {
		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}

		exprCall, ok := exprStmt.X.(*ast.CallExpr)
		if !ok {
			continue
		}

		exprFun, ok := exprCall.Fun.(*ast.SelectorExpr)
		if !ok || exprFun.Sel.Name != "RegisterGRPCGatewayRoutes" {
			continue
		}

		identType, ok := exprFun.X.(*ast.Ident)
		if !ok {
			continue
		}

		pkgName := identType.Name
		if pkgName == "" {
			continue
		}

		packagesRegistered = append(packagesRegistered, identType.Name)
	}

	return packagesRegistered
}
