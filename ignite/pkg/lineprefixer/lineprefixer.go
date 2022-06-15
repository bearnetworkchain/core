// Package lineprefixer 是為新行添加前綴的助手。
package lineprefixer

import (
	"bytes"
	"io"
)

// Writer 是一個帶前綴的行 writer。
type Writer struct {
	prefix       func() string
	w            io.Writer
	shouldPrefix bool
}

// NewWriter 返回一個為每一行添加前綴的新作家
// 書面。然後它將前綴數據流寫入 w。
func NewWriter(w io.Writer, prefix func() string) *Writer {
	return &Writer{
		prefix:       prefix,
		w:            w,
		shouldPrefix: true,
	}
}

// Write 實現 io.Writer.
func (p *Writer) Write(b []byte) (n int, err error) {
	var (
		blen         = len(b)
		lastChar     = b[blen-1]
		newLine      = byte('\n')
		snewLine     = []byte{newLine}
		replaceCount = bytes.Count(b, snewLine)
		prefix       = []byte(p.prefix())
	)
	if lastChar == newLine {
		replaceCount--
	}
	b = bytes.Replace(b, snewLine, append(snewLine, prefix...), replaceCount)
	if p.shouldPrefix {
		b = append(prefix, b...)
	}
	p.shouldPrefix = lastChar == newLine
	if _, err := p.w.Write(b); err != nil {
		return 0, err
	}
	return blen, nil
}
