package tsrelayer

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/gorilla/rpc/v2/json2"
	"golang.org/x/sync/errgroup"

	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/nodetime"
)

//Call 使用 args 調用 ts 中繼器包裝器庫中的方法，並從返回的值中填充回复。
func Call(ctx context.Context, method string, args, reply interface{}) error {
	command, cleanup, err := nodetime.Command(nodetime.CommandXRelayer)
	if err != nil {
		return err
	}
	defer cleanup()

	req, err := json2.EncodeClientRequest(method, args)
	if err != nil {
		return err
	}

	resr, resw := io.Pipe()

	g := errgroup.Group{}
	g.Go(func() error {
		defer resw.Close()

		return cmdrunner.New().Run(
			ctx,
			step.New(
				step.Exec(command[0], command[1:]...),
				step.Write(req),
				step.Stdout(resw),
			),
		)
	})

	// 在發出 jsonrpc 響應之前，其他進程可以將常規日誌打印到標準輸出。
	// 區分兩種類型並模擬打印常規日誌（如果有）。
	sc := bufio.NewScanner(resr)
	for sc.Scan() {
		err = json2.DecodeClientResponse(bytes.NewReader(sc.Bytes()), reply)

		var e *json2.Error
		if errors.As(err, &e) || errors.Is(err, json2.ErrNullResult) { // jsonrpc 返回一個服務器端錯誤。
			return err
		}

		if err != nil { // 由另一個進程打印到標準輸出的一行。
			fmt.Println(sc.Text())
		}
	}

	if err := sc.Err(); err != nil {
		return err
	}

	return g.Wait()
}
