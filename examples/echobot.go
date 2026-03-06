package main

import (
	"github.com/chatmail/rpc-client-go/v2/deltachat"
	"github.com/deltachat-bot/deltabot-cli-go/botcli"
	"github.com/spf13/cobra"
)

func main() {
	cli := botcli.New("echobot")

	cli.OnBotInit(func(cli *botcli.BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		// incoming message handling
		bot.OnNewMsg(func(bot *deltachat.Bot, accId uint32, msgId uint32) {
			msg, _ := bot.Rpc.GetMessage(accId, msgId)
			if msg.FromId > deltachat.ContactLastSpecial && msg.Text != "" {
				bot.Rpc.SendMsg(accId, msg.ChatId, deltachat.MessageData{Text: &msg.Text})
			}
		})
	})
	cli.OnBotStart(func(cli *botcli.BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		cli.Logger.Info("OnBotStart event triggered: bot is about to start!")
	})

	cli.Start()
}
