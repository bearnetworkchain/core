package numbers

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	separator = ","
	sepRange  = "-"
)

// ParseList 將逗號分隔的數字和範圍解析為 []uint64。
func ParseList(arg string) ([]uint64, error) {
	result := make([]uint64, 0)
	listNumbers := make(map[uint64]struct{})
	// 按分隔符分割切片
	for _, numberRange := range strings.Split(arg, separator) {
		trimmedRange := strings.TrimSpace(numberRange)
		if trimmedRange == "" {
			continue
		}

		// 按分隔符範圍拆分數字
		numbers := strings.Split(trimmedRange, sepRange)
		switch len(numbers) {
		// 解析單個數字
		case 1:
			trimmed := strings.TrimSpace(numbers[0])
			i, err := strconv.ParseUint(trimmed, 10, 32)
			if err != nil {
				return nil, err
			}
			if _, ok := listNumbers[i]; ok {
				continue
			}
			listNumbers[i] = struct{}{}
			result = append(result, i)

		// 解析一個範圍數（例如：3-7）
		case 2:
			var (
				startN = strings.TrimSpace(numbers[0])
				endN   = strings.TrimSpace(numbers[1])
			)
			if startN == "" {
				startN = endN
			}
			if endN == "" {
				endN = startN
			}
			if startN == "" {
				continue
			}
			start, err := strconv.ParseUint(startN, 10, 32)
			if err != nil {
				return nil, err
			}
			end, err := strconv.ParseUint(endN, 10, 32)
			if err != nil {
				return nil, err
			}
			if start > end {
				return nil, fmt.Errorf("無法解析反向排序範圍: %s", trimmedRange)
			}
			for ; start <= end; start++ {
				if _, ok := listNumbers[start]; ok {
					continue
				}
				listNumbers[start] = struct{}{}
				result = append(result, start)
			}
		default:
			return nil, fmt.Errorf("無法解析數字範圍: %s", trimmedRange)
		}
	}
	return result, nil
}

// List 為每個 uint64 創建一個帶有可選前綴的逗號分隔的 int 列表。
func List(numbers []uint64, prefix string) string {
	var s []string
	for _, n := range numbers {
		s = append(s, fmt.Sprintf("%s%d", prefix, n))
	}
	return strings.Join(s, ", ")
}
