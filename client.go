package transfer

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	c_BUFFER_SIZE = 1024 * 1024 // 数据缓冲区大小
)

// 文件传输
func transferFile(conn net.Conn, buffer []byte, args []string) {
	filePath := args[1]
	fmt.Printf("正在传输文件：%s\n", filePath)
	startTime := time.Now()
	err := SendFile(conn, buffer, filePath)
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	if err == nil {
		fmt.Printf("文件%s传输完成, 耗时%v \n", filePath, elapsedTime)
	} else {
		fmt.Printf("文件%s传输异常：%s, 连接断开", filePath, err)
		os.Exit(0)
	}
}

// 拉取文件
func fetchFile(conn net.Conn, buffer []byte, args []string) {
	remoteFilePath := args[1]
	saveDir := args[2]
	startTime := time.Now()
	filePath, err := FetchFile(conn, buffer, remoteFilePath, saveDir)
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	if err == nil {
		fmt.Printf("文件%s拉取完成, 耗时%v \n", filePath, elapsedTime)
	} else {
		fmt.Printf("文件%s拉取异常：%s\n", remoteFilePath, err)
	}
}

// 发送消息
func sendMsg(conn net.Conn, args []string) {
	content := strings.Join(args[1:], " ")
	err := WriteMsg(conn, content)
	if err == nil {
		fmt.Printf("消息发送成功！\n")
	} else {
		fmt.Printf("发送消息出现异常：%v\n", err)
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
	fmt.Println("连接成功！目前支持以下指令\n文件传输： transfer-file {本地文件路径}\n文件拉取：fetch-file {远程文件路径} {本地存储路径}\n接收消息：send-msg {消息内容}")

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
		case "fetch-file":
			fetchFile(conn, buffer, items)
		case "send-msg":
			sendMsg(conn, items)
		case "bye":
			fmt.Println("Goodbye!")
			os.Exit(0)
		default:
			fmt.Printf("未知指令: %s\n", order)
		}
	}
}
