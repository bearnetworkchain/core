package xexec

import "os/exec"

//IsCommandAvailable 檢查用戶路徑上的命令是否可用。
func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
