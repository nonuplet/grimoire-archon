package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/nonuplet/grimoire-archon/internal/domain"
	"github.com/nonuplet/grimoire-archon/internal/infra/cli"
	"github.com/nonuplet/grimoire-archon/internal/infra/filesystem"
)

// RestoreFromTmp は archiveDir で指定されたディレクトリ内のバックアップデータをリストアします。
func RestoreFromTmp(archonCfg *domain.ArchonConfig, gameCfg *domain.GameConfig, archiveDir string) error {
	// metadata.yamlのロード
	metaYaml := filepath.Join(archiveDir, "metadata.yaml")
	meta, err := domain.LoadMetaData(metaYaml)
	if err != nil {
		return fmt.Errorf("metadata.yamlのロードに失敗しました: %w", err)
	}

	// OSの違いをチェック
	if osErr := checkDifferentOs(meta.Os); osErr != nil {
		return osErr
	}

	// バックアップリストにないファイルをリストアップ
	notDefined := getNotDefinedFiles(meta, gameCfg)
	if len(notDefined) > 0 {
		fmt.Printf("コンフィグに設定した BackupTargets 以外のファイルが見つかりました。\n\n")
		for _, file := range notDefined {
			fmt.Printf("- %s: %s\n", file.BaseType, file.OriginalPath)
		}
		ok, err := cli.AskYesNo(os.Stdin, "\nリストアを続行してもよろしいですか？", true)
		if err != nil {
			return fmt.Errorf("ユーザーの回答取得に失敗しました: %w", err)
		}
		if !ok {
			return fmt.Errorf("リストアを中止しました")
		}
	}

	// 上書きがあるかチェック
	overwriteFiles := getOverwriteFiles(archonCfg, gameCfg, meta)
	if len(overwriteFiles) > 0 {
		fmt.Printf("以下のファイルは上書きされます。\n\n")
		for _, file := range overwriteFiles {
			fmt.Printf("- %s: %s\n", file.BaseType, file.OriginalPath)
		}
		ok, err := cli.AskYesNo(os.Stdin, "\n対象のファイルを上書きしてもよろしいですか？", true)
		if err != nil {
			return fmt.Errorf("ユーザーの回答取得に失敗しました: %w", err)
		}
		if !ok {
			return fmt.Errorf("リストアを中止しました")
		}
	}

	fmt.Println("バックアップで復元しています...")
	if err := copyArchivedFiles(archonCfg, gameCfg, meta, archiveDir); err != nil {
		return fmt.Errorf("復元に失敗しました: %w", err)
	}

	return nil
}

func checkDifferentOs(archivedOs string) error {
	currentOs := runtime.GOOS
	if currentOs != archivedOs {
		q := fmt.Sprintf("バックアップ元のOS(%s)と現在のOS(%s)が異なります。実行しますか?", archivedOs, currentOs)
		ok, err := cli.AskYesNo(os.Stdin, q, false)
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
// archivedEntries のうち gameCfg.BackupTargets に定義されていないエントリを返します。
func getNotDefinedFiles(meta *domain.Metadata, gameCfg *domain.GameConfig) []domain.FileEntry {
	archivedEntries := meta.Files
	targets := gameCfg.BackupTargets

	// BaseType ごとに対応するターゲットリストへマッピング
	targetMap := map[domain.BaseType][]string{
		domain.BaseTypeInstallDir:      getTargetList(targets, domain.BaseTypeInstallDir),
		domain.BaseTypeUserHome:        getTargetList(targets, domain.BaseTypeUserHome),
		domain.BaseTypeAppdataLocal:    getTargetList(targets, domain.BaseTypeAppdataLocal),
		domain.BaseTypeAppdataLocalLow: getTargetList(targets, domain.BaseTypeAppdataLocalLow),
		domain.BaseTypeAppdataRoaming:  getTargetList(targets, domain.BaseTypeAppdataRoaming),
		domain.BaseTypeWinDocuments:    getTargetList(targets, domain.BaseTypeWinDocuments),
		domain.BaseTypeAbsolute:        getTargetList(targets, domain.BaseTypeAbsolute),
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
func getTargetList(targets *domain.BackupTargetConfig, baseType domain.BaseType) []string {
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
func getOverwriteFiles(archonCfg *domain.ArchonConfig, gameCfg *domain.GameConfig, meta *domain.Metadata) []domain.FileEntry {
	var overwriteFiles []domain.FileEntry
	if meta == nil || len(meta.Files) == 0 {
		return overwriteFiles
	}

	// 各 BaseType のソースディレクトリ解決関数を定義
	resolveInstallDir := func(rel string) (string, error) { // nolint:unparam // resolve関数をmapに入れるため、型を合わせる必要がある
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

	resolvers := map[domain.BaseType]func(string) (string, error){
		domain.BaseTypeInstallDir:      resolveInstallDir,
		domain.BaseTypeUserHome:        resolveUserHome,
		domain.BaseTypeAppdataLocal:    resolveWinDir("Local"),
		domain.BaseTypeAppdataLocalLow: resolveWinDir("LocalLow"),
		domain.BaseTypeAppdataRoaming:  resolveWinDir("Roaming"),
		domain.BaseTypeWinDocuments:    resolveWinDir("Documents"),
		domain.BaseTypeAbsolute:        resolveAbsolute,
	}

	for _, entry := range meta.Files {
		resolver, ok := resolvers[entry.BaseType]
		if !ok {
			continue
		}
		dst, err := resolver(entry.OriginalPath)
		if err != nil {
			continue
		}

		if _, err := filesystem.GetInfo(dst); err == nil {
			overwriteFiles = append(overwriteFiles, entry)
		} else if !os.IsNotExist(err) {
			// 参照権限等の理由で確認できない場合も、上書き対象として扱う
			overwriteFiles = append(overwriteFiles, entry)
		}
	}

	return overwriteFiles
}

func copyArchivedFiles(archonCfg *domain.ArchonConfig, gameCfg *domain.GameConfig, meta *domain.Metadata, archiveDir string) error {
	// 各 BaseType のソースディレクトリ解決関数を定義
	resolveInstallDir := func(rel string) (string, error) { // nolint:unparam // resolve関数をmapに入れるため、型を合わせる必要がある
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

	resolvers := map[domain.BaseType]func(string) (string, error){
		domain.BaseTypeInstallDir:      resolveInstallDir,
		domain.BaseTypeUserHome:        resolveUserHome,
		domain.BaseTypeAppdataLocal:    resolveWinDir("Local"),
		domain.BaseTypeAppdataLocalLow: resolveWinDir("LocalLow"),
		domain.BaseTypeAppdataRoaming:  resolveWinDir("Roaming"),
		domain.BaseTypeWinDocuments:    resolveWinDir("Documents"),
		domain.BaseTypeAbsolute:        resolveAbsolute,
	}

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

		if err := filesystem.CopyFileOrDir(src, dst, true); err != nil {
			return fmt.Errorf("ファイル/ディレクトリのコピーに失敗しました: %w", err)
		}
	}

	return nil
}
