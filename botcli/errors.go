package botcli

// The bot is not configured yet.
type BotNotConfiguredErr struct{}

func (self *BotNotConfiguredErr) Error() string {
	return "bot account not configured"
}
