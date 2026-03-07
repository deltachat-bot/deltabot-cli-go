package botcli

import (
	"fmt"
	"strings"

	"github.com/chatmail/rpc-client-go/v2/deltachat"
	"github.com/spf13/cobra"
)

func initializeRootCmd(cli *BotCli) {
	defDir := getDefaultAppDir(cli.AppName)
	cli.RootCmd.PersistentFlags().StringVarP(&cli.AppDir, "folder", "f", defDir, "program's data folder")
	cli.RootCmd.PersistentFlags().Uint32VarP(&cli.SelectedAccount, "account", "a", 0, "operate over this account ID only when running any subcommand")

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "do initial login configuration of a new Delta Chat account. If only one argument is given it must be a configuration URI (ex. dcaccount:)",
		Args:  cobra.RangeArgs(1, 2),
	}
	cli.AddCommand(initCmd, initCallback)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "show a list of existing bot accounts",
		Args:  cobra.ExactArgs(0),
	}
	cli.AddCommand(listCmd, listCallback)

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "remove Delta Chat accounts from the bot",
		Args:  cobra.ExactArgs(0),
	}
	cli.AddCommand(removeCmd, removeCallback)

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "set/get account configuration values",
		Args:  cobra.MaximumNArgs(2),
	}
	cli.AddCommand(configCmd, configCallback)

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "start processing messages",
		Args:  cobra.ExactArgs(0),
	}
	cli.AddCommand(serveCmd, serveCallback)

	qrCmd := &cobra.Command{
		Use:   "link",
		Short: "print the bot's chat invitation link",
		Args:  cobra.ExactArgs(0),
	}
	cli.AddCommand(qrCmd, qrCallback)

	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "get the invitation link to the bot administration group, WARNING: don't share this",
		Args:  cobra.ExactArgs(0),
	}
	adminCmd.Flags().BoolP("reset", "r", false, "reset admin chat, removes all existing admins")
	cli.AddCommand(adminCmd, adminCallback)
}

func initCallback(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	bot.On(&deltachat.EventTypeConfigureProgress{}, func(bot *deltachat.Bot, accId uint32, event deltachat.EventType) {
		ev := event.(*deltachat.EventTypeConfigureProgress)
		cli.Logger.Infof("[account #%v] Configuration progress: %v", accId, ev.Progress)
	})

	var accId uint32
	var err error
	if cli.SelectedAccount == 0 { // create a new account
		accId, err = bot.Rpc.AddAccount()
	} else { // add relay to the selected account
		accId = cli.SelectedAccount
	}

	if err == nil {
		botFlag := "1"
		err = bot.Rpc.SetConfig(accId, "bot", &botFlag)
	}

	if err != nil {
		cli.Logger.Errorf("Configuration failed: %v", err)
		return
	}

	go func() {
		if len(args) == 2 {
			params := deltachat.EnteredLoginParam{Addr: args[0], Password: args[1]}
			err = bot.Rpc.AddOrUpdateTransport(accId, params)
		} else {
			err = bot.Rpc.AddTransportFromQr(accId, args[0])
		}
		if err != nil {
			cli.Logger.Errorf("Configuration failed: %v", err)
		} else {
			cli.Logger.Infof("Account configured successfully.")
		}
		bot.Stop()
	}()
	bot.Run() //nolint:errcheck
}

func configCallback(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	var err error
	var accounts []uint32
	if cli.SelectedAccount == 0 { // set config for all accounts
		accounts, err = bot.Rpc.GetAllAccountIds()
	} else {
		accounts = []uint32{cli.SelectedAccount}
	}
	if err != nil {
		cli.Logger.Error(err)
		return
	}

	for _, accId := range accounts {
		fmt.Printf("Account #%v:\n", accId)
		configForAcc(cli, bot, cmd, args, accId)
		fmt.Println("")
	}

	if len(accounts) == 0 {
		cli.Logger.Errorf("There are no accounts yet, add a new account using the init subcommand")
	}
}

func configForAcc(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string, accId uint32) {
	if len(args) == 0 {
		keys, _ := bot.Rpc.GetConfig(accId, "sys.config_keys")
		for _, key := range strings.Fields(*keys) {
			val, _ := bot.Rpc.GetConfig(accId, key)
			var strval string
			if val != nil {
				strval = *val
			}
			fmt.Printf("%v=%q\n", key, strval)
		}
		return
	}

	var val *string
	var err error
	if len(args) == 2 {
		err = bot.Rpc.SetConfig(accId, args[0], &args[1])
	}
	if err == nil {
		val, err = bot.Rpc.GetConfig(accId, args[0])
	}
	if err == nil {
		var strval string
		if val != nil {
			strval = *val
		}
		fmt.Printf("%v=%v\n", args[0], strval)
	} else {
		cli.Logger.Error(err)
	}
}

