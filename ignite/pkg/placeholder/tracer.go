package placeholder

import (
	"strings"
)

type iterableStringSet map[string]struct{}

func (set iterableStringSet) Iterate(f func(i int, element string) bool) {
	i := 0
	for key := range set {
		if !f(i, key) {
			return
		}
		i++
	}
}

func (set iterableStringSet) Add(item string) {
	set[item] = struct{}{}
}

// Option用於配置會話。
type Option func(*Tracer)

// WithAdditionalInfo 將信息附加到驗證錯誤。
func WithAdditionalInfo(info string) Option {
	return func(s *Tracer) {
		s.additionalInfo = info
	}
}

// New 使用提供的選項實例化 Session。
func New(opts ...Option) *Tracer {
	s := &Tracer{missing: iterableStringSet{}}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type Replacer interface {
	Replace(content, placeholder, replacement string) string
	ReplaceAll(content, placeholder, replacement string) string
	ReplaceOnce(content, placeholder, replacement string) string
	AppendMiscError(miscError string)
}

// Tracer跟踪丟失的佔位符或與文件修改相關的其他問題。
type Tracer struct {
	missing        iterableStringSet
	miscErrors     []string
	additionalInfo string
}

// ReplaceAll用替換字符串替換內容中的所有佔位符。
func (t *Tracer) ReplaceAll(content, placeholder, replacement string) string {
	if strings.Count(content, placeholder) == 0 {
		t.missing.Add(placeholder)
		return content
	}
	return strings.ReplaceAll(content, placeholder, replacement)
}

// Replace 內容中的佔位符，替換字符串一次。
func (t *Tracer) Replace(content, placeholder, replacement string) string {
	// 注意（dshulyak）我們將計算兩次。一次在這裡，第二次在字符串中。替換
	// 如果結果是問題，請從 strings.Replace 複製代碼。
	if strings.Count(content, placeholder) == 0 {
		t.missing.Add(placeholder)
		return content
	}
	return strings.Replace(content, placeholder, replacement, 1)
}

// 僅當內容中尚未找到替換時，ReplaceOnce 才會替換內容中的佔位符。
func (t *Tracer) ReplaceOnce(content, placeholder, replacement string) string {
	if !strings.Contains(content, replacement) {
		return t.Replace(content, placeholder, replacement)
	}
	return content
}

// AppendMiscError 允許在文件修改期間跟踪與丟失佔位符無關的錯誤
func (t *Tracer) AppendMiscError(miscError string) {
	t.miscErrors = append(t.miscErrors, miscError)
}

// Err 如果在執行期間缺少任何占位符。
func (t *Tracer) Err() error {
	// miscellaneous 錯誤表示阻止與丟失佔位符無關的源修改的錯誤
	var miscErrors error
	if len(t.miscErrors) > 0 {
		miscErrors = &ValidationMiscError{
			errors: t.miscErrors,
		}
	}

	if len(t.missing) > 0 {
		missing := iterableStringSet{}
		for key := range t.missing {
			missing.Add(key)
		}
		return &MissingPlaceholdersError{
			missing:          missing,
			additionalInfo:   t.additionalInfo,
			additionalErrors: miscErrors,
		}
	}

	// 如果沒有丟失佔位符但仍然有雜項錯誤，則返回它們
	return miscErrors
}
