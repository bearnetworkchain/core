package ignitecmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/chaincmd"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/colors"
	"github.com/ignite-hq/cli/ignite/services/chain"
)

const (
	flagOutput         = "output"
	flagRelease        = "release"
	flagReleaseTargets = "release.targets"
	flagReleasePrefix  = "release.prefix"
)

// NewChainBuild returns a new build command to build a blockchain app.
func NewChainBuild() *cobra.Command {
	c := &cobra.Command{
		Use:   "build",
		Short: "Build a node binary",
		Long: `默認情況下，構建您的節點二進製文件並將二進製文件添加到您的 $(go env GOPATH)/bin 路徑。
要為發布構建二進製文件，請使用 --release 標誌。 
應用程序二進製文件為一個或多個指定的發布目標構建在應用程序下的 release/ 目錄中資源。 
使用 GOOS:GOARCH 構建標籤指定發布目標。
如果未指定可選的 --release.targets，則會為您當前的環境創建一個二進製文件。

示例用法：
	- ignite chain build
	- ignite chain build --release -t linux:amd64 -t darwin:amd64 -t darwin:arm64`,
		Args: cobra.NoArgs,
		RunE: chainBuildHandler,
	}

	flagSetPath(c)
	flagSetClearCache(c)
	c.Flags().AddFlagSet(flagSetHome())
	c.Flags().AddFlagSet(flagSetProto3rdParty("Available only without the --release flag"))
	c.Flags().Bool(flagRelease, false, "build for a release")
	c.Flags().StringSliceP(flagReleaseTargets, "t", []string{}, "release targets. Available only with --release flag")
	c.Flags().String(flagReleasePrefix, "", "tarball prefix for each release target. Available only with --release flag")
	c.Flags().StringP(flagOutput, "o", "", "binary output path")
	c.Flags().BoolP("verbose", "v", false, "Verbose output")

	return c
}

func chainBuildHandler(cmd *cobra.Command, _ []string) error {
	var (
		isRelease, _      = cmd.Flags().GetBool(flagRelease)
		releaseTargets, _ = cmd.Flags().GetStringSlice(flagReleaseTargets)
		releasePrefix, _  = cmd.Flags().GetString(flagReleasePrefix)
		output, _         = cmd.Flags().GetString(flagOutput)
	)

	chainOption := []chain.Option{
		chain.LogLevel(logLevel(cmd)),
		chain.KeyringBackend(chaincmd.KeyringBackendTest),
	}

	if flagGetProto3rdParty(cmd) {
		chainOption = append(chainOption, chain.EnableThirdPartyModuleCodegen())
	}

	c, err := newChainWithHomeFlags(cmd, chainOption...)
	if err != nil {
		return err
	}

	cacheStorage, err := newCache(cmd)
	if err != nil {
		return err
	}

	if isRelease {
		releasePath, err := c.BuildRelease(cmd.Context(), cacheStorage, output, releasePrefix, releaseTargets...)
		if err != nil {
			return err
		}

		fmt.Printf("🗃  Release created: %s\n", colors.Info(releasePath))

		return nil
	}

	binaryName, err := c.Build(cmd.Context(), cacheStorage, output)
	if err != nil {
		return err
	}

	if output == "" {
		fmt.Printf("🗃  Installed. Use with: %s\n", colors.Info(binaryName))
	} else {
		binaryPath := filepath.Join(output, binaryName)
		fmt.Printf("🗃  Binary built at the path: %s\n", colors.Info(binaryPath))
	}

	return nil
}
