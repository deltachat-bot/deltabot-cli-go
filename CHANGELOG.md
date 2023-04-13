# Changelog

## Unreleased

## Added

- added more tests and code coverage

## Changed

- adapted to work with recent API changes in deltachat-rpc-client-go v0.16.1-0.20230413050235-ac4cbf9913e8
- `BotCli.Start()` now returns an error instead of calling `os.Exit(1)`

## v0.3.0

- add `qr` subcommand
- switch to zap logger
- update configAction() to print a new line in the returned config value
- panic if deltachat-rpc-server can't be started and provide hint to installation instructions

## v0.2.0

- log info/warning/error core events by default

## v0.1.0

- initial release
