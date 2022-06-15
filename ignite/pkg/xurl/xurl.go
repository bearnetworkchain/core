package xurl

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

const (
	schemeTCP   = "tcp"
	schemeHTTP  = "http"
	schemeHTTPS = "https"
	schemeWS    = "ws"
)

// TCP 確保 URL 包含 TCP 方案。
func TCP(s string) (string, error) {
	u, err := parseURL(s)
	if err != nil {
		return "", err
	}

	u.Scheme = schemeTCP

	return u.String(), nil
}

// HTTP確保 URL 包含 HTTP 方案。
func HTTP(s string) (string, error) {
	u, err := parseURL(s)
	if err != nil {
		return "", err
	}

	u.Scheme = schemeHTTP

	return u.String(), nil
}

// HTTPS確保 URL 包含 HTTPS 方案。
func HTTPS(s string) (string, error) {
	u, err := parseURL(s)
	if err != nil {
		return "", err
	}

	u.Scheme = schemeHTTPS

	return u.String(), nil
}

// 噹噹前方案不是 HTTP 時，MightHTTPS 確保 URL 包含 HTTPS 方案。
// 當 URL 包含 HTTP 方案時，它不會被修改。
func MightHTTPS(s string) (string, error) {
	if strings.HasPrefix(strings.ToLower(s), "http://") {
		return s, nil
	}

	return HTTPS(s)
}

// WS 確保 URL 包含 WS 方案。
func WS(s string) (string, error) {
	u, err := parseURL(s)
	if err != nil {
		return "", err
	}

	u.Scheme = schemeWS

	return u.String(), nil
}

// HTTPEnsurePort 確保 url 具有適合連接類型的端口號。
func HTTPEnsurePort(s string) string {
	u, err := url.Parse(s)
	if err != nil || u.Port() != "" {
		return s
	}

	port := "80"

	if u.Scheme == schemeHTTPS {
		port = "443"
	}

	u.Host = fmt.Sprintf("%s:%s", u.Hostname(), port)

	return u.String()
}

//如果未指定，地址確保地址包含 localhost 作為主機。
func Address(address string) string {
	if strings.HasPrefix(address, ":") {
		return "localhost" + address
	}
	return address
}

func IsHTTP(address string) bool {
	return strings.HasPrefix(address, "http")
}

func parseURL(s string) (*url.URL, error) {
	if s == "" {
		return nil, errors.New("url is empty")
	}

// 處理 URI 是 IP:PORT 或 HOST:PORT 的情況
// 沒有方案前綴，因為這種情況不能被 URL 解析。
// 當 URI 沒有方案時，它被“url.Parse”解析為路徑
// 將冒號放在路徑中，這是無效的。
	if host, isAddrPort := addressPort(s); isAddrPort {
		return &url.URL{Host: host}, nil
	}

	p, err := url.Parse(Address(s))
	return p, err
}

func addressPort(s string) (string, bool) {
	// 檢查該值是否不包含 URI 路徑
	if strings.Index(s, "/") != -1 {
		return "", false
	}

	// 使用網絡拆分功能支持 IPv6 地址
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return "", false
	}
	if host == "" {
		host = "0.0.0.0"
	}
	return net.JoinHostPort(host, port), true
}
