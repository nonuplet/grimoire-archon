//go:build linux

package snapshot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nonuplet/grimoire-archon/internal/config"
)

// DirectoryType はWindows関連ディレクトリ(AppData, Document)の種類を表す型
type DirectoryType string

const (
	// DirectoryLocal AppData/Local
	DirectoryLocal DirectoryType = "Local"
	// DirectoryLocalLow AppData/LocalLow
	DirectoryLocalLow DirectoryType = "LocalLow"
	// DirectoryRoaming AppData/Roaming
	DirectoryRoaming DirectoryType = "Roaming"
	// DirectoryDocuments Documents
	DirectoryDocuments DirectoryType = "Documents"
)

// Valid ディレクトリタイプが有効かどうかを検証
func (d DirectoryType) Valid() error {
	switch d {
	case DirectoryLocal, DirectoryLocalLow, DirectoryRoaming, DirectoryDocuments:
		return nil
	default:
		return fmt.Errorf("サポートされていないディレクトリタイプ %s が指定されました", d)
	}
}

// resolveWinAppdata は RuntimeEnv に応じた Windows AppData の親ディレクトリを返します。
// subDir には "Local", "LocalLow", "Roaming", "Documents" を渡します。
func resolveWinAppdata(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig, subDir string) (string, error) {
	dirType := DirectoryType(subDir)
	if err := dirType.Valid(); err != nil {
		return "", fmt.Errorf("サポートされていないディレクトリタイプ %s が指定されました", dirType)
	}

	// コンフィグで指定されていればオーバーライドする
	if dirType == DirectoryDocuments && archonCfg.DocumentDir != "" {
		return archonCfg.DocumentDir, nil
	}
	if dirType != DirectoryDocuments && archonCfg.AppdataDir != "" {
		return filepath.Join(archonCfg.AppdataDir, string(dirType)), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリが取得できません: %w", err)
	}

	switch gameCfg.RuntimeEnv {
	case "", config.RuntimeEnvNative:
		return "", fmt.Errorf("linuxのネイティブ環境でAppDataを取得しようとしました。")

	case config.RuntimeEnvWine:
		// Wine: ~/.wine/drive_c/users/<user>/AppData/<subDir>
		winePrefix := os.Getenv("WINEPREFIX")
		if winePrefix == "" {
			winePrefix = filepath.Join(home, ".wine")
		}
		if dirType == DirectoryDocuments {
			return filepath.Join(winePrefix, "drive_c", "users", filepath.Base(home), string(DirectoryDocuments)), nil
		}
		return filepath.Join(winePrefix, "drive_c", "users", filepath.Base(home), "AppData", string(dirType)), nil

	case config.RuntimeEnvProton:
		compatDataPath := os.Getenv("STEAM_COMPAT_DATA_PATH")

		// 環境変数で STEAM_COMPAT_DATA_PATH が指定されている場合(Protonを手動実行している場合)
		if compatDataPath != "" {
			pfx := filepath.Join(compatDataPath, "pfx")
			if dirType == DirectoryDocuments {
				return filepath.Join(pfx, "drive_c", "users", "steamuser", string(DirectoryDocuments)), nil
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
		if dirType == DirectoryDocuments {
			return filepath.Join(pfx, "drive_c", "users", "steamuser", string(DirectoryDocuments)), nil
		}
		return filepath.Join(pfx, "drive_c", "users", "steamuser", "AppData", string(dirType)), nil

	default:
		return "", fmt.Errorf("未知の RuntimeEnvが指定されています: %s", gameCfg.RuntimeEnv)
	}
}
