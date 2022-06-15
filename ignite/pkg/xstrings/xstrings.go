package xstrings

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// AllOrSomeFilter 從列表中過濾出元素，因為它們出現在 filterList 和
// 返回剩餘的。
// 如果過濾器列表為空，則返回列表中的所有元素。
func AllOrSomeFilter(list, filterList []string) []string {
	if len(filterList) == 0 {
		return list
	}

	var elems []string

	for _, elem := range list {
		if !SliceContains(filterList, elem) {
			elems = append(elems, elem)
		}
	}

	return elems
}

// 如果 s 是 ss 的成員，則 SliceContains 返回 true。
func SliceContains(ss []string, s string) bool {
	for _, e := range ss {
		if e == s {
			return true
		}
	}

	return false
}

// List 返回在 do 返回的值之後捕獲的字符串切片，即
// 調用 n 次。
func List(n int, do func(i int) string) []string {
	var list []string

	for i := 0; i < n; i++ {
		list = append(list, do(i))
	}

	return list
}

// FormatUsername 格式化用戶名以使其可用作變量
func FormatUsername(s string) string {
	return NoDash(NoNumberPrefix(s))
}

//NoDash 從字符串中刪除破折號
func NoDash(s string) string {
	return strings.ReplaceAll(s, "-", "")
}

// 如果 NoNumberPrefix 以數字開頭，則在字符串開頭添加下劃線
// 這用於原始文件模板的包，因為包名不能以數字開頭。
func NoNumberPrefix(s string) string {
	//檢查它是否以數字開頭
	if unicode.IsDigit(rune(s[0])) {
		return "_" + s
	}
	return s
}

// 標題返回字符串 s 的副本，其中包含以單詞開頭的所有 Unicode 字母
// 映射到他們的 Unicode 標題大小寫。
func Title(title string) string {
	return cases.Title(language.English).String(title)
}