func serveCallback(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	if cli.SelectedAccount != 0 {
		cli.Logger.Errorf("operation not supported for a single account, discard the -a/--account option and try again")
		return
	}

	accounts, err := bot.Rpc.GetAllAccountIds()
	if err != nil {
		cli.Logger.Error(err)
		return
	}
	var inviteLinks []string
	for _, accId := range accounts {
		if isConf, _ := bot.Rpc.IsConfigured(accId); !isConf {
			cli.Logger.Errorf("account #%v not configured", accId)
		} else {
			inviteLink, _ := bot.Rpc.GetChatSecurejoinQrCode(accId, nil)
			inviteLinks = append(inviteLinks, inviteLink)
		}
	}
	if len(inviteLinks) != 0 {
		cli.Logger.Infof("Listening at: %v", strings.Join(inviteLinks, "\n"))
		if cli.onStart != nil {
			cli.onStart(cli, bot, cmd, args)
		}
		bot.Run() //nolint:errcheck
	} else {
		cli.Logger.Errorf("There are no configured accounts to serve")
	}
}

func qrCallback(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	var err error
	var accounts []uint32
	if cli.SelectedAccount == 0 { // for all accounts
		accounts, err = bot.Rpc.GetAllAccountIds()
	} else {
		accounts = []uint32{cli.SelectedAccount}
	}
	if err != nil {
		cli.Logger.Error(err)
		return
	}

	for _, accId := range accounts {
		fmt.Printf("Account #%v:\n", accId)
		qrForAcc(cli, bot, cmd, args, accId)
		fmt.Println("")
	}

	if len(accounts) == 0 {
		cli.Logger.Errorf("There are no accounts yet, add a new account using the init subcommand")
	}
}

func qrForAcc(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string, accId uint32) {
	if isConf, _ := bot.Rpc.IsConfigured(accId); isConf {
		qrdata, err := bot.Rpc.GetChatSecurejoinQrCode(accId, nil)
		if err != nil {
			cli.Logger.Errorf("Failed to generate invite link: %v", err)
			return
		}
		fmt.Println(qrdata)
	} else {
		cli.Logger.Error("account not configured")
	}
}

func adminCallback(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	var err error
	var accounts []uint32
	if cli.SelectedAccount == 0 { // for all accounts
		accounts, err = bot.Rpc.GetAllAccountIds()
		if err == nil && len(accounts) == 0 {
			cli.Logger.Errorf("There are no accounts yet, add a new account using the init subcommand")
		}
	} else {
		accounts = []uint32{cli.SelectedAccount}
	}
	if err != nil {
		cli.Logger.Error(err)
		return
	}

	for _, accId := range accounts {
		fmt.Printf("Account #%v:\n", accId)
		adminForAcc(cli, bot, cmd, args, accId)
		fmt.Println("")
	}
}

func adminForAcc(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string, accId uint32) {
	if isConf, _ := bot.Rpc.IsConfigured(accId); !isConf {
		cli.Logger.Error("account not configured")
		return
	}

	errMsg := "Failed to generate QR: %v"

	reset, err := cmd.Flags().GetBool("reset")
	if err != nil {
		cli.Logger.Errorf(errMsg, err)
		return
	}
	var chatId uint32
	if reset {
		chatId, err = cli.ResetAdminChat(bot, accId)
	} else {
		chatId, err = cli.AdminChat(bot, accId)
	}
	if err != nil {
		cli.Logger.Errorf(errMsg, err)
		return
	}

	qrdata, err := bot.Rpc.GetChatSecurejoinQrCode(accId, &chatId)
	if err != nil {
		cli.Logger.Errorf(errMsg, err)
		return
	}

	fmt.Println("Use this invite link to become bot administrator")
	fmt.Println(qrdata)
}

func listCallback(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	if cli.SelectedAccount != 0 {
		cli.Logger.Errorf("operation not supported for a single account, discard the -a/--account option and try again")
		return
	}

	accounts, err := bot.Rpc.GetAllAccountIds()
	if err != nil {
		cli.Logger.Error(err)
		return
	}
	for _, accId := range accounts {
		relays, err := bot.Rpc.ListTransports(accId)
		if err != nil {
			cli.Logger.Error(err)
			continue
		}

		var addrs string
		for index, relay := range relays {
			if index == 0 {
				addrs = relay.Addr
			} else {
				addrs += ", " + relay.Addr
			}
		}
		if addrs == "" {
			addrs = "(not configured)"
		}
		fmt.Printf("#%v - %v\n", accId, addrs)
	}
}

func removeCallback(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	var err error
	var accounts []uint32
	if cli.SelectedAccount == 0 { // for all accounts
		accounts, err = bot.Rpc.GetAllAccountIds()
		if err == nil && len(accounts) == 0 {
			cli.Logger.Errorf("There are no accounts yet, add a new account using the init subcommand")
		}
	} else {
		accounts = []uint32{cli.SelectedAccount}
	}
	if err != nil {
		cli.Logger.Error(err)
		return
	}

	if len(accounts) > 1 {
		cli.Logger.Error("There are more than one account, to remove one of them, pass the account address with -a/--account option")
		return
	}

	for _, accId := range accounts {
		err = bot.Rpc.RemoveAccount(accId)
		if err != nil {
			cli.Logger.Error(err)
		} else {
			cli.Logger.Infof("Account #%v removed successfully.", accId)
		}
	}
}
