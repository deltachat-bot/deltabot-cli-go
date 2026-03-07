/*
# Webxdc RPC Example

This is an example bot project using the `xdcrpc` package for
communication between the backend bot and a frontend webxdc app.

To run the bot:

```sh
go run . dcaccount:nine.testrun.org
```

NOTE: For this example to work, a app.xdc file must be provided
in thecurrent working dir.
*/
package main

import (
	"github.com/chatmail/rpc-client-go/v2/deltachat"
	"github.com/deltachat-bot/deltabot-cli-go/v2/botcli"
	"github.com/deltachat-bot/deltabot-cli-go/v2/xdcrpc"
	"github.com/spf13/cobra"
)

var cli = botcli.New("webxdcbot")

func main() {
	cli.OnBotInit(func(cli *botcli.BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		bot.OnUnhandledEvent(onEvent)
		bot.OnNewMsg(onNewMsg)
	})
	_ = cli.Start()
}

func onEvent(bot *deltachat.Bot, accId uint32, event deltachat.EventType) {
	switch ev := event.(type) {
	case *deltachat.EventTypeWebxdcStatusUpdate:
		_ = xdcrpc.HandleMessage(bot.Rpc, accId, ev.MsgId, ev.StatusUpdateSerial, &API{})
	}
}

func onNewMsg(bot *deltachat.Bot, accId uint32, msgId uint32) {
	msg, _ := bot.Rpc.GetMessage(accId, msgId)
	logger := cli.GetLogger(accId).With("chat", msg.ChatId)
	if msg.FromId > deltachat.ContactLastSpecial {
		logger.Info("message received, sending the mini-app")
		file := "app.xdc"
		_, err := bot.Rpc.SendMsg(accId, msg.ChatId, deltachat.MessageData{File: &file})
		if err != nil {
			logger.Error(err)
		}
	}
}
