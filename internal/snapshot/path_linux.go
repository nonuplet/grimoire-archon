//go:build linux

package snapshot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nonuplet/grimoire-archon/internal/domain"
)

// resolveWinAppdata は RuntimeEnv に応じた Windows AppData の親ディレクトリを返します。
// subDir には "Local", "LocalLow", "Roaming", "Documents" を渡します。
func resolveWinAppdata(archonCfg *domain.ArchonConfig, gameCfg *domain.GameConfig, subDir string) (string, error) {
	dirType := domain.WinDirectoryType(subDir)
	if err := dirType.GetWinDirType(); err != nil {
		return "", fmt.Errorf("サポートされていないディレクトリタイプ %s が指定されました", dirType)
	}

	// コンフィグで指定されていればオーバーライドする
	if dirType == domain.DirectoryDocuments && archonCfg.DocumentDir != "" {
		return archonCfg.DocumentDir, nil
	}
	if dirType != domain.DirectoryDocuments && archonCfg.AppdataDir != "" {
		return filepath.Join(archonCfg.AppdataDir, string(dirType)), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリが取得できません: %w", err)
	}

	switch gameCfg.RuntimeEnv {
	case "", domain.RuntimeEnvNative:
		return "", fmt.Errorf("linuxのネイティブ環境でAppDataを取得しようとしました。")

	case domain.RuntimeEnvWine:
		// Wine: ~/.wine/drive_c/users/<user>/AppData/<subDir>
		winePrefix := os.Getenv("WINEPREFIX")
		if winePrefix == "" {
			winePrefix = filepath.Join(home, ".wine")
		}
		if dirType == domain.DirectoryDocuments {
			return filepath.Join(winePrefix, "drive_c", "users", filepath.Base(home), string(domain.DirectoryDocuments)), nil
		}
		return filepath.Join(winePrefix, "drive_c", "users", filepath.Base(home), "AppData", string(dirType)), nil

	case domain.RuntimeEnvProton:
		compatDataPath := os.Getenv("STEAM_COMPAT_DATA_PATH")

		// 環境変数で STEAM_COMPAT_DATA_PATH が指定されている場合(Protonを手動実行している場合)
		if compatDataPath != "" {
			pfx := filepath.Join(compatDataPath, "pfx")
			if dirType == domain.DirectoryDocuments {
				return filepath.Join(pfx, "drive_c", "users", "steamuser", string(domain.DirectoryDocuments)), nil
			}
			return filepath.Join(pfx, "drive_c", "users", "steamuser", "AppData", string(dirType)), nil
		}

		// そうでない場合(Steamクライアントを使っている場合)
		// ProtonはデフォルトでSteam の compatdata 以下に保存する
		if gameCfg.Steam == nil || gameCfg.Steam.AppID == "" {
			return "", fmt.Errorf("proton 環境では steam 設定が必要です")
		}

		steamRoot := os.Getenv("STEAM_ROOT")
		if steamRoot == "" {
			steamRoot = filepath.Join(home, ".steam", "steam")
		}

		pfx := filepath.Join(steamRoot, "steamapps", "compatdata", gameCfg.Steam.AppID, "pfx")
		if dirType == domain.DirectoryDocuments {
			return filepath.Join(pfx, "drive_c", "users", "steamuser", string(domain.DirectoryDocuments)), nil
		}
		return filepath.Join(pfx, "drive_c", "users", "steamuser", "AppData", string(dirType)), nil

	default:
		return "", fmt.Errorf("未知の RuntimeEnvが指定されています: %s", gameCfg.RuntimeEnv)
	}
}
