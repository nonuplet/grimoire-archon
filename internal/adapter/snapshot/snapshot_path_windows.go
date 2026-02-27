//go:build windows

package snapshot

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows"

	"github.com/nonuplet/grimoire-archon/internal/domain"
)

// resolveWinAppdata は RuntimeEnv に応じた Windows AppData の親ディレクトリを返します。
// subDir には "Local", "LocalLow", "Roaming", "Documents" を渡します。
func (snap Snapshot) resolveWinAppdata(subDir string) (string, error) {
	dirType := domain.WinDirectoryType(subDir)
	if err := dirType.GetWinDirType(); err != nil {
		return "", fmt.Errorf("サポートされていないディレクトリタイプ %s が指定されました", dirType)
	}

	// コンフィグで指定されていればオーバーライドする
	if dirType == domain.DirectoryDocuments && snap.archonCfg.DocumentDir != "" {
		return snap.archonCfg.DocumentDir, nil
	}
	if dirType != domain.DirectoryDocuments && snap.archonCfg.AppdataDir != "" {
		return filepath.Join(snap.archonCfg.AppdataDir, string(dirType)), nil
	}

	switch snap.gameCfg.RuntimeEnv {
	case domain.RuntimeEnvWine, domain.RuntimeEnvProton:
		return "", fmt.Errorf("Windows環境で wine または proton が選択されています。")

	case "", domain.RuntimeEnvNative:
		var id *windows.KNOWNFOLDERID
		switch dirType {
		case domain.DirectoryLocal:
			id = windows.FOLDERID_LocalAppData
		case domain.DirectoryLocalLow:
			id = windows.FOLDERID_LocalAppDataLow
		case domain.DirectoryRoaming:
			id = windows.FOLDERID_RoamingAppData
		case domain.DirectoryDocuments:
			id = windows.FOLDERID_Documents
		}
		path, err := windows.KnownFolderPath(id, 0)
		if err != nil {
			return "", fmt.Errorf("Windows関連ディレクトリ %s の取得に失敗しました: %w", subDir, err)
		}
		return path, nil

	default:
		return "", fmt.Errorf("未知の RuntimeEnvが指定されています: %s", snap.gameCfg.RuntimeEnv)
	}
}
