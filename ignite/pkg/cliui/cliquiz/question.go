// Package cliquiz is a tool to collect answers from the users on cli.
package cliquiz

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/pflag"
)

// 當第二個答案與第一個答案不同時，返回錯誤確認失敗。
var ErrConfirmationFailed = errors.New("未能確認，你的答案不一樣")

// Question holds information on what to ask to user and where
// the answer stored at.
type Question struct {
	question      string
	defaultAnswer interface{}
	answer        interface{}
	hidden        bool
	shouldConfirm bool
	required      bool
}

// 選項配置問題。
type Option func(*Question)

// DefaultAnswer 設置問題的默認答案。
func DefaultAnswer(answer interface{}) Option {
	return func(q *Question) {
		q.defaultAnswer = answer
	}
}

// 必填項將答案標記為必填項。
func Required() Option {
	return func(q *Question) {
		q.required = true
	}
}

//HideAnswer 隱藏答案以防止機密信息洩露。
func HideAnswer() Option {
	return func(q *Question) {
		q.hidden = true
	}
}

// GetConfirmation 提示確認給定的答案。
func GetConfirmation() Option {
	return func(q *Question) {
		q.shouldConfirm = true
	}
}

// NewQuestion 創建一個新問題。
func NewQuestion(question string, answer interface{}, options ...Option) Question {
	q := Question{
		question: question,
		answer:   answer,
	}
	for _, o := range options {
		o(&q)
	}
	return q
}

func ask(q Question) error {
	var prompt survey.Prompt

	if !q.hidden {
		input := &survey.Input{
			Message: q.question,
		}
		if !q.required {
			input.Message += " (optional)"
		}
		if q.defaultAnswer != nil {
			input.Default = fmt.Sprintf("%v", q.defaultAnswer)
		}
		prompt = input
	} else {
		prompt = &survey.Password{
			Message: q.question,
		}
	}

	if err := survey.AskOne(prompt, q.answer); err != nil {
		return err
	}

	isValid := func() bool {
		if answer, ok := q.answer.(string); ok {
			if strings.TrimSpace(answer) == "" {
				return false
			}
		}
		if reflect.ValueOf(q.answer).Elem().IsZero() {
			return false
		}
		return true
	}

	if q.required && !isValid() {
		fmt.Println("此信息為必填項，請重試:")

		if err := ask(q); err != nil {
			return err
		}
	}

	return nil
}

//提出問題並收集答案。
func Ask(question ...Question) (err error) {
	defer func() {
		if err == terminal.InterruptErr {
			err = context.Canceled
		}
	}()

	for _, q := range question {
		if err := ask(q); err != nil {
			return err
		}

		if q.shouldConfirm {
			var secondAnswer string

			options := []Option{}
			if q.required {
				options = append(options, Required())
			}
			if q.hidden {
				options = append(options, HideAnswer())
			}
			if err := ask(NewQuestion("確認 "+q.question, &secondAnswer, options...)); err != nil {
				return err
			}

			t := reflect.TypeOf(secondAnswer)
			compAnswer := reflect.ValueOf(q.answer).Elem().Convert(t).String()
			if secondAnswer != compAnswer {
				return ErrConfirmationFailed
			}
		}
	}
	return nil
}

// Flag 表示一個 cmd 標誌。
type Flag struct {
	Name       string
	IsRequired bool
}

// NewFlag 創建一個新的 flag.
func NewFlag(name string, isRequired bool) Flag {
	return Flag{name, isRequired}
}

// ValuesFromFlagsOrAsk 返回 map[string]string 中的標誌值，其中 map's
// key是flag的名字，value是flag的值。
// 提供時，通過命令收集值，否則通過提示向用戶詢問。
// 提示時用作消息的標題。
func ValuesFromFlagsOrAsk(fset *pflag.FlagSet, title string, flags ...Flag) (values map[string]string, err error) {
	values = make(map[string]string)

	answers := make(map[string]*string)
	var questions []Question

	for _, f := range flags {
		flag := fset.Lookup(f.Name)
		if flag == nil {
			return nil, fmt.Errorf("flag %q 沒有定義", f.Name)
		}
		if value, _ := fset.GetString(f.Name); value != "" {
			values[f.Name] = value
			continue
		}

		var value string
		answers[f.Name] = &value

		var options []Option
		if f.IsRequired {
			options = append(options, Required())
		}
		questions = append(questions, NewQuestion(flag.Usage, &value, options...))
	}

	if len(questions) > 0 && title != "" {
		fmt.Println(title)
	}
	if err := Ask(questions...); err != nil {
		return values, err
	}

	for name, answer := range answers {
		values[name] = *answer
	}

	return values, nil
}
