package entrywriter

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/pkg/errors"

	"github.com/bearnetworkchain/core/ignite/pkg/xstrings"
)

const (
	None = "-"
)

var ErrInvalidFormat = errors.New("invalid entry format")

// MustWrite 如果條目格式無效，則寫入列表條目並恐慌
func MustWrite(out io.Writer, header []string, entries ...[]string) error {
	err := Write(out, header, entries...)
	if errors.Is(err, ErrInvalidFormat) {
		panic(err)
	}
	return err
}

// Write 寫入列表條目
func Write(out io.Writer, header []string, entries ...[]string) error {
	w := &tabwriter.Writer{}
	w.Init(out, 0, 8, 0, '\t', 0)

	formatLine := func(line []string, title bool) (formatted string) {
		for _, cell := range line {
			if title {
				cell = xstrings.Title(cell)
			}
			formatted += fmt.Sprintf("%s \t", cell)
		}
		return formatted
	}

	if len(header) == 0 {
		return errors.Wrap(ErrInvalidFormat, "空標題")
	}

	// 寫頭
	if _, err := fmt.Fprintln(w, formatLine(header, true)); err != nil {
		return err
	}

	// 寫條目
	for i, entry := range entries {
		if len(entry) != len(header) {
			return errors.Wrapf(ErrInvalidFormat, "入口 %d 與標頭長度不匹配", i)
		}
		if _, err := fmt.Fprintf(w, formatLine(entry, false)+"\n"); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return w.Flush()
}
