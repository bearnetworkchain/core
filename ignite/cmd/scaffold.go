package ignitecmd

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/bearnetworkchain/core/ignite/pkg/cliui/clispinner"
	"github.com/bearnetworkchain/core/ignite/pkg/placeholder"
	"github.com/bearnetworkchain/core/ignite/pkg/xgit"
	"github.com/bearnetworkchain/core/ignite/services/scaffolder"
)

// 與組件腳手架相關的標誌
const (
	flagModule       = "module"
	flagNoMessage    = "no-message"
	flagNoSimulation = "no-simulation"
	flagResponse     = "response"
	flagDescription  = "desc"
)

// NewScaffold 返回一個命令，該命令對與腳手架相關的子命令進行分組。
func NewScaffold() *cobra.Command {
	c := &cobra.Command{
		Use:   "scaffold [command]",
		Short: "搭建新的區塊鏈、模塊、消息、查詢等",
		Long: `腳手架命令創建和修改源代碼文件以添加功能.

CRUD代表“創建、讀取、更新、刪除”.`,
		Aliases: []string{"s"},
		Args:    cobra.ExactArgs(1),
	}

	c.AddCommand(NewScaffoldChain())
	c.AddCommand(addGitChangesVerifier(NewScaffoldModule()))
	c.AddCommand(addGitChangesVerifier(NewScaffoldList()))
	c.AddCommand(addGitChangesVerifier(NewScaffoldMap()))
	c.AddCommand(addGitChangesVerifier(NewScaffoldSingle()))
	c.AddCommand(addGitChangesVerifier(NewScaffoldType()))
	c.AddCommand(addGitChangesVerifier(NewScaffoldMessage()))
	c.AddCommand(addGitChangesVerifier(NewScaffoldQuery()))
	c.AddCommand(addGitChangesVerifier(NewScaffoldPacket()))
	c.AddCommand(addGitChangesVerifier(NewScaffoldBandchain()))
	c.AddCommand(addGitChangesVerifier(NewScaffoldVue()))
	c.AddCommand(addGitChangesVerifier(NewScaffoldFlutter()))
	// c.AddCommand(NewScaffoldWasm())

	return c
}

func scaffoldType(
	cmd *cobra.Command,
	args []string,
	kind scaffolder.AddTypeKind,
) error {
	var (
		typeName          = args[0]
		fields            = args[1:]
		moduleName        = flagGetModule(cmd)
		withoutMessage    = flagGetNoMessage(cmd)
		withoutSimulation = flagGetNoSimulation(cmd)
		signer            = flagGetSigner(cmd)
		appPath           = flagGetPath(cmd)
	)

	var options []scaffolder.AddTypeOption

	if len(fields) > 0 {
		options = append(options, scaffolder.TypeWithFields(fields...))
	}
	if moduleName != "" {
		options = append(options, scaffolder.TypeWithModule(moduleName))
	}
	if withoutMessage {
		options = append(options, scaffolder.TypeWithoutMessage())
	} else {
		if signer != "" {
			options = append(options, scaffolder.TypeWithSigner(signer))
		}
		if withoutSimulation {
			options = append(options, scaffolder.TypeWithoutSimulation())
		}
	}

	s := clispinner.New().SetText("努力創建中...")
	defer s.Stop()

	sc, err := newApp(appPath)
	if err != nil {
		return err
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	sm, err := sc.AddType(cmd.Context(), cacheStorage, typeName, placeholder.New(), kind, options...)
	if err != nil {
		return err
	}

	s.Stop()

	modificationsStr, err := sourceModificationToString(sm)
	if err != nil {
		return err
	}

	fmt.Println(modificationsStr)
	fmt.Printf("\n🎉 %s 添加. \n\n", typeName)

	return nil
}

func addGitChangesVerifier(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().AddFlagSet(flagSetYes())

	preRunFun := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if preRunFun != nil {
			if err := preRunFun(cmd, args); err != nil {
				return err
			}
		}

		appPath := flagGetPath(cmd)

		changesCommitted, err := xgit.AreChangesCommitted(appPath)
		if err != nil {
			return err
		}

		if !getYes(cmd) && !changesCommitted {
			var confirmed bool
			prompt := &survey.Confirm{
				Message: "您保存的項目更改尚未提交。要啟用恢復到當前狀態，請提交您保存的更改。是否要在不提交已保存更改的情況下繼續搭建腳手架",
			}
			if err := survey.AskOne(prompt, &confirmed); err != nil || !confirmed {
				return errors.New("said no")
			}
		}
		return nil
	}
	return cmd
}

func flagSetScaffoldType() *flag.FlagSet {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.String(flagModule, "", "要添加到的模塊。默認是應用程序的主模塊")
	f.Bool(flagNoMessage, false, "禁用 CRUD 交互消息腳手架")
	f.Bool(flagNoSimulation, false, "禁用 CRUD 模擬腳手架")
	f.String(flagSigner, "", "消息簽名者的標籤（默認：創建者）")
	return f
}

func flagGetModule(cmd *cobra.Command) string {
	module, _ := cmd.Flags().GetString(flagModule)
	return module
}

func flagGetNoSimulation(cmd *cobra.Command) bool {
	noMessage, _ := cmd.Flags().GetBool(flagNoSimulation)
	return noMessage
}

func flagGetNoMessage(cmd *cobra.Command) bool {
	noMessage, _ := cmd.Flags().GetBool(flagNoMessage)
	return noMessage
}

func flagGetSigner(cmd *cobra.Command) string {
	signer, _ := cmd.Flags().GetString(flagSigner)
	return signer
}
