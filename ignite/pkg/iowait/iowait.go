package iowait

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// 直到等待 s 在字符串中出現 n 次並且
// 然後停止阻塞。
func Until(r io.Reader, s string, n int) (capturedLines []string, err error) {
	total := n
	scanner := bufio.NewScanner(r)
	for {
		if n == 0 {
			return capturedLines, nil
		}
		if !scanner.Scan() {
			if n != 0 {
				return capturedLines, fmt.Errorf("找不到 %d out of %d", n, total)
			}
			return capturedLines, scanner.Err()
		}
		if strings.Contains(scanner.Text(), s) {
			capturedLines = append(capturedLines, scanner.Text())
			n--
		}
	}
}
