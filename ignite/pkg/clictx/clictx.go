package clictx

import (
	"context"
	"os"
	"os/signal"
)

// From 從 ctx 創建一個新的上下文，當接收到退出信號時該上下文被取消。
func From(ctx context.Context) context.Context {
	var (
		ctxend, cancel = context.WithCancel(ctx)
		quit           = make(chan os.Signal, 1)
	)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		cancel()
	}()
	return ctxend
}
