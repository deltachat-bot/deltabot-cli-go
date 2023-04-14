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

var cli *botcli.BotCli

// Process messages sent to the group of administrators and allow to run privileged commands there.
func onNewMsg(bot *deltachat.Bot, msg *deltachat.Message) {
	snapshot, _ := msg.Snapshot()
	chat := &deltachat.Chat{snapshot.Account, snapshot.ChatId}
	sender := &deltachat.Contact{snapshot.Account, snapshot.FromId}
	isAdmin, _ := cli.IsAdmin(bot, sender)
	adminChat, _ := cli.AdminChat(bot)
	if !isAdmin || chat.Id != adminChat.Id {
		return
	}

	switch snapshot.Text {
	case "/info":
		info, _ := bot.Account.Info()
		var text string
		for key, value := range info {
			text += key + "=" + value + "\n"
		}
		chat.SendText(text)
	}
}

func main() {
	cli = botcli.New("echobot")
	cli.OnBotInit(func(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		bot.OnNewMsg(func(msg *deltachat.Message) { onNewMsg(bot, msg) })
	})
	cli.Start()
}
