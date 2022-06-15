package openapiconsole

import (
	"embed"
	"html/template"
	"net/http"
)

//GO：嵌入 index.tpl
var index embed.FS

// 處理程序返回一個 http 處理程序，該處理程序為 specURL 上的 OpenAPI 規範提供 OpenAPI 控制台服務。
func Handler(title, specURL string) http.HandlerFunc {
	t, _ := template.ParseFS(index, "index.tpl")

	return func(w http.ResponseWriter, req *http.Request) {
		t.Execute(w, struct {
			Title string
			URL   string
		}{
			title,
			specURL,
		})
	}
}
