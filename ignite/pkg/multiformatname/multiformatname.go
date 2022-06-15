// Package multiformatname 提供名稱自動轉換為多個命名約定
package multiformatname

import (
	"errors"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
)

// Name 表示一個名稱，具有多種命名約定表示
// 支持的命名約定有：camel、pascal 和 kebab 大小寫
type Name struct {
	Original   string
	LowerCamel string
	UpperCamel string
	LowerCase  string
	UpperCase  string
	Kebab      string
	Snake      string
}

type Checker func(name string) error

// NewName 從名稱返回一個新的多格式名稱
func NewName(name string, additionalChecks ...Checker) (Name, error) {
	checks := append([]Checker{basicCheckName}, additionalChecks...)

	for _, check := range checks {
		if err := check(name); err != nil {
			return Name{}, err
		}
	}

	return Name{
		Original:   name,
		LowerCamel: strcase.ToLowerCamel(name),
		UpperCamel: strcase.ToCamel(name),
		UpperCase:  strings.ToUpper(name),
		Kebab:      strcase.ToKebab(name),
		Snake:      strcase.ToSnake(name),
		LowerCase:  lowercase(name),
	}, nil
}

// NoNumber 防止在名稱中使用數字
func NoNumber(name string) error {
	for _, c := range name {
		if '0' <= c && c <= '9' {
			return errors.New("名稱不能包含數字")
		}
	}

	return nil
}

// basicCheckName 執行所有名稱通用的基本檢查
func basicCheckName(name string) error {
	if name == "" {
		return errors.New("名稱不能為空")
	}

	// check  characters
	c := name[0]
	authorized := ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
	if !authorized {
		return fmt.Errorf("名稱不能包含 %v 作為第一個字符", string(c))
	}

	for _, c := range name[1:] {
		// 名稱可以包含字母、連字符或下劃線
		authorized := ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') || c == '-' || c == '_'
		if !authorized {
			return fmt.Errorf("名稱不能包含 %v", string(c))
		}
	}

	return nil
}

// lowercase 返回小寫且無特殊字符的名稱
func lowercase(name string) string {
	return strings.ToLower(
		strings.ReplaceAll(
			strings.ReplaceAll(name, "-", ""),
			"_",
			"",
		),
	)
}
