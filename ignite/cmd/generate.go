package ignitecmd

import "github.com/spf13/cobra"

// NewGenerate 返回一個將代碼生成相關子命令分組的命令。
func NewGenerate() *cobra.Command {
	c := &cobra.Command{
		Use:   "generate [command]",
		Short: "從源代碼生成客戶端、API文檔",
		Long: `從源代碼生成客戶端、API文檔.

		例如將協議緩衝區文件編譯成Golang語言或實現特定功能，例如生成 OpenAPI 規範。

		生成的源代碼可以通過再次運行命令重新生成，而不是手動編輯.`,
		Aliases: []string{"g"},
		Args:    cobra.ExactArgs(1),
	}

	flagSetPath(c)
	flagSetClearCache(c)
	c.AddCommand(addGitChangesVerifier(NewGenerateGo()))
	c.AddCommand(addGitChangesVerifier(NewGenerateVuex()))
	c.AddCommand(addGitChangesVerifier(NewGenerateDart()))
	c.AddCommand(addGitChangesVerifier(NewGenerateOpenAPI()))

	return c
}
