package internal

type FileInfo struct {
	FileID      uint32
	FileName    string
	FileSize    uint64
	TotalChunks uint32
	Hash        []byte // 文件哈希值
}

// Packet 代表一个数据包
type Packet struct {
	FileId      uint32
	ChunkId     uint32
	TotalChunks uint32
	Checksum    uint32
	Data        []byte
}

// 控制报文类型
// ACK 代表确认报文
type ACK struct {
	FileID  uint32
	ChunkID uint32
	Status  byte // 0: ACK, 1: NACK
}
