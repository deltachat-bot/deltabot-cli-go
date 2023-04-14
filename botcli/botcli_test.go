package botcli

import (
	"testing"

	"github.com/deltachat/deltachat-rpc-client-go/acfactory"
	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBotCli_AddCommand(t *testing.T) {
	t.Parallel()
	cli := New("testbot")
	var called bool
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "test subcommand",
		Args:  cobra.ExactArgs(0),
	}
	cli.AddCommand(testCmd, func(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		called = true
	})
	_, err := RunCli(cli, "test")
	assert.Nil(t, err)
	assert.True(t, called)
}

func TestBotCli_OnBotStart(t *testing.T) {
	t.Parallel()
	cli := New("testbot")
	var cliBot *deltachat.Bot
	cli.OnBotStart(func(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		cliBot = bot
	})
	go RunConfiguredCli(cli, "serve") //nolint:errcheck
	for {
		if cliBot != nil && cliBot.IsRunning() {
			break
		}
	}
	cliBot.Stop()
}

func TestBotCli_OnBotInit(t *testing.T) {
	t.Parallel()
	cli := New("testbot")
	onEventInfoCalled := make(chan deltachat.Event, 1)
	onNewMsgCalled := make(chan *deltachat.MsgSnapshot, 1)
	var cliBot *deltachat.Bot
	cli.OnBotInit(func(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		cliBot = bot
		bot.On(deltachat.EventInfo{}, func(event deltachat.Event) {
			select {
			case onEventInfoCalled <- event:
			default:
			}
		})
		bot.OnNewMsg(func(msg *deltachat.Message) {
			snapshot, _ := msg.Snapshot()
			select {
			case onNewMsgCalled <- snapshot:
			default:
			}
		})
	})
	go RunConfiguredCli(cli, "serve") //nolint:errcheck
	for {
		if cliBot != nil && cliBot.IsRunning() {
			break
		}
	}
	defer cliBot.Stop()

	user := acfactory.OnlineAccount()
	defer acfactory.StopRpc(user)

	assert.IsType(t, deltachat.EventInfo{}, <-onEventInfoCalled)

	chatWithBot, err := acfactory.CreateChat(user, cliBot.Account)
	assert.Nil(t, err)

	_, err = chatWithBot.SendText("hi")
	assert.Nil(t, err)
	msg := <-onNewMsgCalled
	assert.Equal(t, "hi", msg.Text)
}

func TestBotCli_initCallback(t *testing.T) {
	t.Parallel()
	acc := acfactory.UnconfiguredAccount()
	defer acfactory.StopRpc(acc)

	addr, err := acc.GetConfig("addr")
	assert.Nil(t, err)
	err = acc.SetConfig("addr", "")
	assert.Nil(t, err)
	password, err := acc.GetConfig("mail_pw")
	assert.Nil(t, err)
	err = acc.SetConfig("mail_pw", "")
	assert.Nil(t, err)

	cli := New("testbot")
	cli.OnBotInit(func(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		bot.Account = acc
	})

	configured, _ := acc.IsConfigured()
	assert.False(t, configured)

	_, err = RunCli(cli, "init", addr, password)
	assert.Nil(t, err)

	configured, _ = acc.IsConfigured()
	assert.True(t, configured)
}

func TestBotCli_configCallback(t *testing.T) {
	t.Parallel()
	var err error
	var cliBot *deltachat.Bot
	cli := New("testbot")
	cli.OnBotInit(func(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		cliBot = bot
	})

	_, err = RunCli(cli, "config", "addr")
	assert.Nil(t, err)

	_, err = RunCli(cli, "config", "addr", "test@example.com")
	assert.Nil(t, err)

	assert.Nil(t, cliBot.Account.Manager.Rpc.Start())
	defer acfactory.StopRpc(cliBot)

	addr, err := cliBot.GetConfig("addr")
	assert.Nil(t, err)
	assert.Equal(t, "test@example.com", addr)
}

func TestBotCli_qrCallback(t *testing.T) {
	t.Parallel()
	var err error
	cli := New("testbot")
	_, err = RunCli(cli, "qr")
	assert.Nil(t, err)

	_, err = RunConfiguredCli(cli, "qr")
	assert.Nil(t, err)
}

func TestBotCli_SetConfig(t *testing.T) {
	t.Parallel()
	bot := acfactory.OnlineBot()
	defer acfactory.StopRpc(bot)

	cli := New("testbot")
	assert.Nil(t, cli.SetConfig(bot, "testkey", "testing"))
	value, err := cli.GetConfig(bot, "testkey")
	assert.Nil(t, err)
	assert.Equal(t, "testing", value)
}
