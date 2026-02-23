package usecase

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	"github.com/nonuplet/grimoire-archon/internal/config"
	"github.com/nonuplet/grimoire-archon/pkg"
)

// BackupUsecase backupのユースケース
type BackupUsecase struct{}

// NewBackupUsecase backupユースケースの生成
func NewBackupUsecase() *BackupUsecase {
	return &BackupUsecase{}
}

// Execute backupの実行
func (u *BackupUsecase) Execute(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) error {
	// 必要なコンフィグの情報があるかチェック
	if err := u.checkPreBackup(archonCfg, gameCfg); err != nil {
		return err
	}

	// バックアップディレクトリの存在確認と作成
	if err := u.checkAndCreateBackupDir(archonCfg, gameCfg); err != nil {
		return err
	}

	if err := u.startBackup(archonCfg, gameCfg); err != nil {
		return err
	}

	return nil
}

// checkPreBackup backupの処理前チェック
func (u *BackupUsecase) checkPreBackup(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) error {
	if archonCfg == nil {
		return fmt.Errorf("archonのコンフィグが定義されていません。")
	}
	if archonCfg.BackupDir == "" {
		return fmt.Errorf("バックアップディレクトリが設定されていません。")
	}
	if gameCfg.InstallDir == "" {
		return fmt.Errorf("インストールディレクトリが設定されていません。")
	}
	if len(gameCfg.BackupFiles) == 0 {
		return fmt.Errorf("バックアップ対象が指定されていません。")
	}

	// バックアップ対象のファイル・ディレクトリの存在確認
	for _, backupFile := range gameCfg.BackupFiles {
		targetPath := u.getBackupTargetPath(gameCfg, backupFile)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			return fmt.Errorf("バックアップ対象 '%s' が存在しないか、参照できません", targetPath)
		}
	}

	return nil
}

// checkAndCreateBackupDir バックアップディレクトリの存在確認と作成
func (u *BackupUsecase) checkAndCreateBackupDir(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) error {
	backupPath := filepath.Join(archonCfg.BackupDir, gameCfg.Name)

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		// Ask
		ok, askErr := pkg.AskYesNo(os.Stdin, fmt.Sprintf("バックアップディレクトリ '%s' が存在しません。作成しますか?", backupPath), true)
		if askErr != nil {
			return fmt.Errorf("バックアップディレクトリの作成確認に失敗しました: %w", askErr)
		}

		if !ok {
			return fmt.Errorf("バックアップディレクトリの作成がキャンセルされました。")
		}

		// 処理
		if err := os.MkdirAll(backupPath, 0o755); err != nil {
			return fmt.Errorf("バックアップディレクトリの作成に失敗しました: %w", err)
		}

		fmt.Printf("バックアップディレクトリ '%s' を作成しました。\n", backupPath)
	}

	return nil
}

func (u *BackupUsecase) getBackupTargetPath(gameCfg *config.GameConfig, targetFile string) string {
	// 絶対パスで指定されたとき
	if filepath.IsAbs(targetFile) {
		return targetFile
	}

	// 相対パスの場合はInstallDir以下を見る
	return filepath.Join(gameCfg.InstallDir, targetFile)
}

// startBackup バックアップ処理の実行
func (u *BackupUsecase) startBackup(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) error {
	// バックアップ先パスの設定
	backupPath := filepath.Join(archonCfg.BackupDir, gameCfg.Name)

	// tmpフォルダは必ず消す
	tmpDir := filepath.Join(backupPath, "tmp")
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "一時ディレクトリの削除に失敗しました: %v", err)
		}
	}(tmpDir)

	// 1. config.BackupData を構築し、/tmp/内のディレクトリにコピーするときの対応関係を作る
	backupData := &config.BackupData{
		Files: make([]config.BackupConfig, 0),
	}
	usedNames := make(map[string]int) // コピー先重複チェック

	for _, backupFile := range gameCfg.BackupFiles {
		originalPath := u.getBackupTargetPath(gameCfg, backupFile)
		baseName := filepath.Base(originalPath)

		// 2. コピー時にファイル名/ディレクトリ名が被ったら バックアップ時ファイル名を _1, _2, ...　とリネームする
		backupName := baseName
		if count, exists := usedNames[baseName]; exists {
			usedNames[baseName] = count + 1
			ext := filepath.Ext(baseName)
			nameWithoutExt := baseName[:len(baseName)-len(ext)]
			backupName = fmt.Sprintf("%s_%d%s", nameWithoutExt, count, ext)
		} else {
			usedNames[baseName] = 1
		}

		backupData.Files = append(backupData.Files, config.BackupConfig{
			OriginalPath: originalPath,
			BackupPath:   backupName,
		})
	}

	// 3. config.BackupData をもとに <backupPath>/tmp/ にコピー処理
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("一時ディレクトリの作成に失敗しました: %w", err)
	}

	for _, backupCfg := range backupData.Files {
		dst := filepath.Join(tmpDir, backupCfg.BackupPath)
		if err := pkg.CopyFileOrDir(backupCfg.OriginalPath, dst); err != nil {
			return fmt.Errorf("バックアップのコピーに失敗しました (%s -> %s): %w", backupCfg.OriginalPath, backupCfg.BackupPath, err)
		}
	}

	// 4. backupData をyamlにして tmp 内に一緒に保存する
	yamlData, err := yaml.Marshal(backupData)
	if err != nil {
		return fmt.Errorf("バックアップメタデータのYAML変換に失敗しました: %w", err)
	}

	metadataPath := filepath.Join(tmpDir, "metadata.yaml")
	// #nosec G306
	if err := os.WriteFile(metadataPath, yamlData, 0o755); err != nil {
		return fmt.Errorf("バックアップメタデータの保存に失敗しました: %w", err)
	}

	// 5. <backupPath>/tmp/ を圧縮し、<backupPath>/<gamename>_<timestamp>.zip に保存する
	timestamp := pkg.GetTimestamp()
	zipPath := filepath.Join(backupPath, fmt.Sprintf("%s_%s.zip", gameCfg.Name, timestamp))

	if err := pkg.Zip(tmpDir, zipPath); err != nil {
		return fmt.Errorf("ディレクトリの圧縮に失敗しました: %w", err)
	}

	// 一時ディレクトリの削除
	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("一時ディレクトリの削除に失敗しました: %w", err)
	}

	return nil
}
