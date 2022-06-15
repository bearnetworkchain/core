// Package events 為包提供將其狀態記錄為事件的功能
// 供其他人以有意義的方式消費和展示給最終用戶。
package events

import (
	"fmt"
	"sync"

	"github.com/gookit/color"
)

type (
	// Event 代表一個狀態。
	Event struct {
		// 狀態描述。
		Description string

		// Status 顯示事件的當前狀態。
		Status Status

		// TextColor 的文本。
		TextColor color.Color

		// Icon 的文本。
		Icon string
	}

	// Status 顯示狀態是正在進行還是完成。
	Status int

	// Option 事件選項
	Option func(*Event)
)

const (
	StatusOngoing Status = iota
	StatusDone
	StatusNeutral
)

// TextColor 設置文本顏色
func TextColor(c color.Color) Option {
	return func(e *Event) {
		e.TextColor = c
	}
}

// Icon 設置文本圖標前綴。
func Icon(icon string) Option {
	return func(e *Event) {
		e.Icon = icon
	}
}

// New 使用給定的配置創建一個新事件。
func New(status Status, description string, options ...Option) Event {
	ev := Event{Status: status, Description: description}
	for _, applyOption := range options {
		applyOption(&ev)
	}
	return ev
}

// NewOngoing 創建一個新的 StatusOngoing 事件。
func NewOngoing(description string) Event {
	return New(StatusOngoing, description)
}

// NewNeutral 創建一個新的 StatusNeutral 事件。
func NewNeutral(description string) Event {
	return New(StatusNeutral, description)
}

// NewDone 創建一個新的 StatusDone 事件。
func NewDone(description, icon string) Event {
	return New(StatusDone, description, Icon(icon))
}

// IsOngoing 檢查觸發此事件的狀態更改是否仍在進行中。
func (e Event) IsOngoing() bool {
	return e.Status == StatusOngoing
}

// Text 返回事件的文本狀態。
func (e Event) Text() string {
	text := e.Description
	if e.IsOngoing() {
		text = fmt.Sprintf("%s...", e.Description)
	}
	return e.TextColor.Render(text)
}

// Bus 是一個發送/接收事件總線。
type (
	Bus struct {
		evchan chan Event
		buswg  *sync.WaitGroup
	}

	BusOption func(*Bus)
)

// WithWaitGroup 設置等待組，如果事件總線不為空則阻塞。
func WithWaitGroup(wg *sync.WaitGroup) BusOption {
	return func(bus *Bus) {
		bus.buswg = wg
	}
}

// WithCustomBufferSize 配置底層總線通道的緩衝區大小
func WithCustomBufferSize(size int) BusOption {
	return func(bus *Bus) {
		bus.evchan = make(chan Event, size)
	}
}

// NewBus 創建一個新的事件總線來發送/接收事件。
func NewBus(options ...BusOption) Bus {
	bus := Bus{
		evchan: make(chan Event),
	}

	for _, apply := range options {
		apply(&bus)
	}

	return bus
}

// Send 向總線發送一個新事件。
func (b Bus) Send(e Event) {
	if b.evchan == nil {
		return
	}
	if b.buswg != nil {
		b.buswg.Add(1)
	}
	b.evchan <- e
}

// Events 返回帶有事件的 go channel 只能用於讀取。
func (b *Bus) Events() <-chan Event {
	return b.evchan
}

// Shutdown 關閉事件總線。
func (b Bus) Shutdown() {
	if b.evchan == nil {
		return
	}
	close(b.evchan)
}
