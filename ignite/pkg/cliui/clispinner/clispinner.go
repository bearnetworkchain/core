package clispinner

import (
	"io"
	"time"

	"github.com/briandowns/spinner"
)

var (
	refreshRate  = time.Millisecond * 200
	charset      = spinner.CharSets[4]
	spinnerColor = "blue"
)

type Spinner struct {
	sp *spinner.Spinner
}

type (
	Option func(*Options)

	Options struct {
		writer io.Writer
	}
)

// WithWriter 為微調器配置輸出
func WithWriter(w io.Writer) Option {
	return func(options *Options) {
		options.writer = w
	}
}

// New 創建一個新的微調器。
func New(options ...Option) *Spinner {
	o := Options{}
	for _, apply := range options {
		apply(&o)
	}

	underlyingSpinnerOptions := []spinner.Option{}
	if o.writer != nil {
		underlyingSpinnerOptions = append(underlyingSpinnerOptions, spinner.WithWriter(o.writer))
	}

	sp := spinner.New(charset, refreshRate, underlyingSpinnerOptions...)

	sp.Color(spinnerColor)
	s := &Spinner{
		sp: sp,
	}
	return s.SetText("初始化...")
}

// SetText 設置微調器的文本.
func (s *Spinner) SetText(text string) *Spinner {
	s.sp.Lock()
	s.sp.Suffix = " " + text
	s.sp.Unlock()
	return s
}

// SetPrefix 設置微調器的前綴.
func (s *Spinner) SetPrefix(text string) *Spinner {
	s.sp.Lock()
	s.sp.Prefix = text + " "
	s.sp.Unlock()
	return s
}

// SetCharset設置微調器的前綴.
func (s *Spinner) SetCharset(charset []string) *Spinner {
	s.sp.UpdateCharSet(charset)
	return s
}

// SetColor 設置微調器的前綴.
func (s *Spinner) SetColor(color string) *Spinner {
	s.sp.Color(color)
	return s
}

// Start 開始旋轉.
func (s *Spinner) Start() *Spinner {
	s.sp.Start()
	return s
}

// Stop 停止旋轉.
func (s *Spinner) Stop() *Spinner {
	s.sp.Stop()
	s.sp.Prefix = ""
	s.sp.Color(spinnerColor)
	s.sp.UpdateCharSet(charset)
	s.sp.Stop()
	return s
}

func (s *Spinner) IsActive() bool {
	return s.sp.Active()
}
