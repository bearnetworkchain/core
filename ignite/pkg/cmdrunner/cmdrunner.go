package cmdrunner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"golang.org/x/sync/errgroup"

	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/goenv"
)

// Runner 是運行命令的對象
type Runner struct {
	endSignal   os.Signal
	stdout      io.Writer
	stderr      io.Writer
	stdin       io.Reader
	workdir     string
	runParallel bool
}

// Option 定義運行命令的選項
type Option func(*Runner)

// DefaultStdout 為要運行的命令提供默認標準輸出
func DefaultStdout(writer io.Writer) Option {
	return func(r *Runner) {
		r.stdout = writer
	}
}

// DefaultStderr 為要運行的命令提供默認的標準錯誤
func DefaultStderr(writer io.Writer) Option {
	return func(r *Runner) {
		r.stderr = writer
	}
}

// DefaultStdin 為要運行的命令提供默認標準輸入
func DefaultStdin(reader io.Reader) Option {
	return func(r *Runner) {
		r.stdin = reader
	}
}

// DefaultWorkdir 為要運行的命令提供默認工作目錄
func DefaultWorkdir(path string) Option {
	return func(r *Runner) {
		r.workdir = path
	}
}

// RunParallel 允許命令同時運行
func RunParallel() Option {
	return func(r *Runner) {
		r.runParallel = true
	}
}

// EndSignal 將 s 配置為向進程發出信號以結束它們。
func EndSignal(s os.Signal) Option {
	return func(r *Runner) {
		r.endSignal = s
	}
}

// New 返回一個新的命令運行器
func New(options ...Option) *Runner {
	runner := &Runner{
		endSignal: os.Interrupt,
	}
	for _, apply := range options {
		apply(runner)
	}
	return runner
}

// Run 阻塞，直到所有步驟都完成執行。
func (r *Runner) Run(ctx context.Context, steps ...*step.Step) error {
	if len(steps) == 0 {
		return nil
	}
	g, ctx := errgroup.WithContext(ctx)
	for _, step := range steps {
		// 複製 s 到一個新的變量來分配一個新的地址
		// 所以我們可以安全地在這個循環中產生的 goroutines 中使用它。
		step := step
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := step.PreExec(); err != nil {
			return err
		}
		runPostExecs := func(processErr error) error {
			// 如果上下文被取消，那麼我們可以忽略退出錯誤處理，因為它應該因為取消而退出。
			var err error
			ctxErr := ctx.Err()
			if ctxErr != nil {
				err = ctxErr
			} else {
				err = processErr
			}
			for _, exec := range step.PostExecs {
				if err := exec(err); err != nil {
					return err
				}
			}
			if len(step.PostExecs) > 0 {
				return nil
			}
			return err
		}
		command := r.newCommand(step)
		startErr := command.Start()
		if startErr != nil {
			if err := runPostExecs(startErr); err != nil {
				return err
			}
			continue
		}
		go func() {
			<-ctx.Done()
			command.Signal(r.endSignal)
		}()
		if err := step.InExec(); err != nil {
			return err
		}
		if len(step.WriteData) > 0 {
			if _, err := command.Write(step.WriteData); err != nil {
				return err
			}
		}
		if r.runParallel {
			g.Go(func() error {
				return runPostExecs(command.Wait())
			})
		} else if err := runPostExecs(command.Wait()); err != nil {
			return err
		}
	}
	return g.Wait()
}

// Executor 表示要執行的命令
type Executor interface {
	Wait() error
	Start() error
	Signal(os.Signal)
	Write(data []byte) (n int, err error)
}

// dummyExecutor 是一個什麼都不做的執行者
type dummyExecutor struct{}

func (e *dummyExecutor) Start() error { return nil }

func (e *dummyExecutor) Wait() error { return nil }

func (e *dummyExecutor) Signal(os.Signal) {}

func (e *dummyExecutor) Write([]byte) (int, error) { return 0, nil }

// cmdSignal 是一個帶有信號處理的執行器
type cmdSignal struct {
	*exec.Cmd
}

func (e *cmdSignal) Signal(s os.Signal) { e.Cmd.Process.Signal(s) }

func (e *cmdSignal) Write(data []byte) (n int, err error) { return 0, nil }

// cmdSignalWithWriter 是具有信號處理功能的執行器，可以寫入標準輸入
type cmdSignalWithWriter struct {
	*exec.Cmd
	w io.WriteCloser
}

func (e *cmdSignalWithWriter) Signal(s os.Signal) { e.Cmd.Process.Signal(s) }

func (e *cmdSignalWithWriter) Write(data []byte) (n int, err error) {
	defer e.w.Close()
	return e.w.Write(data)
}

// newCommand 返回要執行的新命令
func (r *Runner) newCommand(step *step.Step) Executor {
	// 在空命令的情況下返回一個虛擬執行器
	if step.Exec.Command == "" {
		return &dummyExecutor{}
	}
	var (
		stdout = step.Stdout
		stderr = step.Stderr
		stdin  = step.Stdin
		dir    = step.Workdir
	)

	//定義標準輸入和輸出
	if stdout == nil {
		stdout = r.stdout
	}
	if stderr == nil {
		stderr = r.stderr
	}
	if stdin == nil {
		stdin = r.stdin
	}
	if dir == "" {
		dir = r.workdir
	}

	// 初始化命令
	command := exec.Command(step.Exec.Command, step.Exec.Args...)
	command.Stdout = stdout
	command.Stderr = stderr
	command.Dir = dir
	command.Env = append(os.Environ(), step.Env...)
	command.Env = append(command.Env, Env("PATH", goenv.Path()))

	// 如果提供了自定義標準輸入，它將作為命令的標準輸入
	if stdin != nil {
		command.Stdin = stdin
		return &cmdSignal{command}
	}

	// 如果沒有自定義標準輸入，執行器可以寫入程序的標準輸入
	writer, err := command.StdinPipe()
	if err != nil {
		// TODO 不要驚慌
		panic(err)
	}
	return &cmdSignalWithWriter{command, writer}
}

// Env 從 key 和 val 返回一個新的 env var 值。
func Env(key, val string) string {
	return fmt.Sprintf("%s=%s", key, val)
}
