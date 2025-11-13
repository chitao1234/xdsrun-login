// main.go
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
	"xdsrun/core"
)

// ================================================================================= //
//                                    主程序入口                                      //
// ================================================================================= //

func main() {
	// 定义命令行标志
	usernameFlag := flag.String("u", "", "您的账号 (学号)")
	passwordFlag := flag.String("p", "", "您的校园网密码")
	domainFlag := flag.String("d", "", "运营商后缀, 如 @dx, @lt, @yd (默认为校园网)")
	statusFlag := flag.Bool("s", false, "查询在线状态 (此模式下无需-u和-p)")
	flag.Parse()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 根据 -s 标志决定执行哪个流程
	if *statusFlag {
		core.CheckStatus(client)
	} else {
		// 检查登录模式下参数是否完整
		if *usernameFlag == "" || *passwordFlag == "" {
			fmt.Println("错误: 登录模式下必须提供 -u (账号) 和 -p (密码) 参数。")
			fmt.Println("用法示例: ./xdsrun -u 你的学号 -p '你的密码'")
			os.Exit(1)
		}
		core.PerformLogin(client, *usernameFlag, *passwordFlag, *domainFlag)
	}
}
