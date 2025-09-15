package botcli

// The bot is not configured yet.
type BotNotConfiguredErr struct{}

func (error *BotNotConfiguredErr) Error() string {
	return "bot account not configured"
}

// The account was not found.
type AccountNotFoundErr struct{ Addr string }

func (error *AccountNotFoundErr) Error() string {
	return "account not found: " + error.Addr
}
