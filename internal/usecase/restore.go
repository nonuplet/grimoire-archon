package usecase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nonuplet/grimoire-archon/internal/config"
	"github.com/nonuplet/grimoire-archon/internal/infra/storage"
	"github.com/nonuplet/grimoire-archon/internal/snapshot"
)

// RestoreUsecase restoreのユースケース
type RestoreUsecase struct{}

// NewRestoreUsecase restoreのユースケースを作成
func NewRestoreUsecase() *RestoreUsecase {
	return &RestoreUsecase{}
}

// Execute restoreの実行
func (u *RestoreUsecase) Execute(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig, zipPath string) error {
	if err := u.checkPreRestore(archonCfg, gameCfg, zipPath); err != nil {
		return err
	}

	if err := snapshot.CheckAndCreateSnapshotDir(archonCfg, gameCfg); err != nil {
		return fmt.Errorf("展開用ディレクトリの作成に失敗しました: %w", err)
	}

	if err := u.restoreSnapshot(archonCfg, gameCfg, zipPath); err != nil {
		return err
	}

	return nil
}

// checkPreRestore restore前チェック
func (u *RestoreUsecase) checkPreRestore(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig, zipPath string) error {
	if archonCfg == nil {
		return fmt.Errorf("archonのコンフィグが定義されていません。")
	}
	if archonCfg.BackupDir == "" {
		return fmt.Errorf("バックアップ先が設定されていません。")
	}
	if gameCfg.InstallDir == "" {
		return fmt.Errorf("%s にインストール先が設定されていません。", gameCfg.Name)
	}

	// バックアップ指定したファイルの数を確認
	if gameCfg.BackupTargets.IsEmpty() {
		return fmt.Errorf("%s にバックアップの対象が指定されていません。", gameCfg.Name)
	}

	if !storage.IsZipFile(zipPath) {
		return fmt.Errorf("指定したファイル %s はzipファイルではありません", zipPath)
	}

	return nil
}

func (u *RestoreUsecase) restoreSnapshot(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig, zipPath string) error {
	// バックアップ先パスの設定
	snapshotPath := filepath.Join(archonCfg.BackupDir, gameCfg.Name)

	// tmpフォルダを作成 作業が終わったら成功しても失敗しても消す
	tmpDir := filepath.Join(snapshotPath, "tmp")
	if err := storage.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("一時ディレクトリの作成に失敗しました: %w", err)
	}
	defer func(path string) {
		err := storage.RemoveAll(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "一時ディレクトリの削除に失敗しました: %v", err)
		}
	}(tmpDir)

	// unzipしたあとのディレクトリ指定に使う
	archiveName := strings.TrimSuffix(filepath.Base(zipPath), filepath.Ext(zipPath))
	archiveDir := filepath.Join(tmpDir, archiveName)

	// <backup_dir>/<game_name>/tmp/ に展開
	fmt.Printf("zipファイル '%s' を展開しています... \n", filepath.Join(archonCfg.BackupDir, gameCfg.Name))
	if err := storage.Unzip(zipPath, tmpDir); err != nil {
		return fmt.Errorf("zipファイルの展開に失敗しました: %w", err)
	}

	// リストア
	if err := snapshot.RestoreFromTmp(archonCfg, gameCfg, archiveDir); err != nil {
		return fmt.Errorf("リストアに失敗しました: %w", err)
	}

	return nil
}
