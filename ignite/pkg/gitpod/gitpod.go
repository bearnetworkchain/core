package gitpod

import (
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/exec"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
)

// IsOnGitpod 報告是否在 Gitpod 上運行。
func IsOnGitpod() bool {
	return os.Getenv("GITPOD_WORKSPACE_ID") != ""
}

func URLForPort(ctx context.Context, port string) (string, error) {
	buf := bytes.Buffer{}
	if err := exec.Exec(ctx, []string{"gp", "url", port}, exec.StepOption(step.Stdout(&buf))); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}
