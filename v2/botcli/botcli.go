package botcli

import (
	"context"
	"os"
	"strconv"

	"github.com/chatmail/rpc-client-go/v2/deltachat"
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
	// SelectedAccount can be set by the --account flag in command line, if empty it means "all accounts"
	SelectedAccount uint32
	RootCmd         *cobra.Command
	Logger          *zap.SugaredLogger
	cmdsMap         map[string]Callback
	parsedCmd       *_ParsedCmd
	onInit          Callback
	onStart         Callback
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
func (botcli *BotCli) OnBotInit(callback Callback) {
	botcli.onInit = callback
}

// Register function to be called if the bot is about to start serving requests.
func (botcli *BotCli) OnBotStart(callback Callback) {
	botcli.onStart = callback
}

// Run the CLI program.
func (botcli *BotCli) Start() error {
	defer botcli.Logger.Sync() //nolint:errcheck
	err := botcli.RootCmd.Execute()
	if err != nil {
		return err
	}

	if botcli.parsedCmd != nil {
		err = os.MkdirAll(botcli.AppDir, os.ModePerm)
		if err != nil {
			return err
		}

		trans := deltachat.NewIOTransport()
		trans.AccountsDir = getAccountsDir(botcli.AppDir)
		rpc := &deltachat.Rpc{Context: context.Background(), Transport: trans}
		defer trans.Close()
		if err := trans.Open(); err != nil {
			botcli.Logger.Panicf("Failed to start RPC server, read https://github.com/chatmail/core/tree/master/deltachat-rpc-server for installation instructions. Error message: %v", err)
		}

		info, err := rpc.GetSystemInfo()
		if err != nil {
			botcli.Logger.Panic(err)
		}
		botcli.Logger.Infof("Running deltachat core %v", info["deltachat_core_version"])

		bot := deltachat.NewBot(rpc)
		bot.On(&deltachat.EventTypeInfo{}, func(bot *deltachat.Bot, accId uint32, event deltachat.EventType) {
			botcli.GetLogger(accId).Info(event.(*deltachat.EventTypeInfo).Msg)
		})
		bot.On(&deltachat.EventTypeWarning{}, func(bot *deltachat.Bot, accId uint32, event deltachat.EventType) {
			botcli.GetLogger(accId).Warn(event.(*deltachat.EventTypeWarning).Msg)
		})
		bot.On(&deltachat.EventTypeError{}, func(bot *deltachat.Bot, accId uint32, event deltachat.EventType) {
			botcli.GetLogger(accId).Error(event.(*deltachat.EventTypeError).Msg)
		})
		if botcli.onInit != nil {
			botcli.onInit(botcli, bot, botcli.parsedCmd.cmd, botcli.parsedCmd.args)
		}
		callback := botcli.cmdsMap[botcli.parsedCmd.cmd.Use]
		callback(botcli, bot, botcli.parsedCmd.cmd, botcli.parsedCmd.args)
	}

	return nil
}

// Get a logger for the given account.
func (botcli *BotCli) GetLogger(accId uint32) *zap.SugaredLogger {
	return botcli.Logger.With("acc", accId)
}

// Add a subcommand to the CLI. The given callback will be executed when the command is used.
func (botcli *BotCli) AddCommand(cmd *cobra.Command, callback Callback) {
	if cmd.Run != nil {
		panic("Can not set cmd.Run property, it would be overriden")
	}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		botcli.parsedCmd = &_ParsedCmd{cmd, args}
	}
	botcli.RootCmd.AddCommand(cmd)
	botcli.cmdsMap[cmd.Use] = callback
}

// Store a custom program setting in the given bot. The setting is specific to your application.
func (botcli *BotCli) SetConfig(bot *deltachat.Bot, accId uint32, key string, value *string) error {
	return bot.Rpc.SetConfig(accId, "ui."+botcli.AppName+"."+key, value)
}

// Get a custom program setting from the given bot. The setting is specific to your application.
func (botcli *BotCli) GetConfig(bot *deltachat.Bot, accId uint32, key string) (*string, error) {
	return bot.Rpc.GetConfig(accId, "ui."+botcli.AppName+"."+key)
}

// Get the group of bot administrators.
func (botcli *BotCli) AdminChat(bot *deltachat.Bot, accId uint32) (uint32, error) {
	if isConf, _ := bot.Rpc.IsConfigured(accId); !isConf {
		return 0, &BotNotConfiguredErr{}
	}

	value, err := botcli.GetConfig(bot, accId, "admin-chat")
	if err != nil {
		return 0, err
	}

	var chatId uint32

	if value != nil {
		chatIdInt, err := strconv.ParseUint(*value, 10, 0)
		if err != nil {
			return 0, err
		}
		chatId = uint32(chatIdInt)
		selfInGroup, err := bot.Rpc.CanSend(accId, chatId)
		if err != nil {
			return 0, err
		}
		if !selfInGroup {
			value = nil
		}
	}

	if value == nil {
		chatId, err = botcli.ResetAdminChat(bot, accId)
		if err != nil {
			return 0, err
		}
	}

	return chatId, nil
}

// Reset the group of bot administrators, all the members of the old group are no longer admins.
func (botcli *BotCli) ResetAdminChat(bot *deltachat.Bot, accId uint32) (uint32, error) {
	if isConf, _ := bot.Rpc.IsConfigured(accId); !isConf {
		return 0, &BotNotConfiguredErr{}
	}

	chatId, err := bot.Rpc.CreateGroupChat(accId, "Bot Administrators", false)
	if err != nil {
		return 0, err
	}
	value := strconv.FormatUint(uint64(chatId), 10)
	err = botcli.SetConfig(bot, accId, "admin-chat", &value)
	if err != nil {
		return 0, err
	}

	return chatId, nil
}

// Returns true if contact is in the bot administrators group, false otherwise.
func (botcli *BotCli) IsAdmin(bot *deltachat.Bot, accId uint32, contactId uint32) (bool, error) {
	chatId, err := botcli.AdminChat(bot, accId)
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
