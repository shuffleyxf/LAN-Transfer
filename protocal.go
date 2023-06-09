package transfer

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	PACKET_TYPE_LEN = 1 // 数据包类型字段长度所占字节数

	FILE_TRANSFER = 1 // 文件传输类型
	FILE_FETCH    = 2 // 文件拉取类型
	FILE_NAME_LEN = 2 // 文件名长度所占字节数
	FILE_LEN      = 8 // 文件长度所占字节数

	MSG     = 3 //消息类型
	MSG_LEN = 2 //消息长度所占字节数
)

// ReadUInt 从TCP连接读取指定长度的无符号整数
func ReadUInt(conn net.Conn, length uint8) (uint64, error) {
	data := make([]byte, length)
	_, err := conn.Read(data)
	if err != nil {
		return 0, err
	}
	if length == 1 {
		return uint64(data[0]), nil
	} else if length == 2 {
		return uint64(binary.LittleEndian.Uint16(data)), nil
	} else if length == 4 {
		return uint64(binary.LittleEndian.Uint32(data)), nil
	} else if length == 8 {
		return binary.LittleEndian.Uint64(data), nil
	} else {
		return 0, fmt.Errorf("非法字段长度：%d", length)
	}
}

// WriteUInt 向TCP连接写入指定长度的无符号整数
func WriteUInt(conn net.Conn, val uint64, length uint8) error {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, val)
	_, err := conn.Write(data[:length])
	return err
}

// ReadStringUTF8 从TCP连接读取utf8字符串
func ReadStringUTF8(conn net.Conn, length uint32) (string, error) {
	data := make([]byte, length)
	_, err := conn.Read(data)
	if err == nil {
		return string(data), nil
	} else {
		return "", err
	}
}

// WriteStringUTF8 向连接写入utf8字符串
func WriteStringUTF8(conn net.Conn, content string) error {
	data := []byte(content)
	_, err := conn.Write(data)
	return err
}

// SendFile 通过TCP连接发送文件
func SendFile(conn net.Conn, buffer []byte, filePath string) error {
	// 先读取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("读取文件信息失败：%v", err)
	}

	// 写入包类型 1byte
	err = WriteUInt(conn, FILE_TRANSFER, PACKET_TYPE_LEN)
	if err != nil {
		return fmt.Errorf("包类型写入失败: %v", err)
	}

	// 写入文件名大小 2byte
	filePath = strings.Replace(filePath, "\\", "/", -1)
	fileName := path.Base(filePath)
	nameLength := len(fileName)
	err = WriteUInt(conn, uint64(nameLength), FILE_NAME_LEN)
	if err != nil {
		return fmt.Errorf("文件名长度写入失败: %v", err)
	}

	// 写入文件名
	err = WriteStringUTF8(conn, fileName)
	if err != nil {
		return fmt.Errorf("文件名写入失败: %v", err)
	}

	// 写入文件大小
	fileSize := fileInfo.Size()
	if err != nil {
		return fmt.Errorf("文件长度读取失败: %v", err)
	}
	err = WriteUInt(conn, uint64(fileSize), FILE_LEN)
	if err != nil {
		return fmt.Errorf("文件长度写入失败: %v", err)
	}

	// 写入文件流
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		// 处理读取的字节块
		_, err = conn.Write(buffer[:n])
		if err != nil {
			return fmt.Errorf("将文件流写入文件失败: %v", err)
		}
	}

	return nil
}

