package ignitecmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/nodetime"
	"github.com/ignite-hq/cli/ignite/pkg/protoc"
)

// NewTools 返回一個命令，其中各種工具（二進製文件）作為子命令附加
// 對於高級用戶。
func NewTools() *cobra.Command {
	c := &cobra.Command{
		Use:   "tools",
		Short: "高級用戶工具",
	}
	c.AddCommand(NewToolsIBCSetup())
	c.AddCommand(NewToolsIBCRelayer())
	c.AddCommand(NewToolsProtoc())
	c.AddCommand(NewToolsCompletions())
	return c
}

func NewToolsIBCSetup() *cobra.Command {
	return &cobra.Command{
		Use:   "ibc-setup [--] [...]",
		Short: "快速設置中繼器的命令集合",
		RunE:  toolsNodetimeProxy(nodetime.CommandIBCSetup),
		Example: `ignite tools ibc-setup -- -h
ignite tools ibc-setup -- init --src relayer_test_1 --dest relayer_test_2`,
	}
}

func NewToolsIBCRelayer() *cobra.Command {
	return &cobra.Command{
		Use:     "ibc-relayer [--] [...]",
		Short:   "IBC 中繼器的打字稿實現",
		RunE:    toolsNodetimeProxy(nodetime.CommandIBCRelayer),
		Example: `ignite tools ibc-relayer -- -h`,
	}
}

func NewToolsProtoc() *cobra.Command {
	return &cobra.Command{
		Use:     "protoc [--] [...]",
		Short:   "執行協議命令",
		Long:    "協議命令。您不需要設置全局協議包含文件夾 -I, 它是自動處理的",
		RunE:    toolsProtocProxy,
		Example: `ignite tools protoc -- --version`,
	}
}

func toolsNodetimeProxy(c nodetime.CommandName) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		command, cleanup, err := nodetime.Command(c)
		if err != nil {
			return err
		}
		defer cleanup()

		return toolsProxy(cmd.Context(), append(command, args...))
	}
}

func toolsProtocProxy(cmd *cobra.Command, args []string) error {
	command, cleanup, err := protoc.Command()
	if err != nil {
		return err
	}
	defer cleanup()

	return toolsProxy(cmd.Context(), append(command.Command, args...))
}

func toolsProxy(ctx context.Context, command []string) error {
	return cmdrunner.New().Run(
		ctx,
		step.New(
			step.Exec(command[0], command[1:]...),
			step.Stdout(os.Stdout),
			step.Stderr(os.Stderr),
		),
	)
}

func NewToolsCompletions() *cobra.Command {

	// completionCmd 表示完成命令
	c := &cobra.Command{
		Use:   "completions",
		Short: "生成完成腳本",
		Long: ` 補全命令輸出一個補全腳本，你可以在你的 shell 中使用. 輸出腳本需要
那 [bash-completion](https://github.com/scop/bash-completion)已安裝並在您的
系統. 由於大多數類 Unix 操作系統默認帶有 bash-completion，因此 bash-completion
可能已經安裝並運行。

Bash:

  $ source <(ignite  tools completions bash)

  要為每個新會話加載完成，請運行:

  ** Linux **
  $ ignite  tools completions bash > /etc/bash_completion.d/ignite

  ** macOS **
  $ ignite  tools completions bash > /usr/local/etc/bash_completion.d/ignite

Zsh:

  如果您的環境中尚未啟用 shell 完成，則需要啟用它。您可以執行以下一次:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  要為每個會話加載完成，請執行一次:
  
  $ ignite  tools completions zsh > "${fpath[1]}/_ignite"

  您需要啟動一個新的 shell 才能使此設置生效.

fish:

  $ ignite  tools completions fish | source

 要為每個會話加載完成，請執行一次:
  
  $ ignite  tools completions fish > ~/.config/fish/completionss/ignite.fish

PowerShell:

  PS> ignite  tools completions powershell | Out-String | Invoke-Expression

  要為每個新會話加載完成，請運行:
  
  PS> ignite  tools completions powershell > ignite.ps1
  
  並從您的 PowerShell 配置文件中獲取此文件.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
	return c
}
