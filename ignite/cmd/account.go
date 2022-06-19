package ignitecmd

import (
	"os"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/ignite-hq/cli/ignite/pkg/cliui/cliquiz"
	"github.com/ignite-hq/cli/ignite/pkg/cliui/entrywriter"
	"github.com/ignite-hq/cli/ignite/pkg/cosmosaccount"
)

const (
	flagAddressPrefix  = "address-prefix"
	flagPassphrase     = "passphrase"
	flagNonInteractive = "non-interactive"
	flagKeyringBackend = "keyring-backend"
	flagFrom           = "from"
)

func NewAccount() *cobra.Command {
	c := &cobra.Command{
		Use:   "account [command]",
		Short: "用於管理帳戶的命令",
		Long: `用於管理帳戶的命令。 帳戶是一對私鑰和公鑰。
		Ignite CLI使用賬戶與Ignite區塊鏈交互，使用IBC中繼器等.`,
		Aliases: []string{"a"},
		Args:    cobra.ExactArgs(1),
	}

	c.AddCommand(NewAccountCreate())
	c.AddCommand(NewAccountDelete())
	c.AddCommand(NewAccountShow())
	c.AddCommand(NewAccountList())
	c.AddCommand(NewAccountImport())
	c.AddCommand(NewAccountExport())

	return c
}

func printAccounts(cmd *cobra.Command, accounts ...cosmosaccount.Account) error {
	var accEntries [][]string
	for _, acc := range accounts {
		accEntries = append(accEntries, []string{acc.Name, acc.Address(getAddressPrefix(cmd)), acc.PubKey()})
	}
	return entrywriter.MustWrite(os.Stdout, []string{"name", "address", "public key"}, accEntries...)
}

func flagSetKeyringBackend() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(flagKeyringBackend, "test", "密鑰環後端存儲您的帳戶密鑰")
	return fs
}

func getKeyringBackend(cmd *cobra.Command) cosmosaccount.KeyringBackend {
	backend, _ := cmd.Flags().GetString(flagKeyringBackend)
	return cosmosaccount.KeyringBackend(backend)
}

func flagSetAccountPrefixes() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(flagAddressPrefix, cosmosaccount.AccountPrefixCosmos, "Account address prefix")
	return fs
}

func getAddressPrefix(cmd *cobra.Command) string {
	prefix, _ := cmd.Flags().GetString(flagAddressPrefix)
	return prefix
}

func getFrom(cmd *cobra.Command) string {
	prefix, _ := cmd.Flags().GetString(flagFrom)
	return prefix
}

func flagSetAccountImportExport() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Bool(flagNonInteractive, false, "不要進入交互模式")
	fs.String(flagPassphrase, "", "Account passphrase")
	return fs
}

func getIsNonInteractive(cmd *cobra.Command) bool {
	is, _ := cmd.Flags().GetBool(flagNonInteractive)
	return is
}

func getPassphrase(cmd *cobra.Command) (string, error) {
	pass, _ := cmd.Flags().GetString(flagPassphrase)

	if pass == "" && !getIsNonInteractive(cmd) {
		if err := cliquiz.Ask(
			cliquiz.NewQuestion("Passphrase",
				&pass,
				cliquiz.HideAnswer(),
				cliquiz.GetConfirmation(),
			)); err != nil {
			return "", err
		}
	}

	return pass, nil
}
