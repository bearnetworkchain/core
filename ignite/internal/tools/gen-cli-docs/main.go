// 此工具生成 Ignite CLI 文檔以放置在 docs/cli 目錄中並部署
// on docs.ignite.com
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	ignitecmd "github.com/ignite-hq/cli/ignite/cmd"
)

const head = `---
側邊欄位置: 7
描述: Ignite CLI docs.
---

# CLI 參考

Ignite CLI 文檔.
`

func main() {
	outPath := flag.String("out", ".", ".md file path to place Ignite CLI docs inside")
	flag.Parse()

	if err := generate(ignitecmd.New(), *outPath); err != nil {
		log.Fatal(err)
	}
}

func generate(cmd *cobra.Command, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(outPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := fmt.Fprint(f, head); err != nil {
		return err
	}

	return generateCmd(cmd, f)
}

func generateCmd(cmd *cobra.Command, w io.Writer) error {
	cmd.DisableAutoGenTag = true

	b := &bytes.Buffer{}
	if err := doc.GenMarkdownCustom(cmd, b, linkHandler); err != nil {
		return err
	}

// 這裡我們將子標題改為粗體樣式。否則，這些標題將獲得
// 打印在 docs.starport.network 的右側菜單中，這令人不快，因為
// 我們只想看到所有可用命令的列表，沒有額外的噪音。
	sc := bufio.NewScanner(b)
	for sc.Scan() {
		t := sc.Text()
		if strings.HasPrefix(t, "###") {
			t = strings.TrimPrefix(t, "### ")
			t = fmt.Sprintf("**%s**", t)
		}
		if _, err := fmt.Fprintln(w, t); err != nil {
			return err
		}
	}

	for _, cmd := range cmd.Commands() {
		if !cmd.IsAvailableCommand() || cmd.IsAdditionalHelpTopicCommand() {
			continue
		}

		io.WriteString(w, "\n")

		if err := generateCmd(cmd, w); err != nil {
			return err
		}
	}

	return nil
}

func linkHandler(link string) string {
	link = strings.ReplaceAll(link, "_", "-")
	link = strings.TrimSuffix(link, ".md")
	link = "#" + link
	return link
}
