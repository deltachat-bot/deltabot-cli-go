package botcli

import (
	"context"
	"os"
	"strconv"

	"github.com/chatmail/rpc-client-go/deltachat"
	"github.com/chatmail/rpc-client-go/deltachat/option"
	"github.com/chatmail/rpc-client-go/deltachat/transport"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type _ParsedCmd struct {
	cmd  *cobra.Command
	args []string
}

// A function that can be used as callback in OnBotInit(), OnBotStart() and AddCommand().
type Callback func(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string)

// A CLI program, with subcommands that help configuring and running a Delta Chat bot.
type BotCli struct {
	AppName string
	// AppDir can be set by the --folder flag in command line
	AppDir string
	// SelectedAddr can be set by the --account flag in command line, if empty it means "all accounts"
	SelectedAddr string
	RootCmd      *cobra.Command
	Logger       *zap.SugaredLogger
	cmdsMap      map[string]Callback
	parsedCmd    *_ParsedCmd
	onInit       Callback
	onStart      Callback
}

// Create a new BotCli instance.
func New(appName string) *BotCli {
	cli := &BotCli{
		AppName: appName,
		RootCmd: &cobra.Command{Use: os.Args[0]},
		Logger:  getLogger(),
		cmdsMap: make(map[string]Callback),
	}
	initializeRootCmd(cli)
	return cli
}

// Register function to be called when the bot is initialized.
func (self *BotCli) OnBotInit(callback Callback) {
	self.onInit = callback
}

// Register function to be called if the bot is about to start serving requests.
func (self *BotCli) OnBotStart(callback Callback) {
	self.onStart = callback
}

// Run the CLI program.
func (self *BotCli) Start() error {
	defer self.Logger.Sync() //nolint:errcheck
	err := self.RootCmd.Execute()
	if err != nil {
		return err
	}

	if self.parsedCmd != nil {
		err = os.MkdirAll(self.AppDir, os.ModePerm)
		if err != nil {
			return err
		}

		trans := transport.NewIOTransport()
		trans.AccountsDir = getAccountsDir(self.AppDir)
		rpc := &deltachat.Rpc{Context: context.Background(), Transport: trans}
		defer trans.Close()
		if err := trans.Open(); err != nil {
			self.Logger.Panicf("Failed to start RPC server, read https://github.com/chatmail/core/tree/master/deltachat-rpc-server for installation instructions. Error message: %v", err)
		}

		info, err := rpc.GetSystemInfo()
		if err != nil {
			self.Logger.Panic(err)
		}
		self.Logger.Infof("Running deltachat core %v", info["deltachat_core_version"])

		bot := deltachat.NewBot(rpc)
		bot.On(deltachat.EventInfo{}, func(bot *deltachat.Bot, accId deltachat.AccountId, event deltachat.Event) {
			self.GetLogger(accId).Info(event.(deltachat.EventInfo).Msg)
		})
		bot.On(deltachat.EventWarning{}, func(bot *deltachat.Bot, accId deltachat.AccountId, event deltachat.Event) {
			self.GetLogger(accId).Warn(event.(deltachat.EventWarning).Msg)
		})
		bot.On(deltachat.EventError{}, func(bot *deltachat.Bot, accId deltachat.AccountId, event deltachat.Event) {
			self.GetLogger(accId).Error(event.(deltachat.EventError).Msg)
		})
		if self.onInit != nil {
			self.onInit(self, bot, self.parsedCmd.cmd, self.parsedCmd.args)
		}
		callback := self.cmdsMap[self.parsedCmd.cmd.Use]
		callback(self, bot, self.parsedCmd.cmd, self.parsedCmd.args)
	}

	return nil
}

// Get a logger for the given account.
func (self *BotCli) GetLogger(accId deltachat.AccountId) *zap.SugaredLogger {
	return self.Logger.With("acc", accId)
}

// Add a subcommand to the CLI. The given callback will be executed when the command is used.
func (self *BotCli) AddCommand(cmd *cobra.Command, callback Callback) {
	if cmd.Run != nil {
		panic("Can not set cmd.Run property, it would be overriden")
	}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		self.parsedCmd = &_ParsedCmd{cmd, args}
	}
	self.RootCmd.AddCommand(cmd)
	self.cmdsMap[cmd.Use] = callback
}

// Store a custom program setting in the given bot. The setting is specific to your application.
//
// The setting is stored using Bot.SetUiConfig() and the key is prefixed with BotCli.AppName.
func (self *BotCli) SetConfig(bot *deltachat.Bot, accId deltachat.AccountId, key string, value option.Option[string]) error {
	return bot.SetUiConfig(accId, self.AppName+"."+key, value)
}

