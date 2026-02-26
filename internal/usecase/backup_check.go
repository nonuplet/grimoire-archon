package usecase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nonuplet/grimoire-archon/internal/domain"
)

// Check はバックアップの状態を確認を行い、必要があればバックアップを実行する
// cleanの前処理として実行される
func (u *BackupUsecase) Check() error {
	// 削除前にバックアップがあるかチェック
	condition, err := u.checkBackupCondition()
	if err != nil {
		return fmt.Errorf("バックアップチェックに失敗しました: %w", err)
	}

	// ユーザに確認
	var ok bool
	switch condition {
	case domain.SnapshotDirNotFound, domain.SnapshotFileNotFound:
		ok, err = u.cli.AskYesNo(os.Stdin, "バックアップが存在しません！本当に削除してもよろしいですか？", false)
		if err != nil {
			return fmt.Errorf("確認に失敗しました: %w", err)
		}
	case domain.SnapshotFileTooOld:
		ok, err = u.cli.AskYesNo(os.Stdin, "古いバックアップしかありませんが、本当に削除してもよろしいですか？", false)
		if err != nil {
			return fmt.Errorf("確認に失敗しました: %w", err)
		}
	case domain.SnapshotHealthy:
		ok, err = u.cli.AskYesNo(os.Stdin, fmt.Sprintf("%s を削除してもよろしいですか？", u.gameCfg.Name), true)
		if err != nil {
			return fmt.Errorf("確認に失敗しました: %w", err)
		}
	}

	// バックアップの状態が良くない場合、しつこく再確認
	if condition != domain.SnapshotHealthy && ok {
		ok, err = u.cli.AskYesNo(os.Stdin, "本当に大丈夫？消しますよ？", false)
		if err != nil {
			return fmt.Errorf("確認に失敗しました: %w", err)
		}
	}

	if !ok {
		return fmt.Errorf("%s の削除処理をキャンセルしました。", u.gameCfg.Name)
	}

	return nil
}

// checkBackupCondition 24時間以内に作成されたzipファイルが、バックアップディレクトリ内にあるかチェック
func (u *BackupUsecase) checkBackupCondition() (domain.SnapshotCondition, error) {
	backupPath := filepath.Join(u.archonCfg.BackupDir, u.gameCfg.Name)

	// バックアップディレクトリをチェック
	if _, err := u.fs.Stat(backupPath); os.IsNotExist(err) {
		ok, err := u.askAndBackup(fmt.Sprintf("バックアップ先 '%s' が存在しません。バックアップしますか？", backupPath))
		if err != nil {
			return -1, fmt.Errorf("バックアップに失敗しました: %w", err)
		}
		if !ok {
			return domain.SnapshotDirNotFound, nil
		}
		return domain.SnapshotHealthy, nil
	}

	// バックアップディレクトリ内に .zip ファイルがあるかチェック
	zipPattern := filepath.Join(backupPath, "*.zip")
	matches, err := filepath.Glob(zipPattern)
	if err != nil {
		return -1, fmt.Errorf("バックアップファイルの確認に失敗しました: %w", err)
	}

	// バックアップディレクトリはあるが、.zipファイルが見つからない場合
	if len(matches) > 0 {
		ok, zipErr := u.askAndBackup("バックアップ先にzipが一つもありません。バックアップしますか？")
		if zipErr != nil {
			return -1, fmt.Errorf("バックアップに失敗しました: %w", err)
		}
		if !ok {
			return domain.SnapshotFileNotFound, nil
		}
		return domain.SnapshotHealthy, nil
	}

	// 24時間に作成されたバックアップがあるかチェック
	exist, err := u.checkBackupZip(backupPath, u.gameCfg.Name)
	if err != nil {
		return -1, fmt.Errorf("バックアップファイルの確認に失敗しました: %w", err)
	}
	if !exist {
		ok, err := u.askAndBackup("24時間以内に作成されたバックアップがないようです。バックアップしますか？")
		if err != nil {
			return -1, fmt.Errorf("バックアップに失敗しました: %w", err)
		}
		if !ok {
			return domain.SnapshotFileTooOld, nil
		}
	}

	return domain.SnapshotHealthy, nil
}

// askAndBackup バックアップの確認と実行
func (u *BackupUsecase) askAndBackup(msg string) (bool, error) {
	ok, err := u.cli.AskYesNo(os.Stdin, msg, true)
	if err != nil {
		return false, fmt.Errorf("確認に失敗しました: %w", err)
	}

	// バックアップをしない
	if !ok {
		return false, nil
	}

	// バックアップする
	err = u.Execute()
	if err != nil {
		return false, fmt.Errorf("バックアップに失敗しました: %w", err)
	}
	return true, nil
}

// checkBackupZip 指定されたディレクトリ内に、24時間以内に作成されたバックアップZIPが存在するか確認
func (u *BackupUsecase) checkBackupZip(backupPath, gameName string) (bool, error) {
	files, err := u.fs.ReadDir(backupPath)
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
