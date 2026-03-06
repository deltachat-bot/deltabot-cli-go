// This example demonstrates how to create bots that have administrators.
//
// The bot has the /info command that can only be executed by bot administrators (members of the admins chat).
// To become admin you must use the `admin` subcommand in the cli, and open the invite link that will be shown.
package main

import (
	"fmt"

	"github.com/chatmail/rpc-client-go/v2/deltachat"
	"github.com/deltachat-bot/deltabot-cli-go/botcli"
	"github.com/spf13/cobra"
)

var cli *botcli.BotCli = botcli.New("infobot")

// Process messages sent by administrators.
func onNewMsg(bot *deltachat.Bot, accId uint32, msgId uint32) {
	msg, _ := bot.Rpc.GetMessage(accId, msgId)
	if msg.FromId <= deltachat.ContactLastSpecial { // ignore message from self
		return
	}

	isAdmin, _ := cli.IsAdmin(bot, accId, msg.FromId)
	if isAdmin {
		switch msg.Text {
		case "/info":
			info, _ := bot.Rpc.GetInfo(accId)
			var text string
			for key, value := range info {
				text += key + "=" + value + "\n"
			}
			bot.Rpc.SendMsg(accId, msg.ChatId, deltachat.MessageData{Text: &text})
		}
	}
}

func main() {
	// add an "info" CLI subcommand as example
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "display information about the Delta Chat core running in this system or about an specific account if one was selected with -a/--account",
		Args:  cobra.ExactArgs(0),
	}
	cli.AddCommand(infoCmd, func(cli *botcli.BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		var info map[string]string
		if cli.SelectedAccount == 0 { // no account selected with --a/--account, show system info
			info, _ = bot.Rpc.GetSystemInfo()
		} else { // account selected, show info about that account
			info, _ = bot.Rpc.GetInfo(cli.SelectedAccount)
		}
		for key, val := range info {
			fmt.Printf("%v=%#v\n", key, val)
		}
	})

	cli.OnBotInit(func(cli *botcli.BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		bot.OnNewMsg(onNewMsg)
	})
	cli.Start()
}
