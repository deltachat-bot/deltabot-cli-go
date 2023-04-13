package botcli

import (
	"os"

	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type _ParsedCmd struct {
	cmd  *cobra.Command
	args []string
}

type BotCli struct {
	AppName       string
	AppDir        string
	RootCmd       *cobra.Command
	Logger        *zap.SugaredLogger
	actionsMap    map[string]CommandAction
	parsedCmd     *_ParsedCmd
	onInitAction  func(bot *deltachat.Bot, cmd *cobra.Command, args []string)
	onStartAction func(bot *deltachat.Bot, cmd *cobra.Command, args []string)
}

// Create a new BotCli instance
func New(appName string) *BotCli {
	cli := &BotCli{
		AppName:    appName,
		RootCmd:    &cobra.Command{Use: os.Args[0]},
		Logger:     getLogger(),
		actionsMap: make(map[string]CommandAction),
	}
	initializeRootCmd(cli)
	return cli
}

// Register function to be called when the bot is initialized.
func (self *BotCli) OnBotInit(action func(bot *deltachat.Bot, cmd *cobra.Command, args []string)) {
	self.onInitAction = action
}

// Register function to be called if the bot is about to start serving requests.
func (self *BotCli) OnBotStart(action func(bot *deltachat.Bot, cmd *cobra.Command, args []string)) {
	self.onStartAction = action
}

// Run the CLI program
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
		if self.onInitAction != nil {
			self.onInitAction(bot, self.parsedCmd.cmd, self.parsedCmd.args)
		}
		action := self.actionsMap[self.parsedCmd.cmd.Use]
		action(bot, self.parsedCmd.cmd, self.parsedCmd.args)
	}

	return nil
}

// Add a subcommand to the CLI. The given action will be executed when the command is used.
func (self *BotCli) AddCommand(cmd *cobra.Command, action CommandAction) {
	if cmd.Run != nil {
		panic("Can not set cmd.Run property, it would be overriden")
	}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		self.parsedCmd = &_ParsedCmd{cmd, args}
	}
	self.RootCmd.AddCommand(cmd)
	self.actionsMap[cmd.Use] = action
}
