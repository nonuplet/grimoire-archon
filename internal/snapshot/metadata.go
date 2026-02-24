package snapshot

import (
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

// MetaVersion はメタデータスキーマのバージョンです。
// フィールドの追加・変更が生じた際にインクリメントします。
const MetaVersion = "1"

// BaseType はファイルのリストア起点となるディレクトリの種別です。
type BaseType string

const (
	// BaseTypeInstallDir はゲームのインストールディレクトリ配下のファイル, ディレクトリです。
	BaseTypeInstallDir BaseType = "install_dir"

	// BaseTypeUserHome はユーザーのホームディレクトリ配下のファイル, ディレクトリです。
	BaseTypeUserHome BaseType = "user_home"

	// BaseTypeAppdataLocal はWindowsの AppData/Local に対応するファイル, ディレクトリです。
	BaseTypeAppdataLocal BaseType = "win_local"
	// BaseTypeAppdataLocalLow はWindowsの AppData/LocalLow に対応するファイル, ディレクトリです。Wine/Proton内ではその環境に合わせて展開します。
	BaseTypeAppdataLocalLow BaseType = "win_locallow" // %LocalAppData%/../ = LocalLow
	// BaseTypeAppdataRoaming はWindowsの AppData/Roaming に対応するファイル, ディレクトリです。Wine/Proton内ではその環境に合わせて展開します。
	BaseTypeAppdataRoaming BaseType = "win_roaming" // %AppData% = Roaming
	// BaseTypeWinDocuments はWindowsの %Documents% に対応するファイル, ディレクトリです。Wine/Proton内ではその環境に合わせて展開します。
	BaseTypeWinDocuments BaseType = "win_documents" // %Documents%

	// BaseTypeAbsolute は絶対パスを示します。
	BaseTypeAbsolute BaseType = "absolute"
)

// Metadata はスナップショットzip内の matadata.yaml に書き出す構造体です。
type Metadata struct {
	Version     string      `yaml:"version"` // MetaVersion
	Name        string      `yaml:"name"`
	CreatedAt   time.Time   `yaml:"created_at"`
	ToolVersion string      `yaml:"tool_version"`
	Os          string      `yaml:"os"`
	Files       []FileEntry `yaml:"files"`
}

// Save メタデータの保存
func (m *Metadata) Save(path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("metadata.yamlのマーシャリングに失敗しました: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("metadata.yamlの書き込みに失敗しました: %w", err)
	}

	return nil
}

// FileEntry はzipに含まれる1ファイル分のメタデータです。
type FileEntry struct {
	ModifiedAt   time.Time `yaml:"modified_at"`
	ArchivePath  string    `yaml:"archive_path"`
	BaseType     BaseType  `yaml:"type"`
	OriginalPath string    `yaml:"original_path"`
}
