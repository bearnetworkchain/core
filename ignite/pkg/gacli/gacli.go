// Package gacli 是 Google Analytics 的客戶端，用於發送提示類型 = 事件的數據點。
package gacli

import (
	"net/http"
	"net/url"
)

const (
	endpoint = "https://www.google-analytics.com/collect"
)

// 客戶端是分析客戶端。
type Client struct {
	gaid string
}

// New 使用 Segment 為 Segment.io 創建一個新的分析客戶端
// 端點和訪問密鑰。
func New(gaid string) *Client {
	return &Client{
		gaid: gaid,
	}
}

// Metric 表示一個數據點。
type Metric struct {
	Category string
	Action   string
	Label    string
	Value    string
	User     string
	Version  string
}

// Send 將指標發送到 GA。
func (c *Client) Send(metric Metric) error {
	v := url.Values{
		"v":   {"1"},
		"tid": {c.gaid},
		"cid": {metric.User},
		"t":   {"event"},
		"ec":  {metric.Category},
		"ea":  {metric.Action},
		"ua":  {"Opera/9.80 (Windows NT 6.0) Presto/2.12.388 Version/12.14"},
	}
	if metric.Label != "" {
		v.Set("el", metric.Label)
	}
	if metric.Value != "" {
		v.Set("ev", metric.Value)
	}
	if metric.Version != "" {
		v.Set("an", metric.Version)
		v.Set("av", metric.Version)
	}
	resp, err := http.PostForm(endpoint, v)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
