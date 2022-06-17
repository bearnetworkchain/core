package xgenny

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/logger"
	"github.com/gobuffalo/packd"

	"github.com/bearnetworkchain/core/ignite/pkg/placeholder"
	"github.com/bearnetworkchain/core/ignite/pkg/validation"
)

var _ validation.Error = (*dryRunError)(nil)

type dryRunError struct {
	error
}

//ValidationInfo 返回驗證信息
func (d *dryRunError) ValidationInfo() string {
	return d.Error()
}

//DryRunner 是一個帶有記錄器的 genny DryRunner
func DryRunner(ctx context.Context) *genny.Runner {
	runner := genny.DryRunner(ctx)
	runner.Logger = logger.New(genny.DefaultLogLvl)
	return runner
}

// RunWithValidation 用乾運行檢查生成器，然後對生成器執行濕流道
func RunWithValidation(
	tracer *placeholder.Tracer,
	gens ...*genny.Generator,
) (sm SourceModification, err error) {
	// run 使用提供的生成器執行提供的運行器
	run := func(runner *genny.Runner, gen *genny.Generator) error {
		err := runner.With(gen)
		if err != nil {
			return err
		}
		return runner.Run()
	}
	for _, gen := range gens {
		// 用乾流道檢查發電機
		dryRunner := DryRunner(context.Background())
		if err := run(dryRunner, gen); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return sm, &dryRunError{err}
			}
			return sm, err
		}
		if err := tracer.Err(); err != nil {
			return sm, err
		}

		// 獲取源修改
		sm = NewSourceModification()
		for _, file := range dryRunner.Results().Files {
			fileName := file.Name()
			_, err := os.Stat(fileName)

			//不願意：gocritic
			if os.IsNotExist(err) {
				// 如果該文件在源中不存在，則表示它已由運行程序創建
				sm.AppendCreatedFiles(fileName)
			} else if err != nil {
				return sm, err
			} else {
				// 該文件已被跑步者修改
				sm.AppendModifiedFiles(fileName)
			}
		}

		// 使用濕流道執行修改
		if err := run(genny.WetRunner(context.Background()), gen); err != nil {
			return sm, err
		}
	}
	return sm, nil
}

// Box 將掛載 Box 中的每個文件並進行包裝，已經存在的文件將被忽略
func Box(g *genny.Generator, box packd.Walker) error {
	return box.Walk(func(path string, bf packd.File) error {
		f := genny.NewFile(path, bf)
		f, err := g.Transform(f)
		if err != nil {
			return err
		}
		filePath := strings.TrimSuffix(f.Name(), ".plush")
		_, err = os.Stat(filePath)
		if os.IsNotExist(err) {
			// path doesn't exist. move on.
			g.File(f)
			return nil
		}
		return err
	})
}
