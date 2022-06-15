package ctxticker

import (
	"context"
	"time"
)

// Do 每 d 調用一次 fn 直到 ctx 取消或 fn 返回一個非零錯誤。
func Do(ctx context.Context, d time.Duration, fn func() error) error {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			if err := fn(); err != nil {
				return err
			}
		}
	}
}

// DoNow 與 Do 相同，只是它在開始時對 fn 進行 +1 調用。
func DoNow(ctx context.Context, d time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		return err
	}
	return Do(ctx, d, fn)
}
