package main

import (
	"fmt"
	"log"
	"net"
)

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

	for {

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
