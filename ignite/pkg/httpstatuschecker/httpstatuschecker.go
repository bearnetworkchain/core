// httpstatuschecker 是一個檢查http頁面健康狀況的工具。
package httpstatuschecker

import (
	"context"
	"net/http"
)

type checker struct {
	c      *http.Client
	addr   string
	method string
}

// Option用於自定義檢查器。
type Option func(*checker)

// Method配置http方法。
func Method(name string) Option {
	return func(cr *checker) {
		cr.method = name
	}
}

// Check通過應用選項檢查給定的 http 地址是否存在。
func Check(ctx context.Context, addr string, options ...Option) (isAvailable bool, err error) {
	cr := &checker{
		c:      http.DefaultClient,
		addr:   addr,
		method: http.MethodGet,
	}
	for _, o := range options {
		o(cr)
	}
	return cr.check(ctx)
}

func (c *checker) check(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, c.method, c.addr, nil)
	if err != nil {
		return false, err
	}
	res, err := c.c.Do(req)
	if err != nil {
		return false, nil
	}
	defer res.Body.Close()
	isOKStatus := res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices
	return isOKStatus, nil
}
