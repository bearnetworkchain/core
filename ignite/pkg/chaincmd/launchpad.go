package chaincmd

import "github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"

const (
	commandConfig     = "config"
	commandRestServer = "rest-server"

	optionUnsafeCors = "--unsafe-cors"
	optionAPIAddress = "--laddr"
	optionRPCAddress = "--node"
	optionName       = "--name"
)

// 啟動板設置配置命令
func (c ChainCmd) launchpadSetConfigCommand(name string, value string) step.Option {
	command := []string{
		commandConfig,
		name,
		value,
	}

	return c.cliCommand(command)
}

// 啟動板RestServerCommand
func (c ChainCmd) launchpadRestServerCommand(apiAddress string, rpcAddress string) step.Option {
	command := []string{
		commandRestServer,
		optionUnsafeCors,
		optionAPIAddress,
		apiAddress,
		optionRPCAddress,
		rpcAddress,
	}
	return c.cliCommand(command)
}

// attachCLIHome 將 home 標誌附加到提供的 CLI 命令
func (c ChainCmd) attachCLIHome(command []string) []string {
	if c.cliHome != "" {
		command = append(command, []string{optionHome, c.cliHome}...)
	}
	return command
}
