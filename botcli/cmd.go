package botcli

import (
	"fmt"
	"strings"

	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
	"github.com/deltachat/deltachat-rpc-client-go/deltachat/option"
	"github.com/spf13/cobra"
)

func initializeRootCmd(cli *BotCli) {
	defDir := getDefaultAppDir(cli.AppName)
	cli.RootCmd.PersistentFlags().StringVarP(&cli.AppDir, "folder", "f", defDir, "program's data folder")
	cli.RootCmd.PersistentFlags().StringVarP(&cli.SelectedAddr, "account", "a", "", "operate over this account only when running any subcommand")

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "do initial login configuration of a new Delta Chat account, if the account already exist, the credentials are updated",
		Args:  cobra.ExactArgs(2),
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
		Use:   "qr",
		Short: "get bot's verification QR",
		Args:  cobra.ExactArgs(0),
	}
	qrCmd.Flags().BoolP("invert", "i", false, "invert QR colors")
	cli.AddCommand(qrCmd, qrCallback)

	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "get the invitation QR to the bot administration group, WARNING: don't share this QR",
		Args:  cobra.ExactArgs(0),
	}
	adminCmd.Flags().BoolP("invert", "i", false, "invert QR colors")
	adminCmd.Flags().BoolP("reset", "r", false, "reset admin chat, removes all existing admins")
	cli.AddCommand(adminCmd, adminCallback)
}

func initCallback(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	bot.On(deltachat.EventConfigureProgress{}, func(bot *deltachat.Bot, accId deltachat.AccountId, event deltachat.Event) {
		ev := event.(deltachat.EventConfigureProgress)
		addr, _ := cli.GetAddress(bot.Rpc, accId)
		if addr == "" {
			addr = fmt.Sprintf("account #%v", accId)
		}
		cli.Logger.Infof("[%v] Configuration progress: %v", addr, ev.Progress)
	})

	var accId deltachat.AccountId
	var err error
	if cli.SelectedAddr == "" { // auto-select based on first argument (or create a new one if not found)
		accId, err = cli.GetOrCreateAccount(bot.Rpc, args[0])
	} else { // re-configure the selected account
		accId, err = cli.GetAccount(bot.Rpc, cli.SelectedAddr)
		if err == nil {
			_, err = cli.GetAccount(bot.Rpc, args[0])
			if err == nil {
				cli.Logger.Errorf("Configuration failed: an account with address %q already exists", args[0])
				return
			}
		}
	}
	if err != nil {
		cli.Logger.Errorf("Configuration failed: %v", err)
		return
	}

	go func() {
		if err := bot.Configure(accId, args[0], args[1]); err != nil {
			cli.Logger.Errorf("Configuration failed: %v", err)
		} else {
			cli.Logger.Infof("Account %q configured successfully.", args[0])
		}
		bot.Stop()
	}()
	bot.Run() //nolint:errcheck
}

func configCallback(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	var err error
	var accounts []deltachat.AccountId
	if cli.SelectedAddr == "" { // set config for all accounts
		accounts, err = bot.Rpc.GetAllAccountIds()
	} else {
		var accId deltachat.AccountId
		accId, err = cli.GetAccount(bot.Rpc, cli.SelectedAddr)
		accounts = []deltachat.AccountId{accId}
	}
	if err != nil {
		cli.Logger.Error(err)
		return
	}

	for _, accId := range accounts {
		addr, err := cli.GetAddress(bot.Rpc, accId)
		if err != nil {
			cli.Logger.Error(err)
			continue
		}
		fmt.Printf("Account #%v (%v):\n", accId, addr)
		configForAcc(cli, bot, cmd, args, accId)
		fmt.Println("")
	}

	if len(accounts) == 0 {
		cli.Logger.Errorf("There are no accounts yet, add a new account using the init subcommand")
	}
}

func configForAcc(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string, accId deltachat.AccountId) {
	if len(args) == 0 {
		keys, _ := bot.Rpc.GetConfig(accId, "sys.config_keys")
		for _, key := range strings.Fields(keys.Unwrap()) {
			val, _ := bot.Rpc.GetConfig(accId, key)
			fmt.Printf("%v=%q\n", key, val.UnwrapOr(""))
		}
		return
	}

	var val option.Option[string]
	var err error
	if len(args) == 2 {
		err = bot.Rpc.SetConfig(accId, args[0], option.Some(args[1]))
	}
	if err == nil {
		val, err = bot.Rpc.GetConfig(accId, args[0])
	}
	if err == nil {
		fmt.Printf("%v=%v\n", args[0], val.UnwrapOr(""))
	} else {
		cli.Logger.Error(err)
	}
}

