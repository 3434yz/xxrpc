package protocol

import (
	"encoding/binary"
	"io"
	"net"
)

// 读取一条完整的消息
func ReadFrame(conn net.Conn) ([]byte, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(conn, lenBuf[:]); err != nil {
		return nil, err
	}

	msgLen := binary.BigEndian.Uint32(lenBuf[:])
	if msgLen == 0 {
		return nil, nil
	}

	buf := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}
	return buf, nil
}
