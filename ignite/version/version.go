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

	"github.com/bearnetworkchain/core/ignite/pkg/cmdrunner/exec"
	"github.com/bearnetworkchain/core/ignite/pkg/cmdrunner/step"
	"github.com/bearnetworkchain/core/ignite/pkg/gitpod"
	"github.com/bearnetworkchain/core/ignite/pkg/xexec"
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

	write("Ignite CLI version", Version)
	write("Ignite CLI build date", Date)
	write("Ignite CLI source hash", Head)

	write("Your OS", runtime.GOOS)
	write("Your arch", runtime.GOARCH)

	cmdOut := &bytes.Buffer{}

	err := exec.Exec(ctx, []string{"go", "version"}, exec.StepOption(step.Stdout(cmdOut)))
	if err != nil {
		panic(err)
	}
	write("你的GO版本", strings.TrimSpace(cmdOut.String()))

	unameCmd := "uname"
	if xexec.IsCommandAvailable(unameCmd) {
		cmdOut.Reset()

		err := exec.Exec(ctx, []string{unameCmd, "-a"}, exec.StepOption(step.Stdout(cmdOut)))
		if err == nil {
			write("Your uname -a", strings.TrimSpace(cmdOut.String()))
		}
	}

	if cwd, err := os.Getwd(); err == nil {
		write("Your cwd", cwd)
	}

	write("Is on Gitpod", gitpod.IsOnGitpod())

	w.Flush()

	return b.String()
}
