package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/goccy/go-yaml"

	"github.com/nonuplet/grimoire-archon/internal/domain"
)

// RestoreFromTmp は archiveDir で指定されたディレクトリ内のバックアップデータをリストアします。
func (snap Snapshot) RestoreFromTmp(archiveDir string) error {
	// metadata.yamlのロード
	metaYaml := filepath.Join(archiveDir, "metadata.yaml")
	meta, err := snap.loadMetaData(metaYaml)
	if err != nil {
		return fmt.Errorf("metadata.yamlのロードに失敗しました: %w", err)
	}

	// OSの違いをチェック
	if osErr := snap.checkDifferentOs(meta.Os); osErr != nil {
		return osErr
	}

	// バックアップリストにないファイルをリストアップ
	notDefined := snap.getNotDefinedFiles(meta)
	if len(notDefined) > 0 {
		fmt.Printf("コンフィグに設定した BackupTargets 以外のファイルが見つかりました。\n\n")
		for _, file := range notDefined {
			fmt.Printf("- %s: %s\n", file.BaseType, file.OriginalPath)
		}
		ok, err := snap.cli.AskYesNo(os.Stdin, "\nリストアを続行してもよろしいですか？", true)
		if err != nil {
			return fmt.Errorf("ユーザーの回答取得に失敗しました: %w", err)
		}
		if !ok {
			return fmt.Errorf("リストアを中止しました")
		}
	}

	// 上書きがあるかチェック
	overwriteFiles := snap.getOverwriteFiles(meta)
	if len(overwriteFiles) > 0 {
		fmt.Printf("以下のファイルは上書きされます。\n\n")
		for _, file := range overwriteFiles {
			fmt.Printf("- %s: %s\n", file.BaseType, file.OriginalPath)
		}
		ok, err := snap.cli.AskYesNo(os.Stdin, "\n対象のファイルを上書きしてもよろしいですか？", true)
		if err != nil {
			return fmt.Errorf("ユーザーの回答取得に失敗しました: %w", err)
		}
		if !ok {
			return fmt.Errorf("リストアを中止しました")
		}
	}

	fmt.Println("バックアップで復元しています...")
	if err := snap.copyArchivedFiles(meta, archiveDir); err != nil {
		return fmt.Errorf("復元に失敗しました: %w", err)
	}

	return nil
}

// loadMetaData 指定したパスからメタデータをロード
func (snap Snapshot) loadMetaData(path string) (*domain.Metadata, error) {
	if _, err := snap.fs.Stat(path); err != nil {
		return nil, fmt.Errorf("metadata.yamlの取得に失敗しました: %w", err)
	}

	data, err := snap.fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("metadata.yamlのリードに失敗しました: %w", err)
	}

	var meta domain.Metadata
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("metadata.yamlのデコードに失敗しました: %w", err)
	}
	return &meta, nil
}

func (snap Snapshot) checkDifferentOs(archivedOs string) error {
	currentOs := runtime.GOOS
	if currentOs != archivedOs {
		q := fmt.Sprintf("バックアップ元のOS(%s)と現在のOS(%s)が異なります。実行しますか?", archivedOs, currentOs)
		ok, err := snap.cli.AskYesNo(os.Stdin, q, false)
		if err != nil {
			return fmt.Errorf("ユーザーの回答取得に失敗しました: %w", err)
		}
		if !ok {
			return fmt.Errorf("リストアを中止しました")
		}
	}
	return nil
}