func serveCallback(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	if cli.SelectedAddr != "" {
		cli.Logger.Errorf("operation not supported for a single account, discard the -a/--account option and try again")
		return
	}

	accounts, err := bot.Rpc.GetAllAccountIds()
	if err != nil {
		cli.Logger.Error(err)
		return
	}
	var addrs []string
	for _, accId := range accounts {
		if isConf, _ := bot.Rpc.IsConfigured(accId); !isConf {
			cli.Logger.Errorf("account #%v not configured", accId)
		} else {
			addr, _ := bot.Rpc.GetConfig(accId, "configured_addr")
			if addr.UnwrapOr("") != "" {
				addrs = append(addrs, addr.Unwrap())
			}
		}
	}
	if len(addrs) != 0 {
		cli.Logger.Infof("Listening at: %v", strings.Join(addrs, ", "))
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
	var accounts []deltachat.AccountId
	if cli.SelectedAddr == "" { // for all accounts
		accounts, err = bot.Rpc.GetAllAccountIds()
	} else {
		var accId deltachat.AccountId
		accId, err = cli.GetAccount(bot.Rpc, cli.SelectedAddr)
		accounts = []deltachat.AccountId{accId}
	}
	if err != nil {
		cli.Logger.Error(err)
		return
	}

	for _, accId := range accounts {
		addr, err := cli.GetAddress(bot.Rpc, accId)
		if err != nil {
			cli.Logger.Error(err)
			continue
		}
		fmt.Printf("Account #%v (%v):\n", accId, addr)
		qrForAcc(cli, bot, cmd, args, accId, addr)
		fmt.Println("")
	}

	if len(accounts) == 0 {
		cli.Logger.Errorf("There are no accounts yet, add a new account using the init subcommand")
	}
}

func qrForAcc(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string, accId deltachat.AccountId, addr string) {
	if isConf, _ := bot.Rpc.IsConfigured(accId); isConf {
		qrdata, _, err := bot.Rpc.GetChatSecurejoinQrCodeSvg(accId, option.None[deltachat.ChatId]())
		if err != nil {
			cli.Logger.Errorf("Failed to generate QR: %v", err)
			return
		}
		fmt.Println("Scan this QR to verify", addr)
		invert, _ := cmd.Flags().GetBool("invert")
		printQr(qrdata, invert)
		fmt.Printf(GenerateInviteLink(qrdata))
	} else {
		cli.Logger.Error("account not configured")
	}
}

func adminCallback(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	var err error
	var accounts []deltachat.AccountId
	if cli.SelectedAddr == "" { // for all accounts
		accounts, err = bot.Rpc.GetAllAccountIds()
		if err == nil && len(accounts) == 0 {
			cli.Logger.Errorf("There are no accounts yet, add a new account using the init subcommand")
		}
	} else {
		var accId deltachat.AccountId
		accId, err = cli.GetAccount(bot.Rpc, cli.SelectedAddr)
		accounts = []deltachat.AccountId{accId}
	}
	if err != nil {
		cli.Logger.Error(err)
		return
	}

	for _, accId := range accounts {
		addr, err := cli.GetAddress(bot.Rpc, accId)
		if err != nil {
			cli.Logger.Error(err)
			continue
		}
		fmt.Printf("Account #%v (%v):\n", accId, addr)
		adminForAcc(cli, bot, cmd, args, accId)
		fmt.Println("")
	}
}

func adminForAcc(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string, accId deltachat.AccountId) {
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
	var chatId deltachat.ChatId
	if reset {
		chatId, err = cli.ResetAdminChat(bot, accId)
	} else {
		chatId, err = cli.AdminChat(bot, accId)
	}
	if err != nil {
		cli.Logger.Errorf(errMsg, err)
		return
	}

	qrdata, _, err := bot.Rpc.GetChatSecurejoinQrCodeSvg(accId, option.Some(chatId))
	if err != nil {
		cli.Logger.Errorf(errMsg, err)
		return
	}

	fmt.Println("Scan this QR to become bot administrator")
	invert, _ := cmd.Flags().GetBool("invert")
	printQr(qrdata, invert)
	fmt.Println(qrdata)
}

func listCallback(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	if cli.SelectedAddr != "" {
		cli.Logger.Errorf("operation not supported for a single account, discard the -a/--account option and try again")
		return
	}

	accounts, err := bot.Rpc.GetAllAccountIds()
	if err != nil {
		cli.Logger.Error(err)
		return
	}
	for _, accId := range accounts {
		addr, err := cli.GetAddress(bot.Rpc, accId)
		if err != nil {
			cli.Logger.Error(err)
			continue
		}

		if isConf, _ := bot.Rpc.IsConfigured(accId); !isConf {
			addr = addr + " (not configured)"
		}
		fmt.Printf("#%v - %v\n", accId, addr)
	}
}

func removeCallback(cli *BotCli, bot *deltachat.Bot, cmd *cobra.Command, args []string) {
	var err error
	var accounts []deltachat.AccountId
	if cli.SelectedAddr == "" { // for all accounts
		accounts, err = bot.Rpc.GetAllAccountIds()
		if err == nil && len(accounts) == 0 {
			cli.Logger.Errorf("There are no accounts yet, add a new account using the init subcommand")
		}
	} else {
		var accId deltachat.AccountId
		accId, err = cli.GetAccount(bot.Rpc, cli.SelectedAddr)
		accounts = []deltachat.AccountId{accId}
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
		addr, err := cli.GetAddress(bot.Rpc, accId)
		if err != nil {
			cli.Logger.Error(err)
		}
		err = bot.Rpc.RemoveAccount(accId)
		if err != nil {
			cli.Logger.Error(err)
		} else {
			cli.Logger.Infof("Account #%v (%q) removed successfully.", accId, addr)
		}
	}
}
