package botcli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mdp/qrterminal/v3"
)

// Generate an invite link (https://i.delta.chat) for the given invitation-QR data
func GenerateInviteLink(qrdata string) string {
	fragment := strings.Replace(strings.SplitN(qrdata, ":", 2)[1], "#", "&", 1)
	return "https://i.delta.chat/#" + fragment
}

func getDefaultAppDir(appName string) string {
	cfgDir, _ := os.UserConfigDir()
	return filepath.Join(cfgDir, appName)
}

func getAccountsDir(appDir string) string {
	return filepath.Join(appDir, "accounts")
}

func printQr(qrdata string, invert bool) {
	config := qrterminal.Config{
		Level:          qrterminal.M,
		Writer:         os.Stdout,
		HalfBlocks:     true,
		BlackChar:      qrterminal.BLACK_BLACK,
		WhiteBlackChar: qrterminal.WHITE_BLACK,
		WhiteChar:      qrterminal.WHITE_WHITE,
		BlackWhiteChar: qrterminal.BLACK_WHITE,
		QuietZone:      4,
	}
	if invert {
		config.BlackChar = qrterminal.WHITE_WHITE
		config.WhiteBlackChar = qrterminal.BLACK_WHITE
		config.WhiteChar = qrterminal.BLACK_BLACK
		config.BlackWhiteChar = qrterminal.WHITE_BLACK
	}
	qrterminal.GenerateWithConfig(qrdata, config)
}
