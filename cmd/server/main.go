package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

// closing 用于标记是否正在关闭连接
var closing int32

func main() {

	// 检查地址，并得到需要监听的 UDP 的地址信息
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:9999")
	if err != nil {
		log.Fatalf("Error resolving UDP address: %v", err)
	}

	// 打印解析后的地址
	fmt.Println("Resolved address: ", udpAddr)

	// 监听 UDP 地址
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Error listening on UDP address: %v", err)
	}
	// 退出程序前，关闭连接
	defer conn.Close()

	fmt.Println("Listening on: ", udpAddr)

	handleSignals(conn)

	handleUDPConnection(conn)
}

func handleUDPConnection(conn *net.UDPConn) {

	for {

		if atomic.LoadInt32(&closing) == 1 {
			break
		}

		buf := make([]byte, 1024) // 创建一个 1024 字节的缓冲区用于接收数据

		// 读取数据
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading from UDP: ", err)
			continue
		}

		// 打印接收到的数据
		fmt.Printf("Received %s from %s\n", buf[:n], addr)
	}

}

func handleSignals(conn *net.UDPConn) {

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
