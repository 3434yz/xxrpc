package buffer

import (
	"sync"
)

var bufPool = sync.Pool{
	New: func() any {
		// 初始化容量，比如 4KB
		b := make([]byte, 0, 1024)
		return &b
	},
}

func GetBuffer() *[]byte {
	return bufPool.Get().(*[]byte)
}

func PutBuffer(b *[]byte) {
	// 可选：防止过大的 buf 回池
	if cap(*b) > 64*1024 {
		return
	}
	*b = (*b)[:0] // 清空但不释放
	bufPool.Put(b)
}
