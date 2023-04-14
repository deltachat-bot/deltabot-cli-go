package botcli

import (
	"os"
	"strconv"

	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type _ParsedCmd struct {
	cmd  *cobra.Command
	args []string
}

// A function that can be used as callback in OnBotInit(), OnBotStart() and AddCommand().
type Callback func(bot *deltachat.Bot, cmd *cobra.Command, args []string)

// A CLI program, with subcommands that help configuring and running a Delta Chat bot.
type BotCli struct {
	AppName   string
	AppDir    string
	RootCmd   *cobra.Command
	Logger    *zap.SugaredLogger
	cmdsMap   map[string]Callback
	parsedCmd *_ParsedCmd
	onInit    Callback
	onStart   Callback
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
		rpc := deltachat.NewRpcIO()
		rpc.AccountsDir = getAccountsDir(self.AppDir)
		defer rpc.Stop()
		if err := rpc.Start(); err != nil {
			self.Logger.Panicf("Failed to start RPC server, read https://github.com/deltachat/deltachat-core-rust/tree/master/deltachat-rpc-server for installation instructions. Error message: %v", err)
		}
		bot := deltachat.NewBotFromAccountManager(&deltachat.AccountManager{Rpc: rpc})
		bot.On(deltachat.EventInfo{}, func(event deltachat.Event) {
			self.Logger.Info(event.(deltachat.EventInfo).Msg)
		})
		bot.On(deltachat.EventWarning{}, func(event deltachat.Event) {
			self.Logger.Warn(event.(deltachat.EventWarning).Msg)
		})
		bot.On(deltachat.EventError{}, func(event deltachat.Event) {
			self.Logger.Error(event.(deltachat.EventError).Msg)
		})
		if self.onInit != nil {
			self.onInit(bot, self.parsedCmd.cmd, self.parsedCmd.args)
		}
		callback := self.cmdsMap[self.parsedCmd.cmd.Use]
		callback(bot, self.parsedCmd.cmd, self.parsedCmd.args)
	}

	return nil
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
func (self *BotCli) SetConfig(bot *deltachat.Bot, key, value string) error {
	return bot.SetUiConfig(self.AppName+"."+key, value)
}

// Get a custom program setting from the given bot. The setting is specific to your application.
//
// The setting is retrieved using Bot.GetUiConfig() and the key is prefixed with BotCli.AppName.
func (self *BotCli) GetConfig(bot *deltachat.Bot, key string) (string, error) {
	return bot.GetUiConfig(self.AppName + "." + key)
}

// Get the group of bot administrators.
func (self *BotCli) AdminChat(bot *deltachat.Bot) (*deltachat.Chat, error) {
	if !bot.IsConfigured() {
		return nil, &BotNotConfiguredErr{}
	}

	value, err := self.GetConfig(bot, "admin-chat")
	if err != nil {
		return nil, err
	}

	var chat *deltachat.Chat

	if value != "" {
		chatId, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return nil, err
		}
		chat = &deltachat.Chat{Account: bot.Account, Id: deltachat.ChatId(chatId)}
		var selfInGroup bool
		contacts, err := chat.Contacts()
		if err != nil {
			return nil, err
		}
		me := bot.Me()
		for _, contact := range contacts {
			if me.Id == contact.Id {
				selfInGroup = true
				break
			}
		}
		if !selfInGroup {
			value = ""
		}
	}

	if value == "" {
		chat, err = self.ResetAdminChat(bot)
		if err != nil {
			return nil, err
		}
	}

	return chat, nil
}

// Reset the group of bot administrators, all the members of the old group are no longer admins.
func (self *BotCli) ResetAdminChat(bot *deltachat.Bot) (*deltachat.Chat, error) {
	if !bot.IsConfigured() {
		return nil, &BotNotConfiguredErr{}
	}

	chat, err := bot.Account.CreateGroup("Bot Administrators", true)
	if err != nil {
		return nil, err
	}
	value := strconv.FormatUint(uint64(chat.Id), 10)
	err = self.SetConfig(bot, "admin-chat", value)
	if err != nil {
		return nil, err
	}

	return chat, nil
}

// Returns true if contact is in the bot administrators group, false otherwise.
func (self *BotCli) IsAdmin(bot *deltachat.Bot, contact *deltachat.Contact) (bool, error) {
	chat, err := self.AdminChat(bot)
	if err != nil {
		return false, err
	}
	contacts, err := chat.Contacts()
	if err != nil {
		return false, err
	}
	for _, member := range contacts {
		if contact.Id == member.Id {
			return true, nil
		}
	}

	return false, nil
}
