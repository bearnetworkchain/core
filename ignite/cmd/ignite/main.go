package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	ignitecmd "github.com/bearnetworkchain/core/ignite/cmd"
	"github.com/bearnetworkchain/core/pkg/clictx"
	"github.com/bearnetworkchain/core/pkg/validation"
)

func main() {
	ctx := clictx.From(context.Background())

	err := ignitecmd.New().ExecuteContext(ctx)

	if ctx.Err() == context.Canceled || err == context.Canceled {
		fmt.Println("中止")
		return
	}

	if err != nil {
		var validationErr validation.Error

		if errors.As(err, &validationErr) {
			fmt.Println(validationErr.ValidationInfo())
		} else {
			fmt.Println(err)
		}

		os.Exit(1)
	}
}
