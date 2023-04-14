package botcli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
	"github.com/spf13/cobra"
)

func initializeRootCmd(cli *BotCli) {
	defDir := getDefaultAppDir(cli.AppName)
	cli.RootCmd.PersistentFlags().StringVarP(&cli.AppDir, "folder", "f", defDir, "program's data folder")

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "initialize the Delta Chat account",
		Args:  cobra.ExactArgs(2),
	}
	cli.AddCommand(initCmd, cli.initCallback)

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "set/get account configuration values",
		Args:  cobra.MaximumNArgs(2),
	}
	cli.AddCommand(configCmd, cli.configCallback)

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "start processing messages",
		Args:  cobra.ExactArgs(0),
	}
	cli.AddCommand(serveCmd, cli.serveCallback)

	qrCmd := &cobra.Command{
		Use:   "qr",
		Short: "get bot's verification QR",
		Args:  cobra.ExactArgs(0),
	}
	qrCmd.Flags().BoolP("invert", "i", false, "invert QR colors")
	cli.AddCommand(qrCmd, cli.qrCallback)

	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "get the invitation QR to the bot administration group, WARNING: don't share this QR",
		Args:  cobra.ExactArgs(0),
	}
	adminCmd.Flags().BoolP("invert", "i", false, "invert QR colors")
	adminCmd.Flags().BoolP("reset", "r", false, "reset admin chat, removes all existing admins")
	cli.AddCommand(adminCmd, cli.adminCallback)
}

func (self *BotCli) initCallback(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	bot.On(deltachat.EventConfigureProgress{}, func(event deltachat.Event) {
		ev := event.(deltachat.EventConfigureProgress)
		self.Logger.Infof("Configuration progress: %v", ev.Progress)
	})

	go func() {
		if err := bot.Configure(args[0], args[1]); err != nil {
			self.Logger.Errorf("Configuration failed: %v", err)
		} else {
			self.Logger.Info("Account configured successfully.")
		}
		bot.Stop()
	}()
	bot.Run() //nolint:errcheck
}

func (self *BotCli) configCallback(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
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
		fmt.Printf("%v=%v\n", args[0], val)
	} else {
		self.Logger.Error(err)
	}
}

func (self *BotCli) serveCallback(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	if bot.IsConfigured() {
		if self.onStart != nil {
			self.onStart(bot, self.parsedCmd.cmd, self.parsedCmd.args)
		}
		bot.Run() //nolint:errcheck
	} else {
		self.Logger.Error("account not configured")
	}
}

func (self *BotCli) qrCallback(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	if bot.IsConfigured() {
		qrdata, _, err := bot.Account.QrCode()
		if err != nil {
			self.Logger.Errorf("Failed to generate QR: %v", err)
			return
		}
		addr, _ := bot.GetConfig("configured_addr")
		fmt.Println("Scan this QR to verify", addr)
		invert, _ := cmd.Flags().GetBool("invert")
		printQr(qrdata, invert)
		fmt.Println(qrdata)
	} else {
		self.Logger.Error("account not configured")
	}
}

func (self *BotCli) adminCallback(bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	if !bot.IsConfigured() {
		self.Logger.Error("account not configured")
		return
	}

	errMsg := "Failed to generate QR: %v"

	reset, err := cmd.Flags().GetBool("reset")
	if err != nil {
		self.Logger.Errorf(errMsg, err)
		return
	}
	var value string
	if !reset {
		value, err = self.GetConfig(bot, "admin-chat")
		if err != nil {
			self.Logger.Errorf(errMsg, err)
			return
		}
	}

	var chat *deltachat.Chat

	if value != "" {
		chatId, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			self.Logger.Errorf(errMsg, err)
			return
		}
		chat = &deltachat.Chat{Account: bot.Account, Id: deltachat.ChatId(chatId)}
		var selfInGroup bool
		contacts, err := chat.Contacts()
		if err != nil {
			self.Logger.Errorf(errMsg, err)
			return
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
		chat, err = bot.Account.CreateGroup("Bot Administrators", true)
		if err != nil {
			self.Logger.Errorf(errMsg, err)
			return
		}
		value = strconv.FormatUint(uint64(chat.Id), 10)
		err = self.SetConfig(bot, "admin-chat", value)
		if err != nil {
			self.Logger.Errorf(errMsg, err)
			return
		}
	}

	qrdata, _, err := chat.QrCode()
	if err != nil {
		self.Logger.Errorf(errMsg, err)
		return
	}

	fmt.Println("Scan this QR to become bot administrator")
	invert, _ := cmd.Flags().GetBool("invert")
	printQr(qrdata, invert)
	fmt.Println(qrdata)
}
