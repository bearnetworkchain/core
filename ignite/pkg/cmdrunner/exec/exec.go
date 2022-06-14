// Package exec provides easy access to command execution for basic uses.
package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
)

// ExitError 是 exec.ExitError 的別名
type ExitError = exec.ExitError

type execConfig struct {
	stepOptions           []step.Option
	includeStdLogsToError bool
}

type Option func(*execConfig)

func StepOption(o step.Option) Option {
	return func(c *execConfig) {
		c.stepOptions = append(c.stepOptions, o)
	}
}

func IncludeStdLogsToError() Option {
	return func(c *execConfig) {
		c.includeStdLogsToError = true
	}
}

// Exec executes 帶有 args 的命令，它是基本命令執行的快捷方式。
func Exec(ctx context.Context, fullCommand []string, options ...Option) error {
	errb := &bytes.Buffer{}
	logs := &bytes.Buffer{}

	c := &execConfig{
		stepOptions: []step.Option{
			step.Exec(fullCommand[0], fullCommand[1:]...),
			step.Stdout(logs),
			step.Stderr(errb),
		},
	}

	for _, apply := range options {
		apply(c)
	}

	err := cmdrunner.New().Run(ctx, step.New(c.stepOptions...))
	if err != nil {
		return &Error{
			Err:                   errors.Wrap(err, errb.String()),
			Command:               strings.Join(fullCommand, " "),
			StdLogs:               logs.String(),
			includeStdLogsToError: c.includeStdLogsToError,
		}
	}

	return nil
}

// Error 提供來自已執行程序的詳細錯誤。
type Error struct {
	Err                   error
	Command               string
	StdLogs               string // 從代碼生成工具收集的日誌。
	includeStdLogsToError bool
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	message := fmt.Sprintf("運行命令時出錯 %s: %s", e.Command, e.Err.Error())
	if e.includeStdLogsToError && strings.TrimSpace(e.StdLogs) != "" {
		return fmt.Sprintf("%s\n\n%s", message, e.StdLogs)
	}
	return message
}
