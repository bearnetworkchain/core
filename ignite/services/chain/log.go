package chain

import (
	"io"
	"os"
	"strings"

	"github.com/bearnetworkchain/core/ignite/pkg/lineprefixer"
	"github.com/bearnetworkchain/core/ignite/pkg/prefixgen"
)

//prefixes 保存日誌消息的前綴配置。
var prefixes = map[logType]struct {
	Name  string
	Color uint8
}{
	logStarport: {"starport", 202},
	logBuild:    {"build", 203},
	logAppd:     {"%s daemon", 204},
}

// logType 表示不同類型的日誌。
type logType int

const (
	logStarport logType = iota
	logBuild
	logAppd
)

type std struct {
	out, err io.Writer
}

// std 返回 stdout 和 stderr 以按 logType 輸出日誌。
func (c *Chain) stdLog() std {
	prefixed := func(w io.Writer) *lineprefixer.Writer {
		var (
			prefix    = prefixes[logStarport]
			prefixStr string
			options   = prefixgen.Common(prefixgen.Color(prefix.Color))
			gen       = prefixgen.New(prefix.Name, options...)
		)
		if strings.Count(prefix.Name, "%s") > 0 {
			prefixStr = gen.Gen(c.app.Name)
		} else {
			prefixStr = gen.Gen()
		}
		return lineprefixer.NewWriter(w, func() string { return prefixStr })
	}
	var (
		stdout io.Writer = prefixed(c.stdout)
		stderr io.Writer = prefixed(c.stderr)
	)
	if c.logLevel == LogRegular {
		stdout = os.Stdout
		stderr = os.Stderr
	}
	return std{
		out: stdout,
		err: stderr,
	}
}

func (c *Chain) genPrefix(logType logType) string {
	prefix := prefixes[logType]

	return prefixgen.
		New(prefix.Name, prefixgen.Common(prefixgen.Color(prefix.Color))...).
		Gen(c.app.Name)
}
