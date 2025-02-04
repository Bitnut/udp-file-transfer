package main

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/quic-go/quic-go"
)

func main() {

	// 检查命令行参数
	if len(os.Args) != 3 {
		fmt.Println("usage: client <文件路径> <服务端地址>")
		fmt.Println("example: client ./test.txt 127.0.0.1:4242")
		return
	}

	filePath := os.Args[1]
	serverAddr := os.Args[2]

	// 打开本地文件
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("open file error: ", err)
		panic(err)
	}
	defer file.Close()

	// 跳过证书验证（仅测试用）
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-file-transfer"},
	}

	// 连接到服务端
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := quic.DialAddr(ctx, serverAddr, tlsConf, &quic.Config{})
	if err != nil {
		fmt.Println("dial server error: ", err)
		panic(err)
	}
	defer conn.CloseWithError(0, "")

	// 创建 QUIC 流
	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		fmt.Println("open quic stream error: ", err)
		panic(err)
	}
	defer stream.Close()

	// 发送文件名
	fileName := filepath.Base(filePath)
	fileNameLen := int32(len(fileName))

	// 1. 先发送文件名长度（4字节大端序）
	if err := binary.Write(stream, binary.BigEndian, fileNameLen); err != nil {
		fmt.Println("send file name length error: ", err)
		panic(err)
	}

	// 2. 发送文件名内容
	if _, err := stream.Write([]byte(fileName)); err != nil {
		fmt.Println("send file name error: ", err)
		panic(err)
	}

	// 3. 发送文件内容
	if _, err := io.Copy(stream, file); err != nil {
		fmt.Println("send file content failed:", err)
		return
	}

	fmt.Println("send file success")
}