// ReceiveFile 通过TCP连接接收文件
func ReceiveFile(conn net.Conn, buffer []byte, saveDir string) (string, error) {
	// 读取文件名长度
	nameLength, err := ReadUInt(conn, FILE_NAME_LEN)
	if err != nil {
		return "", fmt.Errorf("文件名长度读取失败: %v", err)
	}

	// 读取文件名
	fileName, err := ReadStringUTF8(conn, uint32(nameLength))
	if err != nil {
		return "", fmt.Errorf("文件名读取失败: %v", err)
	}

	// 读取文件长度
	fileLength, err := ReadUInt(conn, FILE_LEN)
	fileLengthInt := int(fileLength)
	if err != nil {
		return "", fmt.Errorf("文件长度读取失败: %v", err)
	}

	// 读取文件流并写入本地
	filePath := filepath.Join(saveDir, fileName)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("文件%s创建失败: %v", filePath, err)
	}
	defer file.Close()

	for {
		n, err := conn.Read(buffer)
		fileLengthInt -= n
		if err != nil {
			return "", fmt.Errorf("从连接读取文件流失败: %v", err)
		}
		if n == 0 {
			continue
		}

		_, err = file.Write(buffer[:n])
		if err != nil {
			return "", fmt.Errorf("将文件流写入文件失败: %v", err)
		}

		if fileLengthInt == 0 {
			break
		}
	}
	return filePath, nil
}

// FetchFile 通过TCP连接拉取文件
func FetchFile(conn net.Conn, buffer []byte, filePath, saveDir string) (string, error) {
	// 写入包类型 1byte
	err := WriteUInt(conn, FILE_FETCH, PACKET_TYPE_LEN)
	if err != nil {
		return "", fmt.Errorf("包类型写入失败: %v", err)
	}

	// 写入文件名大小 2byte
	filePath = strings.Replace(filePath, "\\", "/", -1)
	fileName := path.Base(filePath)
	nameLength := len(fileName)
	err = WriteUInt(conn, uint64(nameLength), FILE_NAME_LEN)
	if err != nil {
		return "", fmt.Errorf("文件名长度写入失败: %v", err)
	}

	// 写入文件名
	err = WriteStringUTF8(conn, fileName)
	if err != nil {
		return "", fmt.Errorf("文件名写入失败: %v", err)
	}

	packetType, err := ReadUInt(conn, 1)
	if err != nil {
		return "", fmt.Errorf("文件拉取失败，读取文件类型异常: %v", err)
	}
	if packetType == FILE_TRANSFER {
		// 拉取文件流
		return ReceiveFile(conn, buffer, saveDir)
	} else if packetType == MSG {
		errorMsg, err := ReadMsg(conn)
		if err == nil {
			return "", fmt.Errorf("文件拉取失败: %v", errorMsg)
		} else {
			return "", fmt.Errorf("文件拉取失败，读取错误消息异常: %v", err)
		}
	} else {
		return "", fmt.Errorf("文件拉取失败，错误的文件类型: %v", packetType)
	}
}

// ParseFileRequest 解析文件请求
func ParseFileRequest(conn net.Conn) (string, error) {
	// 读取文件名长度
	nameLength, err := ReadUInt(conn, FILE_NAME_LEN)
	if err != nil {
		return "", fmt.Errorf("文件名长度读取失败: %v", err)
	}

	// 读取文件名
	filePath, err := ReadStringUTF8(conn, uint32(nameLength))
	if err != nil {
		return "", fmt.Errorf("文件名读取失败: %v", err)
	}
	return filePath, nil
}

// WriteMsg 发送消息
func WriteMsg(conn net.Conn, content string) error {
	err := WriteUInt(conn, MSG, PACKET_TYPE_LEN)
	if err != nil {
		return fmt.Errorf("包类型写入失败: %v", err)
	}

	contentLength := len(content)
	err = WriteUInt(conn, uint64(contentLength), MSG_LEN)
	if err != nil {
		return fmt.Errorf("消息长度写入失败: %v", err)
	}

	err = WriteStringUTF8(conn, content)
	if err != nil {
		return fmt.Errorf("消息内容写入失败: %v", err)
	}

	return nil
}

// ReadMsg 接收消息
func ReadMsg(conn net.Conn) (string, error) {
	contentLength, err := ReadUInt(conn, MSG_LEN)
	if err != nil {
		return "", fmt.Errorf("消息长度读取失败: %v", err)
	}

	content, err := ReadStringUTF8(conn, uint32(contentLength))
	if err != nil {
		return "", fmt.Errorf("消息读取失败: %v", err)
	}

	return content, nil
}
