#  deltabot-cli for Go

[![CI](https://github.com/deltachat-bot/deltabot-cli-go/actions/workflows/ci.yml/badge.svg)](https://github.com/deltachat-bot/deltabot-cli-go/actions/workflows/ci.yml)
![Go version](https://img.shields.io/github/go-mod/go-version/deltachat-bot/deltabot-cli-go)

Library to speedup Delta Chat bot development in Golang.

With this library you can focus on writing your event/message processing logic and let us handle the repetitive
process of creating the bot CLI.

## Install

```sh
go get -u github.com/deltachat-bot/deltabot-cli-go
```

### Installing deltachat-rpc-server

This package depends on a standalone Delta Chat RPC server `deltachat-rpc-server` program that must be
available in your `PATH`. For installation instructions check:
https://github.com/deltachat/deltachat-core-rust/tree/master/deltachat-rpc-server

## Usage

Example echo-bot written with deltabot-cli:

```go
package main

import (
	"github.com/deltachat-bot/deltabot-cli-go/botcli"
	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
	"github.com/spf13/cobra"
)

func main() {
	cli := botcli.New("echobot")
	cli.OnBotInit(func(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
		bot.OnNewMsg(func(msg *deltachat.Message) {
			snapshot, _ := msg.Snapshot()
			chat := deltachat.Chat{bot.Account, snapshot.ChatId}
			chat.SendText(snapshot.Text)
		})
	})
	cli.Start()
}
```

Save the previous code snippet as `echobot.go` then run:

```sh
go run ./echobot.go init bot@example.com PASSWORD
go run ./echobot.go serve
```

Use `go run ./echobot.go --help` to see all the available options.

Check the [examples folder](https://github.com/deltachat-bot/deltabot-cli-go/tree/master/examples) for
more examples.
