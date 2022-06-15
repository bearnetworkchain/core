package network

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SharePercents []SharePercent

func (sp SharePercents) Empty() bool {
	return len(sp) == 0
}

var rePercentageRequired = regexp.MustCompile(`^[0-9]+.[0-9]*%`)

// SharePercent 代表佔總份額的百分比
type SharePercent struct {
	denom string
	// 為了避免使用帶有浮點數的數字
	// 使用分數表示：297/10000 而不是 2.97%
	nominator, denominator uint64
}

// NewSharePercent創建新的份額百分比表示
func NewSharePercent(denom string, nominator, denominator uint64) (SharePercent, error) {
	if denominator < nominator {
		return SharePercent{}, fmt.Errorf("%q 不能大於 100", denom)
	}
	return SharePercent{
		denom:       denom,
		nominator:   nominator,
		denominator: denominator,
	}, nil
}

// Share根據基礎百分比返回總硬幣份額
func (p SharePercent) Share(total uint64) (sdk.Coin, error) {
	resultNominator := total * p.nominator
	if resultNominator%p.denominator != 0 {
		err := fmt.Errorf("%s 佔總數的份額 %d 不是整數: %f",
			p.denom,
			total,
			float64(resultNominator)/float64(p.denominator),
		)
		return sdk.Coin{}, err
	}
	return sdk.NewInt64Coin(p.denom, int64(resultNominator/p.denominator)), nil
}

// SharePercentFromString 從字符串中解析份額百分比
// 格式：11.87%foo
func SharePercentFromString(str string) (SharePercent, error) {
	// 驗證原始百分比格式
	if len(rePercentageRequired.FindStringIndex(str)) == 0 {
		return SharePercent{}, newInvalidPercentageFormat(str)
	}
	var (
		foo        = strings.Split(str, "%")
		fractional = strings.Split(foo[0], ".")
		denom      = foo[1]
	)

	switch len(fractional) {
	case 1:
		nominator, err := strconv.ParseUint(fractional[0], 10, 64)
		if err != nil {
			return SharePercent{}, newInvalidPercentageFormat(str)
		}
		return NewSharePercent(denom, nominator, 100)
	case 2:
		trimmedFractionalPart := strings.TrimRight(fractional[1], "0")
		nominator, err := strconv.ParseUint(fractional[0]+trimmedFractionalPart, 10, 64)
		if err != nil {
			return SharePercent{}, newInvalidPercentageFormat(str)
		}
		return NewSharePercent(denom, nominator, uintPow(10, uint64(len(trimmedFractionalPart)+2)))

	default:
		return SharePercent{}, newInvalidPercentageFormat(str)
	}
}

// ParseSharePercents 從字符串中解析 SharePercentage 列表
// 格式：12.5% Fu, 10% Bar, 0.15% Bass
func ParseSharePercents(percents string) (SharePercents, error) {
	rawPercentages := strings.Split(percents, ",")
	ps := make([]SharePercent, len(rawPercentages))
	for i, percentage := range rawPercentages {
		sp, err := SharePercentFromString(percentage)
		if err != nil {
			return nil, err
		}
		ps[i] = sp

	}

	return ps, nil
}

func uintPow(x, y uint64) uint64 {
	var result = x
	for i := 1; uint64(i) < y; i++ {
		result *= x
	}
	return result
}

func newInvalidPercentageFormat(s string) error {
	return fmt.Errorf("無效的百分比格式 %s", s)
}
