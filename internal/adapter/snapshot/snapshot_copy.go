package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nonuplet/grimoire-archon/internal/domain"
	"github.com/nonuplet/grimoire-archon/internal/infra/filesystem"
)

// CopyToTmp は gameConfig で指定されたバックアップ対象ファイルを tmpDir にコピーします。
// コピーしたファイルの FileEntry 一覧を返します。
func CopyToTmp(tmpDir string, archonCfg *domain.ArchonConfig, gameCfg *domain.GameConfig) ([]domain.FileEntry, error) {
	bt := gameCfg.BackupTargets
	if bt.IsEmpty() {
		return nil, nil
	}

	type targetSpec struct {
		resolvePath func(pattern string) (basePath string, err error)
		baseType    domain.BaseType
		patterns    []string
	}

	// 各 BaseType のソースディレクトリ解決関数を定義
	resolveInstallDir := func(rel string) (string, error) {
		return filepath.Join(gameCfg.InstallDir, rel), nil
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

	// Windows関連ディレクトリ(AppData, Document)解決
	resolveWinDir := func(subDir string) func(string) (string, error) {
		return func(rel string) (string, error) {
			base, err := resolveWinAppdata(archonCfg, gameCfg, subDir)
			if err != nil {
				return "", err
			}
			return filepath.Join(base, rel), nil
		}
	}

	specs := []targetSpec{
		{resolveInstallDir, domain.BaseTypeInstallDir, bt.InstallDir},
		{resolveUserHome, domain.BaseTypeUserHome, bt.UserHome},
		{resolveWinDir("Local"), domain.BaseTypeAppdataLocal, bt.WinAppdataLocal},
		{resolveWinDir("LocalLow"), domain.BaseTypeAppdataLocalLow, bt.WinAppdataLocalLow},
		{resolveWinDir("Roaming"), domain.BaseTypeAppdataRoaming, bt.WinAppdataRoaming},
		{resolveWinDir("Documents"), domain.BaseTypeWinDocuments, bt.WinDocuments},
		{resolveAbsolute, domain.BaseTypeAbsolute, bt.Absolute},
	}

	var entries []domain.FileEntry

	// 各タイプ(install_dir, user_home, ...)ごとに処理
	for _, spec := range specs {
		for _, pattern := range spec.patterns {
			// タイプごとのベースと合わせてsrc/dstパスを構築
			src, err := spec.resolvePath(pattern)
			if err != nil {
				return nil, fmt.Errorf("ベースパスの解決に失敗しました (type=%s, pattern=%s): %w", spec.baseType, pattern, err)
			}
			dst := filepath.Join(tmpDir, string(spec.baseType), pattern)

			// コピーする
			newEntry, err := copyEntries(src, dst, spec.baseType, pattern)
			if err != nil {
				return nil, err
			}
			entries = append(entries, newEntry)
		}
	}

	return entries, nil
}

// copyEntries は src を dst へコピーし、FileEntry 一覧を返します。
// ファイル/ディレクトリの判定は util.CopyFileOrDir に委譲します。
func copyEntries(src, dst string, baseType domain.BaseType, originalPath string) (domain.FileEntry, error) {
	// コピー元のinfo取得
	info, err := filesystem.GetInfo(src)
	if err != nil {
		return domain.FileEntry{}, fmt.Errorf("コピー元のファイル/ディレクトリの情報取得に失敗しました: %w", err)
	}

	// コピー
	if err := filesystem.CopyFileOrDir(src, dst, false); err != nil {
		return domain.FileEntry{}, fmt.Errorf("コピーに失敗しました: %w", err)
	}

	return domain.FileEntry{
		ArchivePath:  filepath.ToSlash(filepath.Join(string(baseType), originalPath)),
		BaseType:     baseType,
		OriginalPath: originalPath,
		ModifiedAt:   info.ModTime().UTC().Truncate(time.Second),
	}, nil
}
