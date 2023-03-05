package botcli

import (
	"fmt"
	"strings"

	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
	"github.com/spf13/cobra"
)

type CommandAction func(bot *deltachat.Bot, cmd *cobra.Command, args []string)

func initializeRootCmd(cli *BotCli) {
	defDir := getDefaultAppDir(cli.AppName)
	cli.RootCmd.PersistentFlags().StringVarP(&cli.AppDir, "folder", "f", defDir, "program's data folder")

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "initialize the Delta Chat account",
		Args:  cobra.ExactArgs(2),
	}
	cli.AddCommand(initCmd, cli.initAction)

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "set/get account configuration values",
		Args:  cobra.MaximumNArgs(2),
	}
	cli.AddCommand(configCmd, cli.configAction)

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "start processing messages",
		Args:  cobra.ExactArgs(0),
	}
	cli.AddCommand(serveCmd, cli.serveAction)
}

func (self *BotCli) initAction(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	bot.On(deltachat.EVENT_CONFIGURE_PROGRESS, func(event *deltachat.Event) {
		self.Logger.Info().Msgf("Configuration progress: %v", event.Progress)
		if event.Progress == 1000 || event.Progress == 0 {
			go bot.Stop()
		}
	})
	result := make(chan error)
	go func() { result <- bot.Configure(args[0], args[1]) }()
	bot.Run()
	if err := <-result; err != nil {
		self.Logger.Error().Msgf("Configuration failed: %v", err)
	} else {
		self.Logger.Info().Msg("Account configured successfully.")
	}
}

func (self *BotCli) configAction(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	var val string
	var err error
	if len(args) == 0 {
		val, _ := bot.GetConfig("sys.config_keys")
		for _, key := range strings.Fields(val) {
			val, _ := bot.GetConfig(key)
			fmt.Printf("%v=%q\n", key, val)
		}
		return
	}

	if len(args) == 2 {
		err = bot.SetConfig(args[0], args[1])
	}
	if err == nil {
		val, err = bot.GetConfig(args[0])
	}
	if err == nil {
		fmt.Printf("%v=%v", args[0], val)
	} else {
		self.Logger.Error().Err(err)
	}
}

func (self *BotCli) serveAction(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	if bot.IsConfigured() {
		if self.onStartAction != nil {
			self.onStartAction(bot, self.parsedCmd.cmd, self.parsedCmd.args)
		}
		bot.Run()
	} else {
		self.Logger.Error().Msg("account not configured")
	}
}
