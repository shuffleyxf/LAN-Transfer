package main

import (
	"fmt"
	"os"
	"transfer"
)

var (
	VERSION   = "0.0.1"                                                                              // 版本号
	HELP_INFO = "传参错误！程序使用方式如下：\n客户端模式: lan-transfer.exe -client \n服务器模式: lan-transfer.exe -server " // 帮助信息
)

// 打印信息并退出
func exit(msg string) {
	fmt.Println(msg)
	os.Exit(0)
}

// 主函数
func main() {
	args := os.Args

	argsLen := len(args)
	if argsLen != 2 {
		exit(HELP_INFO)
	}

	mode := args[1]
	switch mode {
	case "-client":
		transfer.StartClient()
	case "-server":
		transfer.StartServer()
	default:
		exit(HELP_INFO)
	}
}
