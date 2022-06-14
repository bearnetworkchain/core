package ignitecmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"

	"github.com/ignite-hq/cli/ignite/pkg/cliui/cliquiz"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosaccount"
)

const flagSecret = "secret"

func NewAccountImport() *cobra.Command {
	c := &cobra.Command{
		Use:   "import [name]",
		Short: "使用助記詞或私鑰導入賬戶",
		Args:  cobra.ExactArgs(1),
		RunE:  accountImportHandler,
	}

	c.Flags().String(flagSecret, "", "您的助記詞或私鑰的路徑（使用交互模式來安全地傳遞您的助記詞)")
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	c.Flags().AddFlagSet(flagSetAccountImportExport())

	return c
}

func accountImportHandler(cmd *cobra.Command, args []string) error {
	var (
		name      = args[0]
		secret, _ = cmd.Flags().GetString(flagSecret)
	)

	if secret == "" {
		if err := cliquiz.Ask(
			cliquiz.NewQuestion("您的助記符或私鑰路徑", &secret, cliquiz.Required())); err != nil {
			return err
		}
	}

	passphrase, err := getPassphrase(cmd)
	if err != nil {
		return err
	}

	if !bip39.IsMnemonicValid(secret) {
		privKey, err := os.ReadFile(secret)
		if os.IsNotExist(err) {
			return errors.New("助記符無效或在路徑中找不到私鑰")
		}
		if err != nil {
			return err
		}
		secret = string(privKey)
	}

	ca, err := cosmosaccount.New(
		cosmosaccount.WithKeyringBackend(getKeyringBackend(cmd)),
	)
	if err != nil {
		return err
	}

	if _, err := ca.Import(name, secret, passphrase); err != nil {
		return err
	}

	fmt.Printf("帳戶 %q 已匯入.\n", name)
	return nil
}
