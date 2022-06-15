package protoanalysis

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/emicklei/proto"
)

type builder struct {
	p pkg
}

//build 將一個低級的 prot pkg 變成一個高級的 Package。
func build(p pkg) Package {
	br := builder{p}

	pk := Package{
		Name:     p.name,
		Path:     p.dir,
		Files:    br.buildFiles(),
		Messages: br.buildMessages(),
		Services: br.toServices(p.services()),
	}

	for _, option := range p.options() {
		if option.Name == optionGoPkg {
			pk.GoImportName = option.Constant.Source
			break
		}
	}

	return pk
}

func (b builder) buildFiles() (files []File) {
	for _, f := range b.p.files {
		files = append(files, File{f.path, f.imports})
	}

	return
}

func (b builder) buildMessages() (messages []Message) {
	for _, f := range b.p.files {
		for _, message := range f.messages {

			// 查找最高的字段號
			var highestFieldNumber int
			for _, elem := range message.Elements {
				field, ok := elem.(*proto.NormalField)
				if ok {
					if field.Sequence > highestFieldNumber {
						highestFieldNumber = field.Sequence
					}
				}
			}

// 一些原始消息可能在另一個原始消息中定義。
// 為了表示這些類型，使用了下劃線。
// 例如如果 C 消息在 B 中，B 在 A 中：A_B_C。
			var (
				name   = message.Name
				parent = message.Parent
			)
			for {
				if parent == nil {
					break
				}

				parentMessage, ok := parent.(*proto.Message)
				if !ok {
					break
				}

				name = fmt.Sprintf("%s_%s", parentMessage.Name, name)
				parent = parentMessage.Parent
			}

			messages = append(messages, Message{
				Name:               name,
				Path:               f.path,
				HighestFieldNumber: highestFieldNumber,
			})
		}
	}

	return messages
}

func (b builder) toServices(ps []*proto.Service) (services []Service) {
	for _, service := range ps {
		s := Service{
			Name:     service.Name,
			RPCFuncs: b.elementsToRPCFunc(service.Elements),
		}

		services = append(services, s)
	}

	return
}

func (b builder) elementsToRPCFunc(elems []proto.Visitee) (rpcFuncs []RPCFunc) {
	for _, el := range elems {
		rpc, ok := el.(*proto.RPC)
		if !ok {
			continue
		}

		var requestMessage *proto.Message

		for _, message := range b.p.messages() {
			if message.Name != rpc.RequestType {
				continue
			}
			requestMessage = message
		}

		if requestMessage == nil {
			continue
		}

		rf := RPCFunc{
			Name:        rpc.Name,
			RequestType: rpc.RequestType,
			ReturnsType: rpc.ReturnsType,
			HTTPRules:   b.elementsToHTTPRules(requestMessage, rpc.Elements),
		}

		rpcFuncs = append(rpcFuncs, rf)
	}

	return rpcFuncs
}

func (b builder) elementsToHTTPRules(requestMessage *proto.Message, elems []proto.Visitee) (httpRules []HTTPRule) {
	for _, el := range elems {
		option, ok := el.(*proto.Option)
		if !ok {
			continue
		}
		if !strings.Contains(option.Name, "google.api.http") {
			continue
		}

		httpRules = append(httpRules, b.constantToHTTPRules(requestMessage, option.Constant)...)
	}

	return
}

var urlParamRe = regexp.MustCompile(`(?m){(.+?)}`)

func (b builder) constantToHTTPRules(requestMessage *proto.Message, constant proto.Literal) (httpRules []HTTPRule) {
	// find out the endpoint template.
	endpoint := constant.Source

	if endpoint == "" {
		for key, val := range constant.Map {
			switch key {
			case
				"get",
				"post",
				"put",
				"patch",
				"delete":
				endpoint = val.Source
			}
			if endpoint != "" {
				break
			}
		}
	}

	// find out url params.
	var params []string

	match := urlParamRe.FindAllStringSubmatch(endpoint, -1)
	for _, item := range match {
		params = append(params, item[1])
	}

	// 計算 url 參數、查詢參數和正文字段計數。
	var (
		messageFieldsCount = b.messageFieldsCount(requestMessage)
		paramsCount        = len(params)
		bodyFieldsCount    int
	)

	if body, ok := constant.Map["body"]; ok { // 檢查是否指定了正文。
		if body.Source == "*" { // 意味著每個規範不應該有查詢參數。
			bodyFieldsCount = messageFieldsCount - paramsCount
		} else if body.Source != "" {
			bodyFieldsCount = 1 // 表示正文字段分組在單個頂級字段下。
		}
	}

	queryParamsCount := messageFieldsCount - paramsCount - bodyFieldsCount

	// 創建 HTTP 規則並將其添加到列表中。
	httpRule := HTTPRule{
		Params:   params,
		HasQuery: queryParamsCount > 0,
		HasBody:  bodyFieldsCount > 0,
	}

	httpRules = append(httpRules, httpRule)

	// search for nested HTTP rules.
	if constant, ok := constant.Map["additional_bindings"]; ok {
		httpRules = append(httpRules, b.constantToHTTPRules(requestMessage, *constant)...)
	}

	return httpRules
}

func (b builder) messageFieldsCount(message *proto.Message) (count int) {
	for _, el := range message.Elements {
		switch el.(type) {
		case
			*proto.NormalField,
			*proto.MapField,
			*proto.OneOfField:
			count++
		}
	}

	return
}
