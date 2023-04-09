package transfer

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	s_BUFFER_SIZE = 1024 * 1024 //数据缓冲区大小
)

// 返回文件存储路径
func getSaveDir() string {
	curDir, err := os.Getwd()
	if err != nil {
		fmt.Println("获取当前路径失败")
		return ""
	}

	saveDir := filepath.Join(curDir, "data")
	_, err = os.Stat(saveDir)
	if os.IsNotExist(err) {
		// 创建文件夹
		err := os.Mkdir(saveDir, 0755)
		if err != nil {
			fmt.Printf("文件夹创建失败：%v\n", err)
		}
	} else if err != nil {
		fmt.Printf("读取路径状态异常：%v\n", err)
	}
	return saveDir
}

// 接收文件
func acceptFile(conn net.Conn, buffer []byte) {
	saveDir := getSaveDir()
	fmt.Printf("接收文件中，存储路径：%s\n", saveDir)
	startTime := time.Now()
	filePath, err := ReceiveFile(conn, buffer, saveDir)
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	if err == nil {
		fmt.Printf("文件写入成功：%s, 耗时%v \n", filePath, elapsedTime)
	} else {
		fmt.Printf("读取文件到本地失败: %s\n", err)
	}

}

// 处理客户端连接
func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Printf("新的客户端连接： %s\n", conn.RemoteAddr().String())
	typeByte := make([]byte, 1)
	buffer := make([]byte, s_BUFFER_SIZE)
	// 在这里可以读写客户端的数据
	for {
		length, err := conn.Read(typeByte)
		if err != nil {
			fmt.Printf("接受数据出现异常(%s)：%s, 连接将断开\n", conn.RemoteAddr(), err)
			break
		}
		if length == 0 {
			time.Sleep(1 * time.Second)
		}

		packetType := int(typeByte[0])
		switch packetType {
		case FILE_TRANSFER:
			acceptFile(conn, buffer)
		default:
			fmt.Printf("未知的消息类型：%d, 连接(%s)将断开\n", packetType, conn.RemoteAddr().String())
			return
		}
	}

}

// 启动tcp服务器
func start(port int) {
	address := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("创建服务器出现异常:", err)
		return
	}
	defer listener.Close()

	// 循环等待客户端连接
	fmt.Println("正在等待客户端连接...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("建立连接出现异常:", err)
			continue
		}
		go handleConnection(conn)
	}
}

// StartServer 服务器模式启动
func StartServer() {
	scanner := bufio.NewScanner(os.Stdin)
	var port int
	var err error
	for {
		portStr := Input(scanner, "请输入监听端口：")
		port, err = strconv.Atoi(portStr)
		if err == nil && port >= 1 && port <= 65536 {
			break
		}
		fmt.Printf("非法的端口：%s", portStr)
	}
	start(port)
}
