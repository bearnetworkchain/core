package looseerrgroup

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// 等待直到 g.Wait() 返回或 ctx 取消，以先發生者為準。
// 返回的錯誤是 context.Canceled 如果 ctx 取消否則 g.Wait() 返回的錯誤。
//
// 這在 errgroup 不能與 errgroup.WithContext 一起使用時很有用，如果執行會發生這種情況
// func 不支持取消。
func Wait(ctx context.Context, g *errgroup.Group) error {
	doneC := make(chan struct{})

	go func() { g.Wait(); close(doneC) }()

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-doneC:
		return g.Wait()
	}
}
