package transfer

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	c_BUFFER_SIZE = 1024 * 1024 // 数据缓冲区大小
)

// 文件传输
func transferFile(conn net.Conn, buffer []byte, args []string) {
	filePath := args[1]
	fmt.Printf("正在传输文件：%s\n", filePath)
	err := SendFile(conn, buffer, filePath)
	if err == nil {
		fmt.Printf("文件%s传输完成\n", filePath)
	} else {
		fmt.Printf("文件%s传输异常：%s, 连接断开", filePath, err)
		os.Exit(0)
	}
}

// StartClient 客户端模式启动
func StartClient() {
	scanner := bufio.NewScanner(os.Stdin)
	var conn net.Conn
	for {
		var err error
		address := Input(scanner, "请输入目标服务器地址：")
		conn, err = net.Dial("tcp", address)
		if err == nil {
			break
		}
		fmt.Printf("建立连接失败：%s\n", address)
	}
	fmt.Println("连接成功！目前支持以下指令\n文件传输： transfer-file {本地文件路径}")

	var order string
	buffer := make([]byte, c_BUFFER_SIZE)
	for {
		content := Input(scanner, "请输入指令:")
		items := strings.Split(content, " ")
		if len(items) == 0 {
			continue
		}

		order = items[0]
		switch order {
		case "transfer-file":
			transferFile(conn, buffer, items)
		case "bye":
			fmt.Println("Goodbye!")
			os.Exit(0)
		default:
			fmt.Printf("未知指令: %s\n", order)
		}
	}
}
