package protocol

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
)

// 连接内部缓冲池（基于你之前的 bufPool 思路）
var frameBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 64*1024) // 默认 64KB，可按需调大
		return &b
	},
}

var ErrFrameTooLarge = errors.New("frame too large")

// FrameConn wraps net.Conn and provides framed Read/Write with buffer reuse.
// IMPORTANT: Frame slices returned by ReadFrame reference internal buffer.
// Caller MUST NOT hold onto the slice after next ReadFrame call (or Close()).
type FrameConn struct {
	conn  net.Conn
	buf   *[]byte // from frameBufPool
	start int     // start index of unread data in buf
	end   int     // end index (exclusive) of data in buf
	// optional max frame size to protect memory blowup
	MaxFrameSize int
}

// NewFrameConn creates a FrameConn. Caller must call Close() to release resources.
func NewFrameConn(c net.Conn) *FrameConn {
	bp := frameBufPool.Get().(*[]byte)
	return &FrameConn{
		conn:         c,
		buf:          bp,
		start:        0,
		end:          0,
		MaxFrameSize: 4 * 1024 * 1024, // default 4MB limit
	}
}

// Close returns the buffer to pool and closes underlying conn.
func (fc *FrameConn) Close() error {
	// return buffer
	if fc.buf != nil {
		frameBufPool.Put(fc.buf)
		fc.buf = nil
	}
	if fc.conn != nil {
		err := fc.conn.Close()
		fc.conn = nil
		return err
	}
	return nil
}

func (fc *FrameConn) ensureAvailable(n int) (bool, error) {
	if fc.end-fc.start >= n {
		return true, nil
	}

	// 如果 start 和 end 非零，并且未满，则将剩余的字节复制到缓冲区的开头
	if fc.start > 0 && fc.start != fc.end {
		copy((*fc.buf)[:fc.end-fc.start], (*fc.buf)[fc.start:fc.end])
		fc.end = fc.end - fc.start
		fc.start = 0
	} else if fc.start == fc.end {
		fc.start = 0
		fc.end = 0
	}

	// 如果缓冲区容量不够，进行一次读取操作
	capacity := cap(*fc.buf)
	if fc.end+1 > capacity {
		if n > capacity {
			return false, nil
		}
	}

	readInto := (*fc.buf)[fc.end:capacity]
	nr, err := fc.conn.Read(readInto)
	if nr > 0 {
		fc.end += nr
	}
	if err != nil {
		if err == io.EOF && fc.end-fc.start >= n {
			return true, nil
		}
		return false, err
	}

	return fc.end-fc.start >= n, nil
}

// ReadFrame returns a single full frame (payload bytes, without length prefix).
// The returned slice references an internal buffer; do NOT retain it.
func (fc *FrameConn) ReadFrame() ([]byte, error) {
	// First ensure 4 bytes for length
	ok, err := fc.ensureAvailable(4)
	if err != nil {
		return nil, err
	}
	if !ok {
		// Not enough data after single read: try blocking read with io.ReadFull as fallback
		var lenBuf [4]byte
		// If there's partial, copy what we have then read remaining
		avail := fc.end - fc.start
		if avail > 0 {
			copy(lenBuf[:], (*fc.buf)[fc.start:fc.end])
			//want := 4 - avail
			if _, err := io.ReadFull(fc.conn, lenBuf[avail:]); err != nil {
				return nil, err
			}
			// update buf state: we consumed these avail bytes
			fc.start += avail
		} else {
			if _, err := io.ReadFull(fc.conn, lenBuf[:]); err != nil {
				return nil, err
			}
		}
		length := binary.BigEndian.Uint32(lenBuf[:])
		if int(length) > fc.MaxFrameSize {
			return nil, ErrFrameTooLarge
		}
		// now read payload
		// try to use buffer if fits
		capacity := cap(*fc.buf)
		if int(length) <= capacity {
			// ensure we have length bytes in buffer
			// read into buffer directly
			// reset start/end
			fc.start = 0
			fc.end = 0
			if _, err := io.ReadFull(fc.conn, (*fc.buf)[:length]); err != nil {
				return nil, err
			}
			fc.end = int(length)
			return (*fc.buf)[:length], nil
		}
		// else allocate exact slice (not pooled)
		out := make([]byte, length)
		if _, err := io.ReadFull(fc.conn, out); err != nil {
			return nil, err
		}
		return out, nil
	}

	// We have at least 4 bytes in buffer
	lenBytes := (*fc.buf)[fc.start : fc.start+4]
	frameLen := binary.BigEndian.Uint32(lenBytes)
	if int(frameLen) > fc.MaxFrameSize {
		return nil, ErrFrameTooLarge
	}

	// Ensure payload bytes are available in buffer
	totalNeed := 4 + int(frameLen)
	if fc.end-fc.start < totalNeed {
		// try reading more once
		ok2, err := fc.ensureAvailable(totalNeed)
		if err != nil {
			return nil, err
		}
		if !ok2 {
			// fallback to io.ReadFull to fill remaining bytes
			//needed := totalNeed - (fc.end - fc.start)
			// if payload too big to fit in buffer, read the remainder into a new slice:
			if int(frameLen) > cap(*fc.buf) {
				// copy what we have for the header (we'll discard), then read remainder into fresh slice
				// advance start to consume header
				fc.start += 4
				out := make([]byte, frameLen)
				// first copy any partial payload present
				present := fc.end - fc.start
				if present > 0 {
					copy(out[:present], (*fc.buf)[fc.start:fc.end])
				}
				if _, err := io.ReadFull(fc.conn, out[present:]); err != nil {
					return nil, err
				}
				// after consuming, reset buffer
				fc.start = 0
				fc.end = 0
				return out, nil
			}
			// otherwise try to fill buffer
			if _, err := io.ReadFull(fc.conn, (*fc.buf)[fc.end:fc.start+totalNeed]); err != nil {
				return nil, err
			}
			fc.end = fc.start + totalNeed
		}
	}

	// Now we definitely have a full frame in buffer
	payloadStart := fc.start + 4
	payloadEnd := payloadStart + int(frameLen)
	payload := (*fc.buf)[payloadStart:payloadEnd]

	// advance start
	fc.start = payloadEnd
	// if buffer consumed entirely, reset indices
	if fc.start == fc.end {
		fc.start = 0
		fc.end = 0
	}
	return payload, nil
}

// WriteFrame writes a payload with 4-byte length prefix using net.Buffers to reduce syscalls.
func (fc *FrameConn) WriteFrame(payload []byte) error {
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(payload)))

	buf := net.Buffers{lenBuf[:], payload}
	_, err := buf.WriteTo(fc.conn)
	return err
}
