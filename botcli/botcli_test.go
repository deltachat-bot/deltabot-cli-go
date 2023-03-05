package botcli

import (
	"fmt"
	"testing"

	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
	"github.com/spf13/cobra"
)

var cli *BotCli

func logEvent(event *deltachat.Event) {
	switch event.Type {
	case deltachat.EVENT_INFO:
		cli.Logger.Info().Msg(event.Msg)
	case deltachat.EVENT_WARNING:
		cli.Logger.Warn().Msg(event.Msg)
	case deltachat.EVENT_ERROR:
		cli.Logger.Error().Msg(event.Msg)
	}
}

func TestBasic(t *testing.T) {
	cli = New("echobot")

	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "display information about the Delta Chat core running in this system",
		Args:  cobra.ExactArgs(0),
	}
	cli.AddCommand(infoCmd, func(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		sysinfo, _ := bot.Account.Manager.SystemInfo()
		for key, val := range sysinfo {
			fmt.Printf("%v=%q\n", key, val)
		}
	})

	cli.OnBotInit(func(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		bot.On(deltachat.EVENT_INFO, logEvent)
		bot.On(deltachat.EVENT_WARNING, logEvent)
		bot.On(deltachat.EVENT_ERROR, logEvent)
		bot.OnNewMsg(func(msg *deltachat.Message) {
			snapshot, _ := msg.Snapshot()
			chat := deltachat.Chat{bot.Account, snapshot.ChatId}
			chat.SendText(snapshot.Text)
		})
	})
	cli.OnBotStart(func(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		addr, _ := bot.GetConfig("addr")
		cli.Logger.Info().Msgf("Listening at: %v", addr)
	})
	cli.Start()
}
