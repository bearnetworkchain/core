package scaffolder

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/ignite-hq/cli/ignite/pkg/multiformatname"
	"github.com/ignite-hq/cli/ignite/pkg/protoanalysis"
	"github.com/ignite-hq/cli/ignite/templates/field/datatype"
)

const (
	componentType    = "type"
	componentMessage = "message"
	componentQuery   = "query"
	componentPacket  = "packet"

	protoFolder = "proto"
)

// checkComponentValidity執行所有組件通用的各種檢查，以驗證它是否可以搭建
func checkComponentValidity(appPath, moduleName string, compName multiformatname.Name, noMessage bool) error {
	ok, err := moduleExists(appPath, moduleName)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("模塊 %s 不存在", moduleName)
	}

	// Ensure名稱有效，否則會生成錯誤代碼
	if err := checkForbiddenComponentName(compName); err != nil {
		return fmt.Errorf("%s 不能用作組件名稱: %s", compName.LowerCamel, err.Error())
	}

	// 檢查組件名稱尚未使用
	return checkComponentCreated(appPath, moduleName, compName, noMessage)
}

// checkForbiddenComponentName如果名稱被禁止作為組件名稱，則返回 true
func checkForbiddenComponentName(name multiformatname.Name) error {
	// 檢查腳手架代碼中已使用的名稱
	switch name.LowerCase {
	case
		"oracle",
		"logger",
		"keeper",
		"query",
		"genesis",
		"types",
		"tx",
		datatype.TypeCustom:
		return fmt.Errorf("%s 由熊網鏈腳手架使用", name.LowerCamel)
	}

	if strings.HasSuffix(name.LowerCase, "test") {
		return errors.New(`名稱不能以“test”結尾"`)
	}

	return checkGoReservedWord(name.LowerCamel)
}

// checkGoReservedWord檢查名稱是否不能使用，因為它是 go 保留關鍵字
func checkGoReservedWord(name string) error {
	// Check keyword or literal
	if token.Lookup(name).IsKeyword() {
		return fmt.Errorf("%s 是一個 Go 關鍵字", name)
	}

	// Check with builtin identifier
	switch name {
	case
		"panic",
		"recover",
		"append",
		"bool",
		"byte",
		"cap",
		"close",
		"complex",
		"complex64",
		"complex128",
		"uint16",
		"copy",
		"false",
		"float32",
		"float64",
		"imag",
		"int",
		"int8",
		"int16",
		"uint32",
		"int32",
		"int64",
		"iota",
		"len",
		"make",
		"new",
		"nil",
		"uint64",
		"print",
		"println",
		"real",
		"string",
		"true",
		"uint",
		"uint8",
		"uintptr":
		return fmt.Errorf("%s 是 Go 內置標識符", name)
	}
	return nil
}

// checkComponentCreated檢查組件是否已經在項目中使用熊網鏈創建
func checkComponentCreated(appPath, moduleName string, compName multiformatname.Name, noMessage bool) (err error) {

	// 將要檢查的類型與腳手架該類型的組件相關聯
	typesToCheck := map[string]string{
		compName.UpperCamel:                           componentType,
		"QueryAll" + compName.UpperCamel + "Request":  componentType,
		"QueryAll" + compName.UpperCamel + "Response": componentType,
		"QueryGet" + compName.UpperCamel + "Request":  componentType,
		"QueryGet" + compName.UpperCamel + "Response": componentType,
		"Query" + compName.UpperCamel + "Request":     componentQuery,
		"Query" + compName.UpperCamel + "Response":    componentQuery,
		compName.UpperCamel + "PacketData":            componentPacket,
	}

	if !noMessage {
		typesToCheck["MsgCreate"+compName.UpperCamel] = componentType
		typesToCheck["MsgUpdate"+compName.UpperCamel] = componentType
		typesToCheck["MsgDelete"+compName.UpperCamel] = componentType
		typesToCheck["Msg"+compName.UpperCamel] = componentMessage
		typesToCheck["MsgSend"+compName.UpperCamel] = componentPacket
	}

	absPath, err := filepath.Abs(filepath.Join(appPath, "x", moduleName, "types"))
	if err != nil {
		return err
	}
	fileSet := token.NewFileSet()
	all, err := parser.ParseDir(fileSet, absPath, func(os.FileInfo) bool { return true }, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, pkg := range all {
		for _, f := range pkg.Files {
			ast.Inspect(f, func(x ast.Node) bool {
				typeSpec, ok := x.(*ast.TypeSpec)
				if !ok {
					return true
				}

				if _, ok := typeSpec.Type.(*ast.StructType); !ok {
					return true
				}

				// Check if the parsed type is from a scaffolded component with the name
				if compType, ok := typesToCheck[typeSpec.Name.Name]; ok {
					err = fmt.Errorf("零件 %s 名字 %s 已經創建 (類型 %s 存在)",
						compType,
						compName.Original,
						typeSpec.Name.Name,
					)
					return false
				}

				return true
			})
			if err != nil {
				return
			}
		}
	}
	return err
}

// checkForbiddenOracleFieldName如果名稱被禁止作為 oracle 字段名稱，則返回 true
func checkForbiddenOracleFieldName(name string) error {
	mfName, err := multiformatname.NewName(name, multiformatname.NoNumber)
	if err != nil {
		return err
	}

	// 檢查腳手架代碼中已使用的名稱
	switch mfName.UpperCase {
	case
		"CLIENTID",
		"ORACLESCRIPTID",
		"SOURCECHANNEL",
		"CALLDATA",
		"ASKCOUNT",
		"MINCOUNT",
		"FEELIMIT",
		"PREPAREGAS",
		"EXECUTEGAS":
		return fmt.Errorf("%s 由熊網鏈腳手架使用", name)
	}
	return nil
}

// checkCustomTypes 如果其中一種類型無效，則返回錯誤
func checkCustomTypes(ctx context.Context, path, module string, fields []string) error {
	protoPath := filepath.Join(path, protoFolder, module)
	customFields := make([]string, 0)
	for _, name := range fields {
		fieldSplit := strings.Split(name, datatype.Separator)
		if len(fieldSplit) <= 1 {
			continue
		}
		fieldType := datatype.Name(fieldSplit[1])
		if _, ok := datatype.SupportedTypes[fieldType]; !ok {
			customFields = append(customFields, string(fieldType))
		}
	}
	return protoanalysis.HasMessages(ctx, protoPath, customFields...)
}

// containCustomTypes 如果字段列表包含至少一種自定義類型，則返回 true
func containCustomTypes(fields []string) bool {
	for _, name := range fields {
		fieldSplit := strings.Split(name, datatype.Separator)
		if len(fieldSplit) <= 1 {
			continue
		}
		fieldType := datatype.Name(fieldSplit[1])
		if _, ok := datatype.SupportedTypes[fieldType]; !ok {
			return true
		}
	}
	return false
}
