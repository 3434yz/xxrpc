package protocol

import (
	"encoding/binary"
	"net"
)

// 发送一条带长度前缀的消息
func WriteFrame(conn net.Conn, payload []byte) error {
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(payload)))

	// 先写长度，再写数据
	if _, err := conn.Write(lenBuf[:]); err != nil {
		return err
	}
	if _, err := conn.Write(payload); err != nil {
		return err
	}
	return nil
}
