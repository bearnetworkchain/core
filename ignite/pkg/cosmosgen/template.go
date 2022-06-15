package cosmosgen

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/takuoki/gocase"
)

var (
	//go:嵌入模板/*
	templates embed.FS

	templateJSClient  = newTemplateWriter("js")         // js 包裝客戶端.
	templateVuexRoot  = newTemplateWriter("vuex/root")  // vuex 存儲裝載機.
	templateVuexStore = newTemplateWriter("vuex/store") // vuex store.

)

type templateWriter struct {
	templateDir string
}

// tpl 為模板返回一個 func 駐留在 templatePath 以初始化文本模板
// 使用給定的 protoPath。
func newTemplateWriter(templateDir string) templateWriter {
	return templateWriter{
		templateDir,
	}
}

func (t templateWriter) Write(destDir, protoPath string, data interface{}) error {
	base := filepath.Join("templates", t.templateDir)

	// find out templates inside the dir.
	files, err := templates.ReadDir(base)
	if err != nil {
		return err
	}

	var paths []string
	for _, file := range files {
		paths = append(paths, filepath.Join(base, file.Name()))
	}

	funcs := template.FuncMap{
		"camelCase": strcase.ToLowerCamel,
		"camelCaseSta": func(word string) string {
			return gocase.Revert(strcase.ToLowerCamel(word))
		},
		"resolveFile": func(fullPath string) string {
			rel, _ := filepath.Rel(protoPath, fullPath)
			rel = strings.TrimSuffix(rel, ".proto")
			return rel
		},
		"inc": func(i int) int {
			return i + 1
		},
		"replace": strings.ReplaceAll,
	}

	// 渲染和編寫模板。
	write := func(path string) error {
		tpl := template.
			Must(
				template.
					New(filepath.Base(path)).
					Funcs(funcs).
					ParseFS(templates, paths...),
			)

		out := filepath.Join(destDir, strings.TrimSuffix(filepath.Base(path), ".tpl"))

		f, err := os.OpenFile(out, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0766)
		if err != nil {
			return err
		}
		defer f.Close()

		return tpl.Execute(f, data)
	}

	for _, path := range paths {
		if err := write(path); err != nil {
			return err
		}
	}

	return nil
}
