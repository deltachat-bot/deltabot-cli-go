// This example demonstrates how to create bots that have administrators.
//
// The bot has the /info command that can only be executed by bot administrators in the admins chat.
// To become admin you must use the `admin` subcommand in the cli, and scan the QR that will be shown.
package main

import (
	"github.com/deltachat-bot/deltabot-cli-go/botcli"
	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
	"github.com/spf13/cobra"
)

var cli *botcli.BotCli = botcli.New("infobot")

// Process messages sent to the group of administrators and allow to run privileged commands there.
func onNewMsg(bot *deltachat.Bot, accId deltachat.AccountId, msgId deltachat.MsgId) {
	msg, _ := bot.Rpc.GetMessage(accId, msgId)
	if msg.FromId <= deltachat.ContactLastSpecial { // ignore message from self
		return
	}

	adminChatId, _ := cli.AdminChat(bot, accId)
	if msg.ChatId == adminChatId {
		isAdmin, _ := cli.IsAdmin(bot, accId, msg.FromId)
		if isAdmin {
			switch msg.Text {
			case "/info":
				info, _ := bot.Rpc.GetInfo(accId)
				var text string
				for key, value := range info {
					text += key + "=" + value + "\n"
				}
				bot.Rpc.MiscSendTextMessage(accId, msg.ChatId, text)
			}
		}
	}
}

func main() {
	cli.OnBotInit(func(cli *botcli.BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		bot.OnNewMsg(onNewMsg)
	})
	cli.Start()
}
