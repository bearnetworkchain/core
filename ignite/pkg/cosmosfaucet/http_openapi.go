package cosmosfaucet

import (
	_ "embed" // 用於嵌入 openapi 資產。
	"html/template"
	"net/http"
)

const (
	fileNameOpenAPISpec = "openapi/openapi.yml.tmpl"
)

//go:嵌入openapi/openapi.yml.tmpl
var bytesOpenAPISpec []byte

var tmplOpenAPISpec = template.Must(template.New(fileNameOpenAPISpec).Parse(string(bytesOpenAPISpec)))

type openAPIData struct {
	ChainID    string
	APIAddress string
}

func (f Faucet) openAPISpecHandler(w http.ResponseWriter, r *http.Request) {
	tmplOpenAPISpec.Execute(w, f.openAPIData)
}
