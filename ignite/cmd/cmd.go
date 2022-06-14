package ignitecmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/ignite-hq/cli/ignite/chainconfig"
	"github.com/ignite-hq/cli/ignite/pkg/cache"
	"github.com/ignite-hq/cli/ignite/pkg/cliui"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosver"
	"github.com/ignite-hq/cli/ignite/pkg/gitpod"
	"github.com/ignite-hq/cli/ignite/pkg/goenv"
	"github.com/ignite-hq/cli/ignite/pkg/xgenny"
	"github.com/ignite-hq/cli/ignite/services/chain"
	"github.com/ignite-hq/cli/ignite/services/scaffolder"
	"github.com/ignite-hq/cli/ignite/version"
)

const (
	flagPath          = "path"
	flagHome          = "home"
	flagProto3rdParty = "proto-all-modules"
	flagYes           = "yes"
	flagClearCache    = "clear-cache"

	checkVersionTimeout = time.Millisecond * 600
	cacheFileName       = "ignite_cache.db"
)

// New creates a new root command for `Ignite CLI` with its sub commands.
func New() *cobra.Command {
	cobra.EnableCommandSorting = false

	c := &cobra.Command{
		Use:   "ignite",
		Short: "Ignite CLI 提供了搭建、測試、構建和啟動區塊鏈所需的一切",
		Long: `Ignite CLI 是一個使用 Cosmos SDK 創建主權區塊鏈的工具最流行的模塊化區塊鏈框架。 
		Ignite CLI 提供了搭建、測試、構建和啟動區塊鏈所需的一切。

		首先，創建一個區塊鏈：

ignite scaffold chain bnk`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check for new versions only when shell completion scripts are not being
			// generated to avoid invalid output to stdout when a new version is available
			if cmd.Use != "completions" {
				checkNewVersion(cmd.Context())
			}

			return goenv.ConfigurePath()
		},
	}

	c.AddCommand(NewScaffold())
	c.AddCommand(NewChain())
	c.AddCommand(NewGenerate())
	c.AddCommand(NewNetwork())
	c.AddCommand(NewAccount())
	c.AddCommand(NewRelayer())
	c.AddCommand(NewTools())
	c.AddCommand(NewDocs())
	c.AddCommand(NewVersion())
	c.AddCommand(deprecated()...)

	return c
}

func logLevel(cmd *cobra.Command) chain.LogLvl {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		return chain.LogVerbose
	}
	return chain.LogRegular
}

func flagSetPath(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(flagPath, "p", ".", "應用程序路徑")
}

func flagGetPath(cmd *cobra.Command) (path string) {
	path, _ = cmd.Flags().GetString(flagPath)
	return
}

func flagSetHome() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(flagHome, "", "用於區塊鏈的主目錄")
	return fs
}

func flagNetworkFrom() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(flagFrom, cosmosaccount.DefaultAccount, "用於向 SPN 發送交易的帳戶名稱")
	return fs
}

func getHome(cmd *cobra.Command) (home string) {
	home, _ = cmd.Flags().GetString(flagHome)
	return
}

func flagSetYes() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.BoolP(flagYes, "y", false, "用交互式回答 是/否 問題")
	return fs
}

func getYes(cmd *cobra.Command) (ok bool) {
	ok, _ = cmd.Flags().GetBool(flagYes)
	return
}

func flagSetProto3rdParty(additionalInfo string) *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	info := "為您的鏈中使用第三方模塊,啟用原型代碼生成"
	if additionalInfo != "" {
		info += ". " + additionalInfo
	}

	fs.Bool(flagProto3rdParty, false, info)
	return fs
}

func flagGetProto3rdParty(cmd *cobra.Command) bool {
	isEnabled, _ := cmd.Flags().GetBool(flagProto3rdParty)
	return isEnabled
}

func flagSetClearCache(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(flagClearCache, false, "清除構建緩存（高級）")
}

func flagGetClearCache(cmd *cobra.Command) bool {
	clearCache, _ := cmd.Flags().GetBool(flagClearCache)
	return clearCache
}

