#  deltabot-cli for Go

[![Latest release](https://img.shields.io/github/v/tag/deltachat-bot/deltabot-cli-go?label=release)](https://pkg.go.dev/github.com/deltachat-bot/deltabot-cli-go/v2)
[![Go Reference](https://pkg.go.dev/badge/github.com/deltachat-bot/deltabot-cli-go/v2.svg)](https://pkg.go.dev/github.com/deltachat-bot/deltabot-cli-go/v2)
[![CI](https://github.com/deltachat-bot/deltabot-cli-go/actions/workflows/ci.yml/badge.svg)](https://github.com/deltachat-bot/deltabot-cli-go/actions/workflows/ci.yml)
![Coverage](https://img.shields.io/badge/Coverage-47.0%25-yellow)
[![Go Report Card](https://goreportcard.com/badge/github.com/deltachat-bot/deltabot-cli-go/v2)](https://goreportcard.com/report/github.com/deltachat-bot/deltabot-cli-go/v2)

Library to speedup Delta Chat bot development in Golang.

With this library you can focus on writing your event/message
processing logic and let us handle the repetitive process of
creating the bot CLI.

## Install

```sh
go get -u github.com/deltachat-bot/deltabot-cli-go/v2
```

### Installing deltachat-rpc-server

This package depends on a standalone Delta Chat RPC server `deltachat-rpc-server` program that must be
available in your `PATH`. For installation instructions check:
https://github.com/chatmail/core/tree/main/deltachat-rpc-server

## Usage

Example echo-bot written with deltabot-cli:

<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/echobot/echobot.go) -->
<!-- The below code snippet is automatically added from ./examples/echobot/echobot.go -->
```go
package main

import (
	"github.com/chatmail/rpc-client-go/v2/deltachat"
	"github.com/deltachat-bot/deltabot-cli-go/v2/botcli"
	"github.com/spf13/cobra"
)

func main() {
	cli := botcli.New("echobot")

	cli.OnBotInit(func(cli *botcli.BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		// incoming message handling
		bot.OnNewMsg(func(bot *deltachat.Bot, accId uint32, msgId uint32) {
			msg, _ := bot.Rpc.GetMessage(accId, msgId)
			if msg.FromId > deltachat.ContactLastSpecial && msg.Text != "" {
				_, _ = bot.Rpc.SendMsg(accId, msg.ChatId, deltachat.MessageData{Text: &msg.Text})
			}
		})
	})
	cli.OnBotStart(func(cli *botcli.BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		cli.Logger.Info("OnBotStart event triggered: bot is about to start!")
	})

	_ = cli.Start()
}
```
<!-- MARKDOWN-AUTO-DOCS:END -->

Save the previous code snippet as `echobot.go` then run:

```sh
go mod init echobot; go mod tidy
go run ./echobot.go init bot@example.com PASSWORD
go run ./echobot.go serve
```

Use `go run ./echobot.go --help` to see all the available options.

Check the [examples folder](./examples) for more examples.

This package depends on https://github.com/chatmail/rpc-client-go library, check its
documentation to better understand how to use the Delta Chat API.
