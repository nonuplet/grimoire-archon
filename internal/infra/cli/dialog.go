package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// AskYesNo ユーザーにYes/Noの選択を促す
func AskYesNo(r io.Reader, question string, defaultYes bool) (bool, error) {
	reader := bufio.NewReader(r)
	if defaultYes {
		fmt.Printf("%s [Y/n]: ", question)
	} else {
		fmt.Printf("%s [y/N]: ", question)
	}

	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("askでエラーが発生しました: %w", err)
	}
	input = strings.TrimSpace(strings.ToLower(input))

	if defaultYes {
		return input == "" || input == "y" || input == "yes", nil
	}
	return input != "" && input != `n` && input != "no", nil
}
