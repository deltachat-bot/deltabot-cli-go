package botcli

import (
	"testing"

	"github.com/chatmail/rpc-client-go/deltachat"
	"github.com/chatmail/rpc-client-go/deltachat/option"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBotCli_SetConfig(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineBot(func(bot *deltachat.Bot, accId deltachat.AccountId) {
		cli := New("testbot")
		assert.Nil(t, cli.SetConfig(bot, accId, "testkey", option.Some("testing")))
		value, err := cli.GetConfig(bot, accId, "testkey")
		assert.Nil(t, err)
		assert.Equal(t, "testing", value.UnwrapOr(""))
	})
}

func TestBotCli_AdminChat(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineBot(func(bot *deltachat.Bot, accId deltachat.AccountId) {
		cli := New("testbot")
		chatId1, err := cli.AdminChat(bot, accId)
		assert.Nil(t, err)
		chatId2, err := cli.ResetAdminChat(bot, accId)
		assert.Nil(t, err)
		assert.NotEqual(t, chatId2, chatId1)

		isAdmin, err := cli.IsAdmin(bot, accId, deltachat.ContactSelf)
		assert.Nil(t, err)
		assert.True(t, isAdmin)
	})
}

func TestBotCli_AddCommand(t *testing.T) {
	t.Parallel()
	cli := New("testbot")
	var called bool
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "test subcommand",
		Args:  cobra.ExactArgs(0),
	}
	cli.AddCommand(testCmd, func(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
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
	cli.OnBotStart(func(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		cliBot = bot
	})
	go RunConfiguredCli(cli, "serve") //nolint:errcheck
	for cliBot == nil || !cliBot.IsRunning() {
	}
	cliBot.Stop()
}

func TestBotCli_serve(t *testing.T) {
	t.Parallel()
	cli := New("testbot")
	onNewMsgCalled := make(chan *deltachat.MsgSnapshot, 1)
	var cliBot *deltachat.Bot
	cli.OnBotInit(func(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		cliBot = bot
		bot.OnNewMsg(func(bot *deltachat.Bot, accId deltachat.AccountId, msgId deltachat.MsgId) {
			snapshot, _ := bot.Rpc.GetMessage(accId, msgId)
			select {
			case onNewMsgCalled <- snapshot:
			default:
			}
		})
	})
	go RunConfiguredCli(cli, "serve") //nolint:errcheck
	for cliBot == nil || !cliBot.IsRunning() {
	}
	defer cliBot.Stop()

	acfactory.WithOnlineAccount(func(rpc *deltachat.Rpc, accId deltachat.AccountId) {
		chatWithBot := acfactory.CreateChat(rpc, accId, cliBot.Rpc, 1)

		_, err := rpc.MiscSendTextMessage(accId, chatWithBot, "hi")
		assert.Nil(t, err)
		msg := <-onNewMsgCalled
		assert.Equal(t, "hi", msg.Text)
	})
}

func TestInitCallback(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *deltachat.Rpc, accId deltachat.AccountId) {
		addr, err := rpc.GetConfig(accId, "addr")
		assert.Nil(t, err)
		password, err := rpc.GetConfig(accId, "mail_pw")
		assert.Nil(t, err)
		err = rpc.SetConfig(accId, "mail_pw", option.None[string]())
		assert.Nil(t, err)
		configured, _ := rpc.IsConfigured(accId)
		assert.False(t, configured)

		cli := New("testbot")
		cli.OnBotInit(func(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
			bot.Rpc = rpc
		})
		_, err = RunCli(cli, "init", addr.Unwrap(), password.Unwrap())
		assert.Nil(t, err)

		configured, _ = rpc.IsConfigured(accId)
		assert.True(t, configured)
	})
}

func TestConfigCallback(t *testing.T) {
	t.Parallel()
	var err error
	cli := New("testbot")

	_, err = RunCli(cli, "config", "addr")
	assert.Nil(t, err)

	_, err = RunCli(cli, "config", "addr", "test@example.com")
	assert.Nil(t, err)
}

func TestQrCallback(t *testing.T) {
	t.Parallel()
	var err error
	cli := New("testbot")
	_, err = RunCli(cli, "qr")
	assert.Nil(t, err)

	_, err = RunConfiguredCli(cli, "qr")
	assert.Nil(t, err)

	_, err = RunConfiguredCli(cli, "qr", "-i")
	assert.Nil(t, err)
}

func TestAdminCallback(t *testing.T) {
	t.Parallel()
	var err error
	cli := New("testbot")
	_, err = RunCli(cli, "admin")
	assert.Nil(t, err)

	_, err = RunConfiguredCli(cli, "admin")
	assert.Nil(t, err)

	_, err = RunConfiguredCli(cli, "admin", "-i")
	assert.Nil(t, err)

	_, err = RunConfiguredCli(cli, "admin", "-r")
	assert.Nil(t, err)
}
