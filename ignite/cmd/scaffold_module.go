package ignitecmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/cliui/clispinner"
	"github.com/bearnetworkchain/core/ignite/pkg/placeholder"
	"github.com/bearnetworkchain/core/ignite/pkg/validation"
	"github.com/bearnetworkchain/core/ignite/services/scaffolder"
	modulecreate "github.com/bearnetworkchain/core/ignite/templates/module/create"
)

const (
	flagDep                 = "dep"
	flagIBC                 = "ibc"
	flagParams              = "params"
	flagIBCOrdering         = "ordering"
	flagRequireRegistration = "require-registration"
)

// NewScaffoldModule 返回為 Cosmos SDK 模塊搭建基架的命令
func NewScaffoldModule() *cobra.Command {
	c := &cobra.Command{
		Use:   "module [name]",
		Short: "搭建一個 Cosmos SDK 模塊",
		Long:  "在 `x` 目錄中搭建一個新的 Cosmos SDK 模塊",
		Args:  cobra.MinimumNArgs(1),
		RunE:  scaffoldModuleHandler,
	}

	flagSetPath(c)
	flagSetClearCache(c)
	c.Flags().StringSlice(flagDep, []string{}, "模塊依賴項（例如 --dep account,bank）")
	c.Flags().Bool(flagIBC, false, "scaffold an IBC module")
	c.Flags().String(flagIBCOrdering, "none", "IBC 模塊的通道排序 [none|ordered|unordered]")
	c.Flags().Bool(flagRequireRegistration, false, "如果模塊無法註冊，如果 true 命令將失敗")
	c.Flags().StringSlice(flagParams, []string{}, "腳手架模塊參數")

	return c
}

func scaffoldModuleHandler(cmd *cobra.Command, args []string) error {
	var (
		name    = args[0]
		appPath = flagGetPath(cmd)
	)
	s := clispinner.New().SetText("創建中,請耐心等待...")
	defer s.Stop()

	ibcModule, err := cmd.Flags().GetBool(flagIBC)
	if err != nil {
		return err
	}

	ibcOrdering, err := cmd.Flags().GetString(flagIBCOrdering)
	if err != nil {
		return err
	}
	requireRegistration, err := cmd.Flags().GetBool(flagRequireRegistration)
	if err != nil {
		return err
	}

	params, err := cmd.Flags().GetStringSlice(flagParams)
	if err != nil {
		return err
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	options := []scaffolder.ModuleCreationOption{
		scaffolder.WithParams(params),
	}

	// 檢查模塊是否必須是 IBC 模塊
	if ibcModule {
		options = append(options, scaffolder.WithIBCChannelOrdering(ibcOrdering), scaffolder.WithIBC())
	}

	// 獲取模塊依賴
	dependencies, err := cmd.Flags().GetStringSlice(flagDep)
	if err != nil {
		return err
	}
	if len(dependencies) > 0 {
		var formattedDependencies []modulecreate.Dependency

		// 解析提供的依賴項
		for _, dependency := range dependencies {
			var formattedDependency modulecreate.Dependency

			splitted := strings.Split(dependency, ":")
			switch len(splitted) {
			case 1:
				formattedDependency = modulecreate.NewDependency(splitted[0], "")
			case 2:
				formattedDependency = modulecreate.NewDependency(splitted[0], splitted[1])
			default:
				return fmt.Errorf("依賴 %s 無效，必須有 <depName> or <depName>.<depKeeperName>", dependency)
			}
			formattedDependencies = append(formattedDependencies, formattedDependency)
		}
		options = append(options, scaffolder.WithDependencies(formattedDependencies))
	}

	var msg bytes.Buffer
	fmt.Fprintf(&msg, "\n🎉 創建好模塊 %s.\n\n", name)

	sc, err := newApp(appPath)
	if err != nil {
		return err
	}

	sm, err := sc.CreateModule(cacheStorage, placeholder.New(), name, options...)
	s.Stop()
	if err != nil {
		var validationErr validation.Error
		if !requireRegistration && errors.As(err, &validationErr) {
			fmt.Fprintf(&msg, "無法註冊模塊 '%s'.\n", name)
			fmt.Fprintln(&msg, validationErr.ValidationInfo())
		} else {
			return err
		}
	} else {
		modificationsStr, err := sourceModificationToString(sm)
		if err != nil {
			return err
		}

		fmt.Println(modificationsStr)
	}

	if len(dependencies) > 0 {
		dependencyWarning(dependencies)
	}

	io.Copy(cmd.OutOrStdout(), &msg)
	return nil
}

// 在之前搭建的應用程序中，gov keeper 定義在腳手架模塊 keeper 定義的下方
// 因此，如果是這種情況，我們必須警告用戶手動移動定義
// https://github.com/bearnetworkchain/core/issues/818#issuecomment-865736052
const govWarning = `⚠️ 如果您的應用程序是使用 Ignite CLI 0.16.x 或更低版本搭建的
請確保您的模塊管理員定義是在 gov 模塊管理員定義之後定義的 app/app.go:

app.GovKeeper = ...
...
[你的模塊管理員定義]
`

// 如果 gov 作為依賴項提供，dependencyWarning 用於打印警告
func dependencyWarning(dependencies []string) {
	for _, dep := range dependencies {
		if dep == "gov" {
			fmt.Print(govWarning)
		}
	}
}
