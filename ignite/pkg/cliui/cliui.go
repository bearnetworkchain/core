package cliui

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/manifoldco/promptui"

	"github.com/bearnetworkchain/core/ignite/pkg/cliui/cliquiz"
	"github.com/bearnetworkchain/core/ignite/pkg/cliui/clispinner"
	"github.com/bearnetworkchain/core/ignite/pkg/cliui/entrywriter"
	"github.com/bearnetworkchain/core/ignite/pkg/cliui/icons"
	"github.com/bearnetworkchain/core/ignite/pkg/events"
)

// 會話控制與用戶的命令行交互.
type Session struct {
	ev       events.Bus
	eventsWg *sync.WaitGroup

	spinner *clispinner.Spinner

	in          io.Reader
	out         io.Writer
	printLoopWg *sync.WaitGroup
}

type Option func(s *Session)

// WithOutput 設置會話的輸出流.
func WithOutput(output io.Writer) Option {
	return func(s *Session) {
		s.out = output
	}
}

// WithInput 設置會話的輸入流。
func WithInput(input io.Reader) Option {
	return func(s *Session) {
		s.in = input
	}
}

// New 創建新會話。
func New(options ...Option) Session {
	wg := &sync.WaitGroup{}
	session := Session{
		ev:          events.NewBus(events.WithWaitGroup(wg)),
		in:          os.Stdin,
		out:         os.Stdout,
		eventsWg:    wg,
		printLoopWg: &sync.WaitGroup{},
	}
	for _, apply := range options {
		apply(&session)
	}
	session.spinner = clispinner.New(clispinner.WithWriter(session.out))
	session.printLoopWg.Add(1)
	go session.printLoop()
	return session
}

// StopSpinner 返回會話的事件總線。
func (s Session) EventBus() events.Bus {
	return s.ev
}

// StartSpinner 啟動微調器。
func (s Session) StartSpinner(text string) {
	s.spinner.SetText(text).Start()
}

// StopSpinner 停止微調器。
func (s Session) StopSpinner() {
	s.spinner.Stop()
}

// PauseSpinner 暫停微調器，返回恢復函數以再次啟動暫停的微調器。
func (s Session) PauseSpinner() (mightResume func()) {
	isActive := s.spinner.IsActive()
	f := func() {
		if isActive {
			s.spinner.Start()
		}
	}
	s.spinner.Stop()
	return f
}

// Printf 打印格式化的任意消息。
func (s Session) Printf(format string, a ...interface{}) error {
	s.Wait()
	defer s.PauseSpinner()()
	_, err := fmt.Fprintf(s.out, format, a...)
	return err
}

// Println 打印帶有換行符的任意消息。
func (s Session) Println(messages ...interface{}) error {
	s.Wait()
	defer s.PauseSpinner()()
	_, err := fmt.Fprintln(s.out, messages...)
	return err
}

// PrintSaidNo 在確認提示中給出了通知否定的打印消息
func (s Session) PrintSaidNo() error {
	return s.Println("said no")
}

// Println 打印任意消息
func (s Session) Print(messages ...interface{}) error {
	s.Wait()
	defer s.PauseSpinner()()
	_, err := fmt.Fprint(s.out, messages...)
	return err
}

// Ask 在終端提問並收集答案。
func (s Session) Ask(questions ...cliquiz.Question) error {
	s.Wait()
	defer s.PauseSpinner()()
	return cliquiz.Ask(questions...)
}

// AskConfirm 在終端中詢問是/否問題。
func (s Session) AskConfirm(message string) error {
	s.Wait()
	defer s.PauseSpinner()()
	prompt := promptui.Prompt{
		Label:     message,
		IsConfirm: true,
	}
	_, err := prompt.Run()
	return err
}

// PrintTable 打印表格數據。
func (s Session) PrintTable(header []string, entries ...[]string) error {
	s.Wait()
	defer s.PauseSpinner()()
	return entrywriter.MustWrite(s.out, header, entries...)
}

// Wait 阻塞，直到處理完所有排隊的事件。
func (s Session) Wait() {
	s.eventsWg.Wait()
}

// Cleanup 確保微調器停止並且打印循環正確退出。
func (s Session) Cleanup() {
	s.StopSpinner()
	s.ev.Shutdown()
	s.printLoopWg.Wait()
}

// printLoop 處理事件。
func (s Session) printLoop() {
	for event := range s.ev.Events() {
		switch event.Status {
		case events.StatusOngoing:
			s.StartSpinner(event.Text())

		case events.StatusDone:
			if event.Icon == "" {
				event.Icon = icons.OK
			}
			s.StopSpinner()
			fmt.Fprintf(s.out, "%s %s\n", event.Icon, event.Text())

		case events.StatusNeutral:
			resume := s.PauseSpinner()
			fmt.Fprintf(s.out, event.Text())
			resume()
		}

		s.eventsWg.Done()
	}
	s.printLoopWg.Done()
}
