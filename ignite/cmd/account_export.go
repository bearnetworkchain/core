package ignitecmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bearnetworkchain/core/ignite/pkg/cosmosaccount"
)

func NewAccountExport() *cobra.Command {
	c := &cobra.Command{
		Use:   "export [name]",
		Short: "導出帳戶私鑰",
		Args:  cobra.ExactArgs(1),
		RunE:  accountExportHandler,
	}

	c.Flags().AddFlagSet(flagSetKeyringBackend())
	c.Flags().AddFlagSet(flagSetAccountImportExport())
	c.Flags().String(flagPath, "", "導出私鑰的路徑。 默認 : ./key_[name]")

	return c
}

func accountExportHandler(cmd *cobra.Command, args []string) error {
	var (
		name = args[0]
		path = flagGetPath(cmd)
	)

	passphrase, err := getPassphrase(cmd)
	if err != nil {
		return err
	}

	ca, err := cosmosaccount.New(
		cosmosaccount.WithKeyringBackend(getKeyringBackend(cmd)),
	)
	if err != nil {
		return err
	}

	armored, err := ca.Export(name, passphrase)
	if err != nil {
		return err
	}

	if path == "" {
		path = fmt.Sprintf("./key_%s", name)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(armored), 0644); err != nil {
		return err
	}

	fmt.Printf("帳戶 %q 導出到文件: %s\n", name, path)
	return nil
}
