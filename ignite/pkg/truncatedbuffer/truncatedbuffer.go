package truncatedbuffer

import (
	"bytes"
)

// TruncatedBuffer 包含一個容量有限的字節緩衝區
// 如果長度達到最大容量，則緩衝區在寫入時被截斷
// 只保留第一個字節
type TruncatedBuffer struct {
	buf *bytes.Buffer
	cap int
}

// NewTruncatedBuffer 返回一個新的TruncatedBuffer
// 如果提供的上限為 0，則截斷緩衝區沒有截斷限制
func NewTruncatedBuffer(cap int) *TruncatedBuffer {
	return &TruncatedBuffer{
		buf: &bytes.Buffer{},
		cap: cap,
	}
}

//GetCap 返回緩衝區
func (b TruncatedBuffer) GetBuffer() *bytes.Buffer {
	return b.buf
}

// GetCap 返回緩衝區的最大容量
func (b TruncatedBuffer) GetCap() int {
	return b.cap
}

//Write 實現 io.Writer
func (b *TruncatedBuffer) Write(p []byte) (n int, err error) {
	n, err = b.buf.Write(p)
	if err != nil {
		return n, err
	}

	// 檢查剩餘字節
	surplus := b.buf.Len() - b.cap

	if b.cap > 0 && surplus > 0 {
		b.buf.Truncate(b.cap)
	}

	return n, nil
}
