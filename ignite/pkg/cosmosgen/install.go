package cosmosgen

import (
	"bytes"
	"context"

	"github.com/pkg/errors"

	"github.com/bearnetworkchain/core/ignite/pkg/cmdrunner"
	"github.com/bearnetworkchain/core/ignite/pkg/cmdrunner/step"
)

//InstallDependencies 安裝 Cosmos 生態系統所需的 protoc 依賴項。
func InstallDependencies(ctx context.Context, appPath string) error {
	plugins := []string{
		// 安裝 gocosmos 插件。
		"github.com/regen-network/cosmos-proto/protoc-gen-gocosmos",

		// 安裝 Go 代碼生成插件。
		"github.com/golang/protobuf/protoc-gen-go",

		// 安裝 grpc-gateway 插件。
		"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway",
		"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger",
		"github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2",
	}

	errb := &bytes.Buffer{}
	err := cmdrunner.
		New(
			cmdrunner.DefaultStderr(errb),
			cmdrunner.DefaultWorkdir(appPath),
		).
		Run(ctx,
			step.New(step.Exec("go", append([]string{"get"}, plugins...)...)),
			step.New(step.Exec("go", append([]string{"install"}, plugins...)...)),
		)
	return errors.Wrap(err, errb.String())
}
