package snapshot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nonuplet/grimoire-archon/internal/domain"
)

// pathResolver は BaseType ごとのパス解決関数の型です。
type pathResolver = func(pattern string) (string, error)

// buildResolvers は BaseType ごとのパス解決関数マップを返します。
func (snap Snapshot) buildResolvers() map[domain.BaseType]pathResolver {
	resolveInstallDir := func(rel string) (string, error) {
		return filepath.Join(snap.gameCfg.InstallDir, rel), nil
	}

	resolveUserHome := func(rel string) (string, error) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("ユーザーホームディレクトリの取得に失敗しました: %w", err)
		}
		return filepath.Join(home, rel), nil
	}

	resolveAbsolute := func(abs string) (string, error) {
		return abs, nil
	}

	resolveWinDir := func(subDir string) pathResolver {
		return func(rel string) (string, error) {
			base, err := snap.resolveWinAppdata(subDir)
			if err != nil {
				return "", err
			}
			return filepath.Join(base, rel), nil
		}
	}

	return map[domain.BaseType]pathResolver{
		domain.BaseTypeInstallDir:      resolveInstallDir,
		domain.BaseTypeUserHome:        resolveUserHome,
		domain.BaseTypeAppdataLocal:    resolveWinDir("Local"),
		domain.BaseTypeAppdataLocalLow: resolveWinDir("LocalLow"),
		domain.BaseTypeAppdataRoaming:  resolveWinDir("Roaming"),
		domain.BaseTypeWinDocuments:    resolveWinDir("Documents"),
		domain.BaseTypeAbsolute:        resolveAbsolute,
	}
}
