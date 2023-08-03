package main

import (
	"fmt"

	"github.com/deltachat-bot/deltabot-cli-go/botcli"
	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
	"github.com/spf13/cobra"
)

func main() {
	cli := botcli.New("echobot")

	// add an "info" CLI subcommand as example
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "display information about the Delta Chat core running in this system or about an specific account if one was selected with -a/--account",
		Args:  cobra.ExactArgs(0),
	}
	cli.AddCommand(infoCmd, func(cli *botcli.BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		var info map[string]string
		if cli.SelectedAddr == "" { // no account selected with --a/--account, show system info
			info, _ = bot.Rpc.GetSystemInfo()
		} else { // account selected, show info about that account
			accId, _ := cli.GetAccount(bot.Rpc, cli.SelectedAddr)
			info, _ = bot.Rpc.GetInfo(accId)
		}
		for key, val := range info {
			fmt.Printf("%v=%#v\n", key, val)
		}
	})

	// incoming message handling
	cli.OnBotInit(func(cli *botcli.BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		bot.OnNewMsg(func(bot *deltachat.Bot, accId deltachat.AccountId, msgId deltachat.MsgId) {
			msg, _ := bot.Rpc.GetMessage(accId, msgId)
			if msg.FromId > deltachat.ContactLastSpecial && msg.Text != "" {
				bot.Rpc.MiscSendTextMessage(accId, msg.ChatId, msg.Text)
			}
		})
	})
	cli.OnBotStart(func(cli *botcli.BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		cli.Logger.Info("OnBotStart event triggered: bot started!")
	})
	cli.Start()
}
