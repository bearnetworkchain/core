package version

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/blang/semver"
	"github.com/google/go-github/v37/github"

	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/exec"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/gitpod"
	"github.com/ignite-hq/cli/ignite/pkg/xexec"
)

const (
	versionDev     = "development"
	versionNightly = "v0.0.0-nightly"
)

const prefix = "v"

var (
	// Version 是 Ignite CLI 的語義版本。
	Version = versionDev

	// Date 是 Ignite CLI 的構建日期。
	Date = "-"

	// Head 是當前分支的 HEAD。
	Head = "-"
)

// CheckNext 檢查是否有新版本的 Ignite CLI。
func CheckNext(ctx context.Context) (isAvailable bool, version string, err error) {
	if Version == versionDev || Version == versionNightly {
		return false, "", nil
	}

	latest, _, err := github.
		NewClient(nil).
		Repositories.
		GetLatestRelease(ctx, "ignite-hq", "cli")

	if err != nil {
		return false, "", err
	}

	if latest.TagName == nil {
		return false, "", nil
	}

	currentVersion, err := semver.Parse(strings.TrimPrefix(Version, prefix))
	if err != nil {
		return false, "", err
	}

	latestVersion, err := semver.Parse(strings.TrimPrefix(*latest.TagName, prefix))
	if err != nil {
		return false, "", err
	}

	isAvailable = latestVersion.GT(currentVersion)

	return isAvailable, *latest.TagName, nil
}

//Long 生成詳細的版本信息。
func Long(ctx context.Context) string {
	var (
		w = &tabwriter.Writer{}
		b = &bytes.Buffer{}
	)

	write := func(k string, v interface{}) {
		fmt.Fprintf(w, "%s:\t%v\n", k, v)
	}

	w.Init(b, 0, 8, 0, '\t', 0)

	write("熊網鏈版本", Version)
	write("熊網鏈創建日期", Date)
	write("熊網鏈哈希值", Head)

	write("熊網鏈系統", runtime.GOOS)
	write("熊網鏈系統位元", runtime.GOARCH)

	cmdOut := &bytes.Buffer{}

	err := exec.Exec(ctx, []string{"go", "version"}, exec.StepOption(step.Stdout(cmdOut)))
	if err != nil {
		panic(err)
	}
	write("Golang版本", strings.TrimSpace(cmdOut.String()))

	unameCmd := "uname"
	if xexec.IsCommandAvailable(unameCmd) {
		cmdOut.Reset()

		err := exec.Exec(ctx, []string{unameCmd, "-a"}, exec.StepOption(step.Stdout(cmdOut)))
		if err == nil {
			write("熊網鏈系統詳細資訊 -a", strings.TrimSpace(cmdOut.String()))
		}
	}

	if cwd, err := os.Getwd(); err == nil {
		write("建構路徑", cwd)
	}

	write("啟動線上GitPod", gitpod.IsOnGitpod())

	w.Flush()

	return b.String()
}
