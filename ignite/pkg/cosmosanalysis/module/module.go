package module

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bearnetworkchain/core/ignite/pkg/cosmosanalysis"
	"github.com/bearnetworkchain/core/ignite/pkg/cosmosanalysis/app"
	"github.com/bearnetworkchain/core/ignite/pkg/gomodule"
	"github.com/bearnetworkchain/core/ignite/pkg/protoanalysis"
)

// Msgs 是一個模塊導入路徑-SDK 消息對。
type Msgs map[string][]string

// Module 保留有關 Cosmos SDK 模塊的元數據.
type Module struct {
	// Name of the module.
	Name string

	// GoModulePath 定義模塊的應用程序。
	GoModulePath string

	// Pkg 保存原始包信息。
	Pkg protoanalysis.Package

	// Msg 是模塊的 sdk.Msg 實現列表。
	Msgs []Msg

	// HTTPQueries 是模塊查詢的列表。
	HTTPQueries []HTTPQuery

	// Types是模塊可能使用的原型類型列表。
	Types []Type
}

// Msg 保留有關 sdk.Msg 實現的元數據。
type Msg struct {
	// 類型的名稱。
	Name string

	// 類型的 URI。
	URI string

	// FilePath 是定義消息的 .proto 文件的路徑。
	FilePath string
}

// HTTPQuery 是一個 sdk 查詢。
type HTTPQuery struct {
	// RPC 函數的名稱。
	Name string

	// 帶有服務名稱和 rpc 函數名稱的查詢的全名。
	FullName string

	// HTTPAnnotations 保存有關查詢的 http 註釋的信息。
	Rules []protoanalysis.HTTPRule
}

// Type 是模塊可能使用的原型類型。
type Type struct {
	Name string

	// FilePath 是定義消息的 .proto 文件的路徑。
	FilePath string
}

type moduleDiscoverer struct {
	sourcePath        string
	protoPath         string
	basegopath        string
	registeredModules []string
}

// Discover 發現並返回在應用中註冊的模塊及其類型
// chainRoot 是鏈的根路徑
// sourcePath 是 proto dir 所在的 go 模塊的根路徑
//
// 發現算法利用註冊模塊和原型定義來查找相關的註冊模塊
// 它通過以下方式實現：
// 1. 從app中獲取所有註冊的go模塊
// 2. 解析 proto 文件找到服務和消息
// 3. 檢查 proto 服務是否在任何註冊的模塊中實現
func Discover(ctx context.Context, chainRoot, sourcePath, protoDir string) ([]Module, error) {
	// 找出區塊鏈的基本 Go 導入路徑。
	gm, err := gomodule.ParseAt(sourcePath)
	if err != nil {
		if err == gomodule.ErrGoModNotFound {
			return []Module{}, nil
		}
		return nil, err
	}

	registeredModules, err := app.FindRegisteredModules(chainRoot)
	if err != nil {
		return nil, err
	}

	basegopath := gm.Module.Mod.Path

	// 只過濾掉這裡可能不相關的註冊模塊
	potentialModules := make([]string, 0)
	for _, m := range registeredModules {
		if strings.HasPrefix(m, basegopath) {
			potentialModules = append(potentialModules, m)
		}
	}
	if len(potentialModules) == 0 {
		return []Module{}, nil
	}

	md := &moduleDiscoverer{
		protoPath:         filepath.Join(sourcePath, protoDir),
		sourcePath:        sourcePath,
		basegopath:        basegopath,
		registeredModules: potentialModules,
	}

	// 查找屬於 x/ 下模塊​​的 proto 包。
	pkgs, err := md.findModuleProtoPkgs(ctx)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return []Module{}, nil
	}

	var modules []Module

	for _, pkg := range pkgs {
		m, err := md.discover(pkg)
		if err != nil {
			return nil, err
		}

		modules = append(modules, m)
	}

	return modules, nil
}

