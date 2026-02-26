package filesystem

// FileSystem ファイルシステム操作 (コピー/削除/圧縮/etc...)
type FileSystem struct{}

// NewFileSystem FileSystemのインスタンスを生成する
func NewFileSystem() *FileSystem {
	return &FileSystem{}
}
