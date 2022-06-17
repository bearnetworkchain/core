package placeholder

import (
	"fmt"
	"strings"

	"github.com/bearnetworkchain/core/ignite/pkg/validation"
)

var _ validation.Error = (*MissingPlaceholdersError)(nil)

// MissingPlaceholdersError 當源文件缺少佔位符時用作錯誤
type MissingPlaceholdersError struct {
	missing          iterableStringSet
	additionalInfo   string
	additionalErrors error
}

// 如果兩個錯誤具有相同的缺失佔位符列表，則為真。
func (e *MissingPlaceholdersError) Is(err error) bool {
	other, ok := err.(*MissingPlaceholdersError)
	if !ok {
		return false
	}
	if len(other.missing) != len(e.missing) {
		return false
	}
	for i := range e.missing {
		if e.missing[i] != other.missing[i] {
			return false
		}
	}
	return true
}

// Error實現錯誤接口
func (e *MissingPlaceholdersError) Error() string {
	var b strings.Builder
	b.WriteString("missing placeholders: ")
	e.missing.Iterate(func(i int, element string) bool {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(element)
		return true
	})
	return b.String()
}

// ValidationInfo 實現validation.Error接口
func (e *MissingPlaceholdersError) ValidationInfo() string {
	var b strings.Builder
	b.WriteString("Missing placeholders:\n\n")
	e.missing.Iterate(func(i int, element string) bool {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(element)
		return true
	})
	if e.additionalInfo != "" {
		b.WriteString("\n\n")
		b.WriteString(e.additionalInfo)
	}
	if e.additionalErrors != nil {
		b.WriteString("\n\nAdditional errors: ")
		b.WriteString(e.additionalErrors.Error())
	}
	return b.String()
}

var _ validation.Error = (*ValidationMiscError)(nil)

// ValidationMiscError用作與驗證相關的雜項錯誤
type ValidationMiscError struct {
	errors []string
}

// Error 實現錯誤接口
func (e *ValidationMiscError) Error() string {
	return fmt.Sprintf("validation errors: %v", e.errors)
}

// ValidationInfo 實現validation.Error接口
func (e *ValidationMiscError) ValidationInfo() string {
	return fmt.Sprintf("Validation errors:\n\n%v", strings.Join(e.errors, "\n"))
}