// getNotDefinedFiles はバックアップリストにないファイルを取得します。
// archivedEntries のうち snap.gameCfg.BackupTargets に定義されていないエントリを返します。
func (snap Snapshot) getNotDefinedFiles(meta *domain.Metadata) []domain.FileEntry {
	archivedEntries := meta.Files
	targets := snap.gameCfg.BackupTargets

	// BaseType ごとに対応するターゲットリストへマッピング
	targetMap := map[domain.BaseType][]string{
		domain.BaseTypeInstallDir:      snap.getTargetList(targets, domain.BaseTypeInstallDir),
		domain.BaseTypeUserHome:        snap.getTargetList(targets, domain.BaseTypeUserHome),
		domain.BaseTypeAppdataLocal:    snap.getTargetList(targets, domain.BaseTypeAppdataLocal),
		domain.BaseTypeAppdataLocalLow: snap.getTargetList(targets, domain.BaseTypeAppdataLocalLow),
		domain.BaseTypeAppdataRoaming:  snap.getTargetList(targets, domain.BaseTypeAppdataRoaming),
		domain.BaseTypeWinDocuments:    snap.getTargetList(targets, domain.BaseTypeWinDocuments),
		domain.BaseTypeAbsolute:        snap.getTargetList(targets, domain.BaseTypeAbsolute),
	}

	var notDefined []domain.FileEntry
	for _, entry := range archivedEntries {
		list, ok := targetMap[entry.BaseType]
		if !ok {
			// 未知の BaseType はターゲット未定義とみなす
			notDefined = append(notDefined, entry)
			continue
		}

		found := false
		for _, t := range list {
			if t == entry.OriginalPath {
				found = true
				break
			}
		}
		if !found {
			notDefined = append(notDefined, entry)
		}
	}

	return notDefined
}

// getTargetList は BackupTargetConfig から指定した BaseType に対応するパスリストを返します。
func (snap Snapshot) getTargetList(targets *domain.BackupTargetConfig, baseType domain.BaseType) []string {
	if targets == nil {
		return nil
	}
	switch baseType {
	case domain.BaseTypeInstallDir:
		return targets.InstallDir
	case domain.BaseTypeUserHome:
		return targets.UserHome
	case domain.BaseTypeAppdataLocal:
		return targets.WinAppdataLocal
	case domain.BaseTypeAppdataLocalLow:
		return targets.WinAppdataLocalLow
	case domain.BaseTypeAppdataRoaming:
		return targets.WinAppdataRoaming
	case domain.BaseTypeWinDocuments:
		return targets.WinDocuments
	case domain.BaseTypeAbsolute:
		return targets.Absolute
	default:
		return nil
	}
}

// getOverwriteFiles はバックアップ元のファイルリストから、上書き対象のファイルリストを返します。
func (snap Snapshot) getOverwriteFiles(meta *domain.Metadata) []domain.FileEntry {
	var overwriteFiles []domain.FileEntry
	if meta == nil || len(meta.Files) == 0 {
		return overwriteFiles
	}

	resolvers := snap.buildResolvers()

	for _, entry := range meta.Files {
		resolver, ok := resolvers[entry.BaseType]
		if !ok {
			continue
		}
		dst, err := resolver(entry.OriginalPath)
		if err != nil {
			continue
		}

		if _, err := snap.fs.Stat(dst); err == nil {
			overwriteFiles = append(overwriteFiles, entry)
		} else if !os.IsNotExist(err) {
			// 参照権限等の理由で確認できない場合も、上書き対象として扱う
			overwriteFiles = append(overwriteFiles, entry)
		}
	}

	return overwriteFiles
}

func (snap Snapshot) copyArchivedFiles(meta *domain.Metadata, archiveDir string) error {
	resolvers := snap.buildResolvers()

	if len(meta.Files) == 0 {
		return fmt.Errorf("ファイル情報が空です。データが残っている場合、metadata.yamlが破損している可能性があります。")
	}

	for _, entry := range meta.Files {
		src := filepath.Join(archiveDir, entry.ArchivePath)
		resolver, ok := resolvers[entry.BaseType]
		if !ok {
			return fmt.Errorf("サポート外のBaseType(%s)が渡されました: ", entry.BaseType)
		}
		dst, err := resolver(entry.OriginalPath)
		if err != nil {
			return fmt.Errorf("パス解決に失敗しました: %w", err)
		}

		if err := snap.fs.CopyFileOrDir(src, dst, true); err != nil {
			return fmt.Errorf("ファイル/ディレクトリのコピーに失敗しました: %w", err)
		}
	}

	return nil
}
