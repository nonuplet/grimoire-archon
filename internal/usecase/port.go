package usecase

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/nonuplet/grimoire-archon/internal/domain"
)

// -- adapter ---

// Snapshot はスナップショット操作のインターフェース
type Snapshot interface {
	CopyToTmp(tmpDir string) ([]domain.FileEntry, error)
	SaveMetaData(path string, meta *domain.Metadata) error
	CheckAndCreateSnapshotDir() error
	RestoreFromTmp(archiveDir string) error
}

// --- infra ---

// FileSystem はファイル操作のインターフェース
type FileSystem interface {
	// Read
	Stat(path string) (os.FileInfo, error)
	ReadFile(path string) ([]byte, error)
	ReadDir(path string) ([]os.DirEntry, error)
	GetTimestamp() string

	// Write
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	CopyFileOrDir(src, dst string, overwrite bool) error

	// Remove
	ClearDirectoryContents(path string) error
	RemoveAll(path string) error

	// Archive
	IsZipFile(path string) bool
	Zip(srcDir, destZip string) error
	Unzip(src, dest string) error
}

// SteamCmd はsteamcmd操作のインターフェース
type SteamCmd interface {
	Check() error
	Update(ctx context.Context, appID, installDir, platform string) error
}

// Cli はコマンドライン入出力のインターフェース
type Cli interface {
	AskYesNo(r io.Reader, question string, defaultYes bool) (bool, error)
	Writeln(builder *strings.Builder, strs ...string)
}
