package swaggercombine

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"regexp"

	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/exec"
	"github.com/ignite-hq/cli/ignite/pkg/nodetime"
)

//Config 代表 swagger-combine 配置。
type Config struct {
	Swagger string `json:"swagger"`
	Info    Info   `json:"info"`
	APIs    []API  `json:"apis"`
}

type Info struct {
	Title       string `json:"title"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type API struct {
	ID           string       `json:"-"`
	URL          string       `json:"url"`
	OperationIDs OperationIDs `json:"operationIds"`
	Dereference  struct {
		Circular string `json:"circular"`
	} `json:"dereference"`
}

type OperationIDs struct {
	Rename map[string]string `json:"rename"`
}

var opReg = regexp.MustCompile(`(?m)operationId.+?(\w+)`)

// AddSpec 通過 fs 中的路徑和 spec 的唯一 id 向 Config 添加一個新的 OpenAPI 規範。
func (c *Config) AddSpec(id, path string) error {
	// 使 operationId 字段唯一。
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	ops := opReg.FindAllStringSubmatch(string(content), -1)
	rename := make(map[string]string, len(ops))

	for _, op := range ops {
		o := op[1]
		rename[o] = id + o
	}

	// 添加帶有替換操作 ID 的 api。
	c.APIs = append(c.APIs, API{
		ID:           id,
		URL:          path,
		OperationIDs: OperationIDs{Rename: rename},
		// 由於 https://github.com/maxdome/swagger-combine/pull/110 在#835 中啟用更多服務後添加
		Dereference: struct {
			Circular string `json:"circular"`
		}(struct{ Circular string }{Circular: "ignore"}),
	})

	return nil
}

// 將 openapi 規範合併為一個並保存到路徑中。
// specs 是一個規範 id-fs 路徑對。
func Combine(ctx context.Context, c Config, out string) error {
	command, cleanup, err := nodetime.Command(nodetime.CommandSwaggerCombine)
	if err != nil {
		return err
	}
	defer cleanup()

	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	if err := json.NewEncoder(f).Encode(c); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	command = append(command, []string{
		f.Name(),
		"-o", out,
		"-f", "yaml",
		"--continueOnConflictingPaths", "true",
		"--includeDefinitions", "true",
	}...)

	// 執行命令。
	return exec.Exec(ctx, command, exec.IncludeStdLogsToError())
}
