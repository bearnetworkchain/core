package xhttp

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// ShutdownTimeout是等待所有請求完成的超時時間。
const ShutdownTimeout = time.Minute

// Serve啟動 s 服務器並在 ctx 取消後將其關閉。
func Serve(ctx context.Context, s *http.Server) error {
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
		defer cancel()

		s.Shutdown(shutdownCtx)
	}()

	err := s.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}
