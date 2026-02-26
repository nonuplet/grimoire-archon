package domain

import (
	"fmt"
	"time"
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

// SnapshotCondition スナップショットのチェック時の状態を表す
type SnapshotCondition int

const (
	// SnapshotDirNotFound スナップショット用ディレクトリがない
	SnapshotDirNotFound SnapshotCondition = iota
	// SnapshotFileNotFound スナップショットファイルがない
	SnapshotFileNotFound
	// SnapshotFileTooOld スナップショットファイルが古い
	SnapshotFileTooOld
	// SnapshotHealthy スナップショットが正常
	SnapshotHealthy
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

// FileEntry はzipに含まれる1ファイル分のメタデータです。
type FileEntry struct {
	ModifiedAt   time.Time `yaml:"modified_at"`
	ArchivePath  string    `yaml:"archive_path"`
	BaseType     BaseType  `yaml:"type"`
	OriginalPath string    `yaml:"original_path"`
}

// WinDirectoryType はWindows関連ディレクトリ(AppData, Document)の種類を表す型
type WinDirectoryType string

const (
	// DirectoryLocal AppData/Local
	DirectoryLocal WinDirectoryType = "Local"
	// DirectoryLocalLow AppData/LocalLow
	DirectoryLocalLow WinDirectoryType = "LocalLow"
	// DirectoryRoaming AppData/Roaming
	DirectoryRoaming WinDirectoryType = "Roaming"
	// DirectoryDocuments Documents
	DirectoryDocuments WinDirectoryType = "Documents"
)

// GetWinDirType Windows関連ディレクトリタイプが有効かどうかを検証
func (d WinDirectoryType) GetWinDirType() error {
	switch d {
	case DirectoryLocal, DirectoryLocalLow, DirectoryRoaming, DirectoryDocuments:
		return nil
	default:
		return fmt.Errorf("サポートされていないディレクトリタイプ %s が指定されました", d)
	}
}
