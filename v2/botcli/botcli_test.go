package botcli

import (
	"fmt"
	"testing"

	"github.com/chatmail/rpc-client-go/v2/deltachat"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestBotCli_SetConfig(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineBot(func(bot *deltachat.Bot, accId uint32) {
		cli := New("testbot")
		testVal := "testing"
		require.Nil(t, cli.SetConfig(bot, accId, "testkey", &testVal))
		value, err := cli.GetConfig(bot, accId, "testkey")
		require.Nil(t, err)
		require.NotNil(t, value)
		require.Equal(t, testVal, *value)
	})
}

func TestBotCli_AdminChat(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineBot(func(bot *deltachat.Bot, accId uint32) {
		cli := New("testbot")
		chatId1, err := cli.AdminChat(bot, accId)
		require.Nil(t, err)
		chatId2, err := cli.ResetAdminChat(bot, accId)
		require.Nil(t, err)
		require.NotEqual(t, chatId2, chatId1)

		isAdmin, err := cli.IsAdmin(bot, accId, deltachat.ContactSelf)
		require.Nil(t, err)
		require.True(t, isAdmin)
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
	require.Nil(t, err)
	require.True(t, called)
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
	onNewMsgCalled := make(chan *deltachat.Message, 1)
	var cliBot *deltachat.Bot
	cli.OnBotInit(func(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		cliBot = bot
		bot.OnNewMsg(func(bot *deltachat.Bot, accId uint32, msgId uint32) {
			snapshot, _ := bot.Rpc.GetMessage(accId, msgId)
			select {
			case onNewMsgCalled <- &snapshot:
			default:
			}
		})
	})
	go RunConfiguredCli(cli, "serve") //nolint:errcheck
	for cliBot == nil || !cliBot.IsRunning() {
	}
	defer cliBot.Stop()

	acfactory.WithOnlineAccount(func(rpc *deltachat.Rpc, accId uint32) {
		chatWithBot := acfactory.CreateChat(rpc, accId, cliBot.Rpc, 1)

		_, err := rpc.MiscSendTextMessage(accId, chatWithBot, "hi")
		require.Nil(t, err)
		msg := <-onNewMsgCalled
		require.Equal(t, "hi", msg.Text)
	})
}

func TestInitCallback(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *deltachat.Rpc, accId uint32) {
		configured, _ := rpc.IsConfigured(accId)
		require.False(t, configured)

		cli := New("testbot")
		cli.OnBotInit(func(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
			bot.Rpc = rpc
		})
		_, err := RunCli(cli, "init", fmt.Sprintf("-a=%v", accId), acfactory.ConfigQr)
		require.Nil(t, err)

		configured, _ = rpc.IsConfigured(accId)
		require.True(t, configured)
	})
}

func TestConfigCallback(t *testing.T) {
	t.Parallel()
	var err error
	cli := New("testbot")

	_, err = RunCli(cli, "config", "addr")
	require.Nil(t, err)

	_, err = RunCli(cli, "config", "addr", "test@example.com")
	require.Nil(t, err)
}

func TestQrCallback(t *testing.T) {
	t.Parallel()
	var err error
	cli := New("testbot")
	_, err = RunCli(cli, "link")
	require.Nil(t, err)

	_, err = RunConfiguredCli(cli, "link")
	require.Nil(t, err)
}

func TestAdminCallback(t *testing.T) {
	t.Parallel()
	var err error
	cli := New("testbot")
	_, err = RunCli(cli, "admin")
	require.Nil(t, err)

	_, err = RunConfiguredCli(cli, "admin")
	require.Nil(t, err)

	_, err = RunConfiguredCli(cli, "admin", "-r")
	require.Nil(t, err)
}
