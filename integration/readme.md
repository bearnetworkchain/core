# Starport Integration Tests

Starport 集成測試構建一個新的應用程序並運行所有 Starport 命令來檢查 Starport 代碼的完整性。運行器和輔助方法位於此當前文件夾中。測試命令被分成文件夾，為了更好的並發性，每個文件夾都是 CI 工作流程中的並行作業。要創建一個新文件夾，我們只需要創建一個新文件夾。這將被自動檢測並添加到 PR CI 檢查中，或者我們只能在現有文件夾或文件中創建新測試。

Running synchronously all integration tests can be very slow. The command below can run everything:
```shell
go test -v -timeout 120m ./integration
```

Or you can just run a specific test folder, like the `list` types test
```shell
go test -v -timeout 120m ./integration/list
```

# Usage

- Create a new env and scaffold an empty chain:
```go
var (
    env  = envtest.New(t)
    path = env.Scaffold("github.com/test/blog")
)
```

- Now, you can use the env to run the starport commands and check the success status:
```go
env.Must(env.Exec("create a list with bool",
    step.NewSteps(step.New(
        step.Exec(envtest.IgniteApp, "s", "list", "--yes", "document", "signed:bool"),
        step.Workdir(path),
    )),
))
env.EnsureAppIsSteady(path)
```

- To check if the command returns an error, you can add the `envtest.ExecShouldError()` step:
```go
env.Must(env.Exec("should prevent creating a list with duplicated fields",
    step.NewSteps(step.New(
        step.Exec(envtest.IgniteApp, "s", "list", "--yes", "company", "name", "name"),
        step.Workdir(path),
    )),
    envtest.ExecShouldError(),
))
env.EnsureAppIsSteady(path)
```
