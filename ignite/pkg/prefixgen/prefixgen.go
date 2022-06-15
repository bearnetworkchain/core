// Package prefixgen 是日誌消息的前綴生成助手和任何其他種類。
package prefixgen

import (
	"fmt"
	"strings"

	"github.com/gookit/color"
)

// Prefixer生成前綴。
type Prefixer struct {
	format           string
	color            uint8
	left, right      string
	convertUppercase bool
}

//選項配置前綴。
type Option func(p *Prefixer)

//顏色將顏色設置為前綴。
func Color(color uint8) Option {
	return func(p *Prefixer) {
		p.color = color
	}
}

//SquareBrackets 將方括號添加到前綴。
func SquareBrackets() Option {
	return func(p *Prefixer) {
		p.left = "["
		p.right = "]"
	}
}

//SpaceRight 將權限空間添加到前綴。
func SpaceRight() Option {
	return func(p *Prefixer) {
		p.right += " "
	}
}

//大寫將前綴格式化為大寫。
func Uppercase() Option {
	return func(p *Prefixer) {
		p.convertUppercase = true
	}
}

//Common 擁有一些常見的前綴選項並擴展了這些選項
//給定選項的選項。
func Common(options ...Option) []Option {
	return append([]Option{
		SquareBrackets(),
		SpaceRight(),
		Uppercase(),
	}, options...)
}

// New 使用格式和選項創建一個新的前綴。
//Format 是一種類似於 fmt.Sprintf() 的格式，用於動態創建前綴文本
//如所須。
func New(format string, options ...Option) *Prefixer {
	p := &Prefixer{
		format: format,
	}
	for _, o := range options {
		o(p)
	}
	return p
}

//Gen 通過將 s 應用於 New() 期間給出的格式來生成新前綴。
func (p *Prefixer) Gen(s ...interface{}) string {
	format := p.format
	format = p.left + format
	format += p.right
	prefix := fmt.Sprintf(format, s...)
	if p.convertUppercase {
		prefix = strings.ToUpper(prefix)
	}
	if p.color != 0 {
		return color.C256(p.color).Sprint(prefix)
	}
	return prefix
}
