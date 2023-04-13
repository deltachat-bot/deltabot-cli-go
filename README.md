#  deltabot-cli for Go

![Latest release](https://img.shields.io/github/v/tag/deltachat-bot/deltabot-cli-go?label=release)
[![Go Reference](https://pkg.go.dev/badge/github.com/deltachat-bot/deltabot-cli-go.svg)](https://pkg.go.dev/github.com/deltachat-bot/deltabot-cli-go)
[![CI](https://github.com/deltachat-bot/deltabot-cli-go/actions/workflows/ci.yml/badge.svg)](https://github.com/deltachat-bot/deltabot-cli-go/actions/workflows/ci.yml)
![Coverage](https://img.shields.io/badge/Coverage-83.3%25-brightgreen)
[![Go Report Card](https://goreportcard.com/badge/github.com/deltachat-bot/deltabot-cli-go)](https://goreportcard.com/report/github.com/deltachat-bot/deltabot-cli-go)

Library to speedup Delta Chat bot development in Golang.

With this library you can focus on writing your event/message processing logic and let us handle the
repetitive process of creating the bot CLI.

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

<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/echobot.go) -->
<!-- MARKDOWN-AUTO-DOCS:END -->

Save the previous code snippet as `echobot.go` then run:

```sh
go mod init echobot; go mod tidy
go run ./echobot.go init bot@example.com PASSWORD
go run ./echobot.go serve
```

Use `go run ./echobot.go --help` to see all the available options.

Check the [examples folder](https://github.com/deltachat-bot/deltabot-cli-go/tree/master/examples) for
more examples.

This package depends on https://github.com/deltachat/deltachat-rpc-client-go library, check its
documentation to better understand how to use the deltachat API.