// 通過 proto pkg 發現發現和 sdk 模塊。
func (d *moduleDiscoverer) discover(pkg protoanalysis.Package) (Module, error) {
	pkgrelpath := strings.TrimPrefix(pkg.GoImportPath(), d.basegopath)
	pkgpath := filepath.Join(d.sourcePath, pkgrelpath)

	found, err := d.pkgIsFromRegisteredModule(pkg)
	if err != nil {
		return Module{}, err
	}
	if !found {
		return Module{}, nil
	}

	msgs, err := cosmosanalysis.FindImplementation(pkgpath, messageImplementation)
	if err != nil {
		return Module{}, err
	}

	if len(pkg.Services)+len(msgs) == 0 {
		return Module{}, nil
	}

	namesplit := strings.Split(pkg.Name, ".")
	m := Module{
		Name:         namesplit[len(namesplit)-1],
		GoModulePath: d.basegopath,
		Pkg:          pkg,
	}

	// fill sdk Msgs.
	for _, msg := range msgs {
		pkgmsg, err := pkg.MessageByName(msg)
		if err != nil {
			// no msg found in the proto defs 對應於發現的 sdk 消息。
			// 如果找不到，不用擔心，這意味著它被使用了
			// 僅限內部使用，不開放供實際使用。
			continue
		}

		m.Msgs = append(m.Msgs, Msg{
			Name:     msg,
			URI:      fmt.Sprintf("%s.%s", pkg.Name, msg),
			FilePath: pkgmsg.Path,
		})
	}

	// isType 是否可以將 protomsg 作為任何類型添加到模塊。
	isType := func(protomsg protoanalysis.Message) bool {
		// 不要使用創世狀態類型。
		if protomsg.Name == "GenesisState" {
			return false
		}

		// 如果 SDK 消息，請勿使用。
		for _, msg := range msgs {
			if msg == protomsg.Name {
				return false
			}
		}

		// 如果用作 RPC 的請求/返回類型，請勿使用。
		for _, s := range pkg.Services {
			for _, q := range s.RPCFuncs {
				if q.RequestType == protomsg.Name || q.ReturnsType == protomsg.Name {
					return false
				}
			}
		}

		return true
	}

	//填充類型。
	for _, protomsg := range pkg.Messages {
		if !isType(protomsg) {
			continue
		}

		m.Types = append(m.Types, Type{
			Name:     protomsg.Name,
			FilePath: protomsg.Path,
		})
	}

	//填寫查詢。
	for _, s := range pkg.Services {
		for _, q := range s.RPCFuncs {
			if len(q.HTTPRules) == 0 {
				continue
			}
			m.HTTPQueries = append(m.HTTPQueries, HTTPQuery{
				Name:     q.Name,
				FullName: s.Name + q.Name,
				Rules:    q.HTTPRules,
			})
		}
	}

	return m, nil
}

func (d *moduleDiscoverer) findModuleProtoPkgs(ctx context.Context) ([]protoanalysis.Package, error) {
	// 找出區塊鏈中的所有原始包。
	allprotopkgs, err := protoanalysis.Parse(ctx, nil, d.protoPath)
	if err != nil {
		return nil, err
	}

	// 過濾掉不代表區塊鏈 x/ 模塊的 proto 包。
	var xprotopkgs []protoanalysis.Package
	for _, pkg := range allprotopkgs {
		if !strings.HasPrefix(pkg.GoImportName, d.basegopath) {
			continue
		}

		xprotopkgs = append(xprotopkgs, pkg)
	}

	return xprotopkgs, nil
}

// 檢查 pkg 是否在任何已註冊的模塊中實現
func (d *moduleDiscoverer) pkgIsFromRegisteredModule(pkg protoanalysis.Package) (bool, error) {
	for _, rm := range d.registeredModules {
		implRelPath := strings.TrimPrefix(rm, d.basegopath)
		implPath := filepath.Join(d.sourcePath, implRelPath)

		for _, s := range pkg.Services {
			methods := make([]string, len(s.RPCFuncs))
			for i, rpcFunc := range s.RPCFuncs {
				methods[i] = rpcFunc.Name
			}
			found, err := cosmosanalysis.DeepFindImplementation(implPath, methods)
			if err != nil {
				return false, err
			}

			// 在某些情況下，模塊註冊在模塊的另一層子目錄中。
			// 全部: 在 proto 包中找到最近的子目錄。
			if len(found) == 0 && strings.HasPrefix(rm, pkg.GoImportName) {
				altImplRelPath := strings.TrimPrefix(pkg.GoImportName, d.basegopath)
				altImplPath := filepath.Join(d.sourcePath, altImplRelPath)
				found, err = cosmosanalysis.DeepFindImplementation(altImplPath, methods)
				if err != nil {
					return false, err
				}
			}

			if len(found) > 0 {
				return true, nil
			}
		}

	}

	return false, nil
}