// Get a custom program setting from the given bot. The setting is specific to your application.
//
// The setting is retrieved using Bot.GetUiConfig() and the key is prefixed with BotCli.AppName.
func (self *BotCli) GetConfig(bot *deltachat.Bot, accId deltachat.AccountId, key string) (option.Option[string], error) {
	return bot.GetUiConfig(accId, self.AppName+"."+key)
}

// Get the group of bot administrators.
func (self *BotCli) AdminChat(bot *deltachat.Bot, accId deltachat.AccountId) (deltachat.ChatId, error) {
	if isConf, _ := bot.Rpc.IsConfigured(accId); !isConf {
		return 0, &BotNotConfiguredErr{}
	}

	value, err := self.GetConfig(bot, accId, "admin-chat")
	if err != nil {
		return 0, err
	}

	var chatId deltachat.ChatId

	if value.IsSome() {
		chatIdInt, err := strconv.ParseUint(value.Unwrap(), 10, 0)
		if err != nil {
			return 0, err
		}
		chatId = deltachat.ChatId(chatIdInt)
		selfInGroup, err := bot.Rpc.CanSend(accId, chatId)
		if err != nil {
			return 0, err
		}
		if !selfInGroup {
			value = option.None[string]()
		}
	}

	if value.IsNone() {
		chatId, err = self.ResetAdminChat(bot, accId)
		if err != nil {
			return 0, err
		}
	}

	return chatId, nil
}

// Reset the group of bot administrators, all the members of the old group are no longer admins.
func (self *BotCli) ResetAdminChat(bot *deltachat.Bot, accId deltachat.AccountId) (deltachat.ChatId, error) {
	if isConf, _ := bot.Rpc.IsConfigured(accId); !isConf {
		return 0, &BotNotConfiguredErr{}
	}

	chatId, err := bot.Rpc.CreateGroupChat(accId, "Bot Administrators", true)
	if err != nil {
		return 0, err
	}
	value := strconv.FormatUint(uint64(chatId), 10)
	err = self.SetConfig(bot, accId, "admin-chat", option.Some(value))
	if err != nil {
		return 0, err
	}

	return chatId, nil
}

// Returns true if contact is in the bot administrators group, false otherwise.
func (self *BotCli) IsAdmin(bot *deltachat.Bot, accId deltachat.AccountId, contactId deltachat.ContactId) (bool, error) {
	chatId, err := self.AdminChat(bot, accId)
	if err != nil {
		return false, err
	}
	contacts, err := bot.Rpc.GetChatContacts(accId, chatId)
	if err != nil {
		return false, err
	}
	for _, memberId := range contacts {
		if contactId == memberId {
			return true, nil
		}
	}

	return false, nil
}

// Get account for address, if no account exists create a new one
func (self *BotCli) GetOrCreateAccount(rpc *deltachat.Rpc, addr string) (deltachat.AccountId, error) {
	accId, err := self.GetAccount(rpc, addr)
	if err != nil {
		accId, err = rpc.AddAccount()
		if err != nil {
			return 0, err
		}
		rpc.SetConfig(accId, "addr", option.Some(addr)) //nolint:errcheck
	}
	return accId, nil
}

// Get account for address, if no account exists with the given address, an error is returned
func (self *BotCli) GetAccount(rpc *deltachat.Rpc, addr string) (deltachat.AccountId, error) {
	chatIdInt, err := strconv.ParseUint(addr, 10, 0)
	if err == nil {
		return deltachat.AccountId(chatIdInt), nil
	}

	accounts, _ := rpc.GetAllAccountIds()
	for _, accId := range accounts {
		addr2, _ := self.GetAddress(rpc, accId)
		if addr == addr2 {
			return accId, nil
		}
	}
	return 0, &AccountNotFoundErr{Addr: addr}
}

// Get the address of the given account
func (self *BotCli) GetAddress(rpc *deltachat.Rpc, accId deltachat.AccountId) (string, error) {
	var addr option.Option[string]
	var err error
	isConf, err := rpc.IsConfigured(accId)
	if err != nil {
		return "", err
	}
	if isConf {
		addr, err = rpc.GetConfig(accId, "configured_addr")
	} else {
		addr, err = rpc.GetConfig(accId, "addr")
	}
	return addr.UnwrapOr(""), err
}
