package availableport

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

// Find 查找 n 個未使用的端口。
// 不保證這些端口不會被分配到
// 調用 Find() 時的另一個程序。
func Find(n int) (ports []int, err error) {
	min := 44000
	max := 55000

	for i := 0; i < n; i++ {
		for {
			rand.Seed(time.Now().UnixNano())
			port := rand.Intn(max-min+1) + min

			conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
			// 如果有錯誤，這可能意味著沒有人在從這個端口監聽
			// 這就是我們需要的。
			if err == nil {
				conn.Close()
				continue
			}
			ports = append(ports, port)
			break
		}
	}
	return ports, nil
}
