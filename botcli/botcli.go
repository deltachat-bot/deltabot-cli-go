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
	Logger        *zap.Logger
	actionsMap    map[string]CommandAction
	parsedCmd     *_ParsedCmd
	onInitAction  func(bot *deltachat.Bot, cmd *cobra.Command, args []string)
	onStartAction func(bot *deltachat.Bot, cmd *cobra.Command, args []string)
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
func (self *BotCli) Start() {
	defer self.Logger.Sync()
	err := self.RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

	if self.parsedCmd != nil {
		os.MkdirAll(self.AppDir, os.ModeDir)
		rpc := deltachat.NewRpc()
		rpc.AccountsDir = getAccountsDir(self.AppDir)
		defer rpc.Stop()
		rpc.Start()
		bot := deltachat.NewBotFromAccountManager(deltachat.NewAccountManager(rpc))
		if self.onInitAction != nil {
			self.onInitAction(bot, self.parsedCmd.cmd, self.parsedCmd.args)
		}
		action := self.actionsMap[self.parsedCmd.cmd.Use]
		action(bot, self.parsedCmd.cmd, self.parsedCmd.args)
	}
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

// Create a new BotCli instance
func New(appName string) *BotCli {
	logger, _ := zap.NewProduction()
	cli := &BotCli{
		AppName:    appName,
		RootCmd:    &cobra.Command{Use: appName},
		Logger:     logger,
		actionsMap: make(map[string]CommandAction),
	}
	initializeRootCmd(cli)
	return cli
}
