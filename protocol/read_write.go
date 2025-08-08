package protocol

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
)

var bufPool = sync.Pool{
	New: func() any {
		// 默认分配 64KB，可以按你的 RPC 最大包大小调整
		b := make([]byte, 64*1024)
		return &b
	},
}

// PutFrameBuffer 调用方在用完 ReadFrame 返回的 buffer 后调用
func PutFrameBuffer(buf []byte) {
	bufPool.Put(&buf)
}

// ReadFrame 复用 buffer 读取
func ReadFrame(r io.Reader) ([]byte, error) {
	var lengthBuf [4]byte
	if _, err := io.ReadFull(r, lengthBuf[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf[:])

	// 从池里取出大缓冲区
	bufPtr := bufPool.Get().(*[]byte)
	buf := *bufPtr

	// 如果当前 buffer 太小，重新分配
	if uint32(cap(buf)) < length {
		buf = make([]byte, length)
	}

	// 按长度读取
	data := buf[:length]
	if _, err := io.ReadFull(r, data); err != nil {
		bufPool.Put(bufPtr) // 出错也要放回池
		return nil, err
	}

	return data, nil
}

// WriteFrame 使用 net.Buffers 一次性写入长度和数据
func WriteFrame(w io.Writer, data []byte) error {
	var lengthBuf [4]byte
	binary.BigEndian.PutUint32(lengthBuf[:], uint32(len(data)))

	if bw, ok := w.(net.Conn); ok {
		// 聚合两个切片一次性写
		var buff = net.Buffers{lengthBuf[:], data}
		_, err := buff.WriteTo(bw)
		return err
	}

	// 非 net.Conn（比如 bufio.Writer）
	if _, err := w.Write(lengthBuf[:]); err != nil {
		return err
	}
	_, err := w.Write(data)
	return err
}
