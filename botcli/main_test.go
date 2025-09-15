package botcli

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/chatmail/rpc-client-go/deltachat"
	"github.com/chatmail/rpc-client-go/deltachat/transport"
)

var acfactory *deltachat.AcFactory

func TestMain(m *testing.M) {
	acfactory = &deltachat.AcFactory{}
	acfactory.TearUp()
	defer acfactory.TearDown()
	m.Run()
}

func RunConfiguredCli(cli *BotCli, args ...string) (output string, err error) {
	var dir string
	acfactory.WithOnlineBot(func(bot *deltachat.Bot, accId deltachat.AccountId) {
		dir = filepath.Dir(bot.Rpc.Transport.(*transport.IOTransport).AccountsDir)
	})
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
