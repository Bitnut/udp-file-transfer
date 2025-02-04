package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	mrand "math/rand"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/quic-go/quic-go"
)

// closing 用于标记是否正在关闭连接
var closing int32

const (
	saveDir = "./uploads"
)

func main() {

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		panic(err)
	}

	// 检查地址，并得到需要监听的 UDP 的地址信息
	ln, err := quic.ListenAddr(
		"127.0.0.1:4222",
		generateTLSConfig(),
		&quic.Config{},
	)
	if err != nil {
		panic(err)
	}

	// 退出程序前，关闭连接
	defer ln.Close()

	fmt.Println("Listening on: localhost", 4222)

	handleSignals(ln)

	handleUDPConnection(ln)
}

func handleUDPConnection(ln *quic.Listener) {

	for {

		// 如果正在关闭连接，则退出循环
		if atomic.LoadInt32(&closing) == 1 {
			fmt.Println("Closing connection, exiting...")
			break
		}

		conn, err := ln.Accept(context.Background())
		if err != nil {
			fmt.Println("Accept connection err:", err)
			continue
		}

		// 处理连接
		go func(conn quic.Connection) {
			defer conn.CloseWithError(0, "")

			// 接受数据流
			stream, err := conn.AcceptStream(context.Background())
			if err != nil {
				fmt.Println("accept stream err:", err)
				return
			}

			if err := handleStream(&stream); err != nil {
				fmt.Println("handle stream err:", err)
			}

			stream.Close()

		}(conn)

	}

}

func handleStream(stream *quic.Stream) error {

	// 读取文件名（客户端先发送文件名长度）
	var nameLen int32
	if err := binary.Read(*stream, binary.BigEndian, &nameLen); err != nil {
		fmt.Println("Read FileNameLen Error:", err)
		return err
	}

	fileNameBuf := make([]byte, nameLen)
	if _, err := io.ReadFull(*stream, fileNameBuf); err != nil {
		fmt.Println("Read FileName Error:", err)
		return err
	}

	fileName := createRandFilename(string(fileNameBuf))

	// 创建保存目录
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		fmt.Println("Create Folder Error:", err)
		return err
	}

	// 创建目标文件
	filePath := filepath.Join(saveDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Create File Error:", err)
		return err
	}
	defer file.Close()

	// 接收文件数据并写入磁盘
	if _, err := io.Copy(file, *stream); err != nil {
		fmt.Println("Write File Error:", err)
		return err
	}

	fmt.Printf("Receive File Success %s", fileName)

	return nil
}

func handleSignals(conn *quic.Listener) {

	// 捕获 SIGINT/SIGTERM 信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("Received interrupt signal, closing connection...")
		atomic.StoreInt32(&closing, 1)
		conn.Close()
		os.Exit(0)
	}()
}

// 生成自签名TLS证书（仅供测试使用）
func generateTLSConfig() *tls.Config {

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}
	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&key.PublicKey,
		key,
	)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-file-transfer"},
	}
}

func createRandFilename(fileName string) string {

	return fmt.Sprintf("%s-%d-%d", fileName, time.Now().Unix(), mrand.Intn(1000))
}
