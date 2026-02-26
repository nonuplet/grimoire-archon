package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetAbsolutePath はパスを絶対パスに変換します。
// パス先頭の "~" はユーザーのホームディレクトリに展開されます。
// 環境変数は展開されます。
func GetAbsolutePath(path string) (string, error) {
	// 変数の展開
	path = os.ExpandEnv(path)

	// ~ の展開
	path, err := expandTilde(path)
	if err != nil {
		return "", fmt.Errorf("~の展開: %w", err)
	}

	// 絶対パスへの変換
	path, err = filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("絶対パスの取得: %w", err)
	}

	return path, nil
}

// expandTilde はパスの先頭にある "~" をユーザーのホームディレクトリに展開します。
func expandTilde(path string) (string, error) {
	// "~" で始まっていない、または "~" 単体でない場合はそのまま返す
	if path == "" || !strings.HasPrefix(path, "~") {
		return path, nil
	}

	// homeの取得
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリの取得に失敗しました: %w", err)
	}

	// "~" 単体の場合
	if path == "~" {
		return home, nil
	}

	// ~user/foobar 形式は弾く
	hasSeparator := strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\")
	if !hasSeparator {
		return "", errors.New("(~user) の形式には対応していません")
	}

	// path[2:] で "~/foo" の "foo" 部分だけを取り出す
	// filepath.Join が OS 固有のセパレータを適切に挿入してくれる
	return filepath.Join(home, path[2:]), nil
}
