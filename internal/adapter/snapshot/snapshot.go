package snapshot

import (
	"io"
	"os"

	"github.com/nonuplet/grimoire-archon/internal/domain"
)

// Snapshot snapshot関連のアダプター
type Snapshot struct {
	archonCfg *domain.ArchonConfig
	gameCfg   *domain.GameConfig
	fs        FileSystem
	cli       Cli
}

// FileSystem ファイルシステム操作のインターフェース
type FileSystem interface {
	Stat(path string) (os.FileInfo, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	CopyFileOrDir(src, dst string, overwrite bool) error
}

// Cli cli操作のインターフェース
type Cli interface {
	AskYesNo(r io.Reader, question string, defaultYes bool) (bool, error)
}

// NewSnapshot snapshotアダプターの生成
func NewSnapshot(archonCfg *domain.ArchonConfig, gameCfg *domain.GameConfig, fs FileSystem, cli Cli) *Snapshot {
	return &Snapshot{
		archonCfg: archonCfg,
		gameCfg:   gameCfg,
		fs:        fs,
		cli:       cli,
	}
}
