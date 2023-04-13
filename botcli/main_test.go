package botcli

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/deltachat/deltachat-rpc-client-go/acfactory"
	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
)

func RunConfiguredCli(cli *BotCli, args ...string) (output string, err error) {
	bot := acfactory.OnlineBot()
	acfactory.StopRpc(bot)
	dir := filepath.Dir(bot.Account.Manager.Rpc.(*deltachat.RpcIO).AccountsDir)
	args = append([]string{"-f=" + dir}, args...)
	return runCli(cli, args...)
}

func RunCli(cli *BotCli, args ...string) (output string, err error) {
	args = append([]string{"-f=" + acfactory.MkdirTemp()}, args...)
	return runCli(cli, args...)
}

func runCli(cli *BotCli, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	cli.RootCmd.SetOut(buf)
	cli.RootCmd.SetErr(buf)
	cli.RootCmd.SetArgs(args)

	err = cli.Start()
	return buf.String(), err
}

func TestMain(m *testing.M) {
	cfg := map[string]string{
		"mail_server":   "localhost",
		"send_server":   "localhost",
		"mail_port":     "3143",
		"send_port":     "3025",
		"mail_security": "3",
		"send_security": "3",
	}
	acfactory.TearUp(cfg)
	defer acfactory.TearDown()
	m.Run()
}