func newChainWithHomeFlags(cmd *cobra.Command, chainOption ...chain.Option) (*chain.Chain, error) {
	// Check if custom home is provided
	if home := getHome(cmd); home != "" {
		chainOption = append(chainOption, chain.HomePath(home))
	}

	appPath := flagGetPath(cmd)
	absPath, err := filepath.Abs(appPath)
	if err != nil {
		return nil, err
	}

	return chain.New(absPath, chainOption...)
}

var (
	modifyPrefix = color.New(color.FgMagenta).SprintFunc()("modify ")
	createPrefix = color.New(color.FgGreen).SprintFunc()("create ")
	removePrefix = func(s string) string {
		return strings.TrimPrefix(strings.TrimPrefix(s, modifyPrefix), createPrefix)
	}
)

func sourceModificationToString(sm xgenny.SourceModification) (string, error) {
	// get file names and add prefix
	var files []string
	for _, modified := range sm.ModifiedFiles() {
		// get the relative app path from the current directory
		relativePath, err := relativePath(modified)
		if err != nil {
			return "", err
		}
		files = append(files, modifyPrefix+relativePath)
	}
	for _, created := range sm.CreatedFiles() {
		// get the relative app path from the current directory
		relativePath, err := relativePath(created)
		if err != nil {
			return "", err
		}
		files = append(files, createPrefix+relativePath)
	}

	// sort filenames without prefix
	sort.Slice(files, func(i, j int) bool {
		s1 := removePrefix(files[i])
		s2 := removePrefix(files[j])

		return strings.Compare(s1, s2) == -1
	})

	return "\n" + strings.Join(files, "\n"), nil
}

func deprecated() []*cobra.Command {
	return []*cobra.Command{
		{
			Use:        "app",
			Deprecated: "使用 `ignite scaffold chain`.",
		},
		{
			Use:        "build",
			Deprecated: "使用 `ignite chain build`.",
		},
		{
			Use:        "serve",
			Deprecated: "使用 `ignite chain serve`.",
		},
		{
			Use:        "faucet",
			Deprecated: "使用 `ignite chain faucet`.",
		},
	}
}

// relativePath return the relative app path from the current directory
func relativePath(appPath string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	path, err := filepath.Rel(pwd, appPath)
	if err != nil {
		return "", err
	}
	return path, nil
}

func checkNewVersion(ctx context.Context) {
	if gitpod.IsOnGitpod() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, checkVersionTimeout)
	defer cancel()

	isAvailable, next, err := version.CheckNext(ctx)
	if err != nil || !isAvailable {
		return
	}

	fmt.Printf(`·
· 🛸 Ignite CLI %s is available!
·
· 要升級您的 Ignite CLI 版本，請參閱升級文檔: https://docs.ignite.com/guide/install.html#upgrading-your-ignite-cli-installation
·
··

`, next)
}

// newApp create a new scaffold app
func newApp(appPath string) (scaffolder.Scaffolder, error) {
	sc, err := scaffolder.App(appPath)
	if err != nil {
		return sc, err
	}

	if sc.Version.LT(cosmosver.StargateFortyFourVersion) {
		return sc, fmt.Errorf(
			`⚠️ 你的鏈已經使用舊版本的 Cosmos SDK 搭建了腳手架: %[1]v.
			請按照遷移指南將您的鏈升級到最新版本:

https://docs.ignite.com/migration`, sc.Version.String(),
		)
	}
	return sc, nil
}

func printSection(session cliui.Session, title string) error {
	return session.Printf("------\n%s\n------\n\n", title)
}

func newCache(cmd *cobra.Command) (cache.Storage, error) {
	cacheRootDir, err := chainconfig.ConfigDirPath()
	if err != nil {
		return cache.Storage{}, err
	}

	storage, err := cache.NewStorage(filepath.Join(cacheRootDir, cacheFileName))
	if err != nil {
		return cache.Storage{}, err
	}

	if flagGetClearCache(cmd) {
		if err := storage.Clear(); err != nil {
			return cache.Storage{}, err
		}
	}

	return storage, nil
}
