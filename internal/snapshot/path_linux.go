//go:build linux

package snapshot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nonuplet/grimoire-archon/internal/config"
)

// resolveWinAppdata は RuntimeEnv に応じた Windows AppData の親ディレクトリを返します。
// subDir には "Local", "LocalLow", "Roaming", "Documents" を渡します。
func resolveWinAppdata(gameCfg *config.GameConfig, subDir string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリが取得できません: %w", err)
	}

	switch gameCfg.RuntimeEnv {
	case config.RuntimeEnvNative:
		return "", fmt.Errorf("linuxのネイティブ環境でAppDataを取得しようとしました。")
	case config.RuntimeEnvWine:
		// Wine: ~/.wine/drive_c/users/<user>/AppData/<subDir>
		winePrefix := os.Getenv("WINEPREFIX")
		if winePrefix == "" {
			winePrefix = filepath.Join(home, ".wine")
		}
		if subDir == "Documents" {
			return filepath.Join(winePrefix, "drive_c", "users", filepath.Base(home), "Documents"), nil
		}
		return filepath.Join(winePrefix, "drive_c", "users", filepath.Base(home), "AppData", subDir), nil

	case config.RuntimeEnvProton:
		// Proton: Steam の compatdata 以下
		if gameCfg.Steam == nil {
			return "", fmt.Errorf("proton 環境では steam 設定が必要です")
		}
		steamRoot := os.Getenv("STEAM_ROOT")
		if steamRoot == "" {
			steamRoot = filepath.Join(home, ".steam", "steam")
		}
		appID := gameCfg.Steam.AppID
		pfx := filepath.Join(steamRoot, "steamapps", "compatdata", appID, "pfx")
		if subDir == "Documents" {
			return filepath.Join(pfx, "drive_c", "users", "steamuser", "Documents"), nil
		}
		return filepath.Join(pfx, "drive_c", "users", "steamuser", "AppData", subDir), nil

	default:
		return "", fmt.Errorf("未知の RuntimeEnvが指定されています: %s", gameCfg.RuntimeEnv)
	}
}
