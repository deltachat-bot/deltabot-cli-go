package botcli

import (
	"os"
	"path/filepath"
)

func getDefaultAppDir(appName string) string {
	cfgDir, _ := os.UserConfigDir()
	return filepath.Join(cfgDir, appName)
}

func getAccountsDir(appDir string) string {
	return filepath.Join(appDir, "accounts")
}
