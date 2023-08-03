package botcli

// The bot is not configured yet.
type BotNotConfiguredErr struct{}

func (self *BotNotConfiguredErr) Error() string {
	return "bot account not configured"
}

// The account was not found.
type AccountNotFoundErr struct{ Addr string }

func (self *AccountNotFoundErr) Error() string {
	return "account not found: " + self.Addr
}
