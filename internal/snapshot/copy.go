package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nonuplet/grimoire-archon/internal/config"
	"github.com/nonuplet/grimoire-archon/internal/infra/storage"
)

// CopyResult はバックアップコピーの結果を保持します。
type CopyResult struct {
	Entries []FileEntry
}

// CopyToTmp は gameConfig で指定されたバックアップ対象ファイルを tmpDir にコピーします。
// コピーしたファイルの FileEntry 一覧を返します。
func CopyToTmp(tmpDir string, archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) (*CopyResult, error) {
	bt := gameCfg.BackupTargets
	if bt.IsEmpty() {
		return &CopyResult{}, nil
	}

	type targetSpec struct {
		resolvePath func(pattern string) (basePath string, err error)
		baseType    BaseType
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

	// Wine / Proton の AppData ディレクトリ解決。
	// archonCfg.AppdataDir が指定されている場合はそちらを起点とする。
	resolveWinDir := func(subDir string) func(string) (string, error) {
		return func(rel string) (string, error) {
			var base string
			if archonCfg.AppdataDir != "" {
				base = filepath.Join(archonCfg.AppdataDir, subDir)
			} else {
				var err error
				base, err = resolveWinAppdata(gameCfg, subDir)
				if err != nil {
					return "", err
				}
			}
			return filepath.Join(base, rel), nil
		}
	}

	specs := []targetSpec{
		{resolveInstallDir, BaseTypeInstallDir, bt.InstallDir},
		{resolveUserHome, BaseTypeUserHome, bt.UserHome},
		{resolveWinDir("Local"), BaseTypeAppdataLocal, bt.WinAppdataLocal},
		{resolveWinDir("LocalLow"), BaseTypeAppdataLocalLow, bt.WinAppdataLocalLow},
		{resolveWinDir("Roaming"), BaseTypeAppdataRoaming, bt.WinAppdataRoaming},
		{resolveWinDir("Documents"), BaseTypeWinDocuments, bt.WinDocuments},
		{resolveAbsolute, BaseTypeAbsolute, bt.Absolute},
	}

	var entries []FileEntry

	// 各タイプ(install_dir, user_home, ...)ごとに処理
	for _, spec := range specs {
		for _, pattern := range spec.patterns {
			// タイプごとのベースと合わせてパスを構築
			src, err := spec.resolvePath(pattern)
			if err != nil {
				return nil, fmt.Errorf("ベースパスの解決に失敗しました (type=%s, pattern=%s): %w", spec.baseType, pattern, err)
			}

			// コピーする
			dst := filepath.Join(tmpDir, string(spec.baseType), pattern)
			newEntries, err := copyEntries(src, dst, spec.baseType, pattern)
			if err != nil {
				return nil, err
			}
			entries = append(entries, newEntries...)
		}
	}

	return &CopyResult{Entries: entries}, nil
}

// copyEntries は src を dst へコピーし、FileEntry 一覧を返します。
// ファイル/ディレクトリの判定は util.CopyFileOrDir に委譲します。
func copyEntries(src, dst string, baseType BaseType, originalPath string) ([]FileEntry, error) {
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("コピー元のファイル/ディレクトリが存在しません (%s): %w", src, err)
		}
		return nil, fmt.Errorf("stat に失敗しました (%s): %w", src, err)
	}

	// ファイルの場合は dst/<basename> をコピー先ファイルパスとする
	copyDst := dst
	if !info.IsDir() {
		copyDst = filepath.Join(dst, filepath.Base(src))
	}

	if err := storage.CopyFileOrDir(src, copyDst); err != nil {
		return nil, fmt.Errorf("コピーに失敗しました: %w", err)
	}

	return []FileEntry{{
		ArchivePath:  filepath.ToSlash(filepath.Join(string(baseType), originalPath)),
		BaseType:     baseType,
		OriginalPath: originalPath,
		ModifiedAt:   info.ModTime().UTC().Truncate(time.Second),
	}}, nil
}
