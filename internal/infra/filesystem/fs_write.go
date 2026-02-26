package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// MkdirAll は指定されたパスにディレクトリを作成します。
func (f *FileSystem) MkdirAll(path string, perm os.FileMode) error {
	path, err := f.getAbsolutePath(path)
	if err != nil {
		return fmt.Errorf("ディレクトリ (%s) のパス取得に失敗しました: %w", path, err)
	}

	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("ディレクトリ (%s) の作成に失敗しました: %w", path, err)
	}

	return nil
}

// WriteFile は指定されたパスにファイルを作成し、データを書き込みます。
func (f *FileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	path, err := f.getAbsolutePath(path)
	if err != nil {
		return fmt.Errorf("ファイル (%s) のパス取得に失敗しました: %w", path, err)
	}

	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("ファイル (%s) の書き込みに失敗しました: %w", path, err)
	}

	return nil
}

// CopyFileOrDir ファイルまたはディレクトリをコピー。
// overwrite が true の場合、コピー先が既に存在しても上書きする。
// overwrite が false の場合、コピー先が既に存在するとエラーを返す。
func (f *FileSystem) CopyFileOrDir(src, dst string, overwrite bool) error {
	// 絶対パスに変換
	src, err := f.getAbsolutePath(src)
	if err != nil {
		return fmt.Errorf("コピー元パスの取得: %w", err)
	}
	dst, err = f.getAbsolutePath(dst)
	if err != nil {
		return fmt.Errorf("コピー先パスの取得: %w", err)
	}

	info, err := f.Stat(src)
	if err != nil {
		return fmt.Errorf("コピー元 %s が取得できません", src)
	}

	if info.IsDir() {
		// Directory
		fmt.Printf("ディレクトリをコピー中: %s -> %s\n", src, dst)
		if err := f.copyDir(src, dst, overwrite); err != nil {
			return fmt.Errorf("ディレクトリのコピーに失敗しました: %w", err)
		}
	} else {
		// File
		fmt.Printf("ファイルをコピー中: %s -> %s\n", src, dst)
		if err := f.copyFile(src, dst, overwrite); err != nil {
			return fmt.Errorf("ファイルのコピーに失敗しました: %w", err)
		}
	}
	return nil
}

// copyDir はディレクトリ src を dst へ再帰的にコピーします。
// overwrite が false のとき、既存ファイルはスキップせずエラーを返します。
func (f *FileSystem) copyDir(src, dst string, overwrite bool) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("ディレクトリの読み込みに失敗しました (%s): %w", src, err)
	}

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました (%s): %w", dst, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := f.copyDir(srcPath, dstPath, overwrite); err != nil {
				return err
			}
		} else {
			if err := f.copyFile(srcPath, dstPath, overwrite); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile はファイル src を dst へコピーします。
// overwrite が false のとき、dst が既に存在する場合はエラーを返します。
func (f *FileSystem) copyFile(src, dst string, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("コピー先がすでに存在します (%s)", dst)
		}
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("コピー元が取得できません (%s): %w", src, err)
	}
	defer func(in *os.File) {
		inErr := in.Close()
		if inErr != nil {
			fmt.Fprintf(os.Stderr, "コピー元のクローズに失敗しました %v", inErr)
		}
	}(in)

	if mkdirErr := os.MkdirAll(filepath.Dir(dst), 0o755); mkdirErr != nil {
		return fmt.Errorf("親ディレクトリの作成に失敗しました (%s): %w", filepath.Dir(dst), mkdirErr)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました (%s): %w", dst, err)
	}
	defer func(out *os.File) {
		closeErr := out.Close()
		if closeErr != nil {
			fmt.Fprintf(os.Stderr, "コピー先のクローズに失敗しました %v", closeErr)
		}
	}(out)

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("ファイルのコピーに失敗しました (%s -> %s): %w", src, dst, err)
	}
	return nil
}
