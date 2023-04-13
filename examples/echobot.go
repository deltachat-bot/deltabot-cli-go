package main

import (
	"fmt"

	"github.com/deltachat-bot/deltabot-cli-go/botcli"
	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
	"github.com/spf13/cobra"
)

var cli *botcli.BotCli

func main() {
	cli = botcli.New("echobot")

	// add an "info" CLI subcommand as example
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "display information about the Delta Chat core running in this system",
		Args:  cobra.ExactArgs(0),
	}
	cli.AddCommand(infoCmd, func(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		sysinfo, _ := bot.Account.Manager.SystemInfo()
		for key, val := range sysinfo {
			fmt.Printf("%v=%#v\n", key, val)
		}
	})

	// incoming message handling
	cli.OnBotInit(func(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		bot.OnNewMsg(func(msg *deltachat.Message) {
			snapshot, _ := msg.Snapshot()
			chat := deltachat.Chat{bot.Account, snapshot.ChatId}
			if snapshot.Text != "" { // ignore messages without text
				chat.SendText(snapshot.Text)
			}
		})
	})
	cli.OnBotStart(func(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		addr, _ := bot.GetConfig("configured_addr")
		cli.Logger.Infof("Listening at: %v", addr)
	})
	cli.Start()
}
