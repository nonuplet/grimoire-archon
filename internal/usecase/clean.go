package usecase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nonuplet/grimoire-archon/internal/config"
	"github.com/nonuplet/grimoire-archon/pkg"
)

type backupCheckType int

const (
	backupCheckNoDir backupCheckType = iota
	backupCheckNotExist
	backupCheckOld
	backupCheckOk
)

// CleanUsecase cleanのユースケース
type CleanUsecase struct{}

// NewCleanUsecase cleanのユースケースのインスタンスを生成する
func NewCleanUsecase() *CleanUsecase {
	return &CleanUsecase{}
}

// Execute cleanの実行
func (u *CleanUsecase) Execute(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) error {
	// コンフィグのチェック
	if err := u.checkPreClean(archonCfg, gameCfg); err != nil {
		return err
	}

	// 削除前にバックアップがあるかチェック
	fmt.Printf("%s の削除前チェック中...\n", gameCfg.Name)
	backupCheck, err := u.checkBackup(archonCfg, gameCfg)
	if err != nil {
		return fmt.Errorf("バックアップチェックに失敗しました: %w", err)
	}

	//　ユーザに確認
	var ok bool
	switch backupCheck {
	case backupCheckNoDir, backupCheckNotExist:
		ok, err = pkg.AskYesNo(os.Stdin, "バックアップが存在しません！本当に削除してもよろしいですか？", false)
		if err != nil {
			return fmt.Errorf("確認に失敗しました: %w", err)
		}
	case backupCheckOld:
		ok, err = pkg.AskYesNo(os.Stdin, "古いバックアップしかありませんが、本当に削除してもよろしいですか？", false)
		if err != nil {
			return fmt.Errorf("確認に失敗しました: %w", err)
		}
	case backupCheckOk:
		ok, err = pkg.AskYesNo(os.Stdin, fmt.Sprintf("%s を削除してもよろしいですか？", gameCfg.Name), true)
		if err != nil {
			return fmt.Errorf("確認に失敗しました: %w", err)
		}
	}

	// バックアップの状態が良くない場合、しつこく再確認
	if backupCheck != backupCheckOk && ok {
		ok, err = pkg.AskYesNo(os.Stdin, "本当に大丈夫？消しますよ？", false)
		if err != nil {
			return fmt.Errorf("確認に失敗しました: %w", err)
		}
	}

	if !ok {
		return fmt.Errorf("%s の削除処理をキャンセルしました。", gameCfg.Name)
	}

	// ユーザに確認
	fmt.Printf("%s の削除処理を実行します...\n", gameCfg.Name)
	err = pkg.ClearDirectoryContents(gameCfg.InstallDir)
	if err != nil {
		return fmt.Errorf("削除処理に失敗しました: %w", err)
	}

	return nil
}

// checkPreClean cleanの処理前チェック
func (u *CleanUsecase) checkPreClean(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) error {
	if archonCfg == nil {
		return fmt.Errorf("archonのコンフィグが定義されていません。")
	}
	if gameCfg.InstallDir == "" {
		return fmt.Errorf("インストールディレクトリが設定されていません。")
	}

	return nil
}

// checkBackup バックアップが存在するかチェック
func (u *CleanUsecase) checkBackup(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) (backupCheckType, error) {
	backupPath := filepath.Join(archonCfg.BackupDir, gameCfg.Name)

	// バックアップディレクトリをチェック
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		ok, err := askAndBackup(archonCfg, gameCfg, fmt.Sprintf("バックアップ先 '%s' が存在しません。バックアップしますか？", backupPath))
		if err != nil {
			return -1, fmt.Errorf("バックアップに失敗しました: %w", err)
		}
		if !ok {
			return backupCheckNoDir, nil
		}
		return backupCheckOk, nil
	}

	// バックアップディレクトリ内に .zip ファイルがあるかチェック
	zipPattern := filepath.Join(backupPath, "*.zip")
	matches, err := filepath.Glob(zipPattern)
	if err != nil {
		return -1, fmt.Errorf("バックアップファイルの確認に失敗しました: %w", err)
	}

	// バックアップディレクトリはあるが、.zipファイルが見つからない場合
	if len(matches) > 0 {
		ok, zipErr := askAndBackup(archonCfg, gameCfg, "バックアップ先にzipが一つもありません。バックアップしますか？")
		if zipErr != nil {
			return -1, fmt.Errorf("バックアップに失敗しました: %w", err)
		}
		if !ok {
			return backupCheckNotExist, nil
		}
		return backupCheckOk, nil
	}

	// 24時間に作成されたバックアップがあるかチェック
	exist, err := checkBackupZip(backupPath, gameCfg.Name)
	if err != nil {
		return -1, fmt.Errorf("バックアップファイルの確認に失敗しました: %w", err)
	}
	if !exist {
		ok, err := askAndBackup(archonCfg, gameCfg, "24時間以内に作成されたバックアップがないようです。バックアップしますか？")
		if err != nil {
			return -1, fmt.Errorf("バックアップに失敗しました: %w", err)
		}
		if !ok {
			return backupCheckOld, nil
		}
	}

	return backupCheckOk, nil
}

// askAndBackup バックアップの確認と実行
func askAndBackup(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig, msg string) (bool, error) {
	ok, err := pkg.AskYesNo(os.Stdin, msg, true)
	if err != nil {
		return false, fmt.Errorf("確認に失敗しました: %w", err)
	}

	// バックアップをしない
	if !ok {
		return false, nil
	}

	// バックアップする
	backupUsecase := NewBackupUsecase()
	err = backupUsecase.Execute(archonCfg, gameCfg)
	if err != nil {
		return false, fmt.Errorf("バックアップに失敗しました: %w", err)
	}
	return true, nil
}

// checkBackupZip 指定されたディレクトリ内に、24時間以内に作成されたバックアップZIPが存在するか確認
func checkBackupZip(backupPath, gameName string) (bool, error) {
	files, err := os.ReadDir(backupPath)
	if err != nil {
		return false, fmt.Errorf("ディレクトリの読み込みに失敗しました: %w", err)
	}

	prefix := gameName + "_"
	suffix := ".zip"
	layout := "20060102_150405"
	now := time.Now()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) {
			continue
		}

		// タイムスタンプ部分を検証
		tsStr := strings.TrimPrefix(name, prefix)
		tsStr = strings.TrimSuffix(tsStr, suffix)
		if _, err := time.ParseInLocation(layout, tsStr, time.Local); err != nil {
			continue
		}

		// ファイルの情報を取得
		info, err := file.Info()
		if err != nil {
			continue
		}

		// 更新日時が24時間以内か
		if now.Sub(info.ModTime()) <= 24*time.Hour {
			return true, nil
		}
	}

	return false, nil
}
