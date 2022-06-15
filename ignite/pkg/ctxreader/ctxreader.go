// Package ctxreader 將語境帶到 io.讀者
package ctxreader

import (
	"context"
	"io"
	"sync"
)

type cancelableReader struct {
	io.Reader
	ctx context.Context
	m   sync.Mutex
	err error
}

// New 返回一個新的閱讀器，該閱讀器通過其 r.Read() 方法發出上下文錯誤
// 當 ctx 取消時。
func New(ctx context.Context, r io.Reader) io.Reader {
	return &cancelableReader{Reader: r, ctx: ctx}
}

// Read 實現 io.Reader 並在讀取完成時停止阻塞
// 或上下文被取消。
func (r *cancelableReader) Read(data []byte) (n int, err error) {
	r.m.Lock()
	defer r.m.Unlock()

	if r.err != nil {
		return 0, r.err
	}

	var (
		readerN   int
		readerErr error
	)
	isRead := make(chan struct{})
	go func() {
		readerN, readerErr = r.Reader.Read(data)
		close(isRead)
	}()

	select {
	case <-r.ctx.Done():
		r.err = r.ctx.Err()
		return 0, r.ctx.Err()
	case <-isRead:
		r.err = readerErr
		return readerN, readerErr
	}
}
