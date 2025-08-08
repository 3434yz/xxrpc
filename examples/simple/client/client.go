package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"xxrpc/client"
	"xxrpc/examples/simple/echo"
)

// 配置参数
const (
	maxGoroutines      = 100           // 最大goroutine数量限制
	runtimeMinSec      = 300           // 最小运行时间(秒)
	connectionPoolSize = maxGoroutines // 连接池大小，根据服务器承载能力调整
	statsInterval      = 5             // 统计信息打印间隔(秒)
)

// 连接池结构体，管理可复用的RPC连接
type ConnectionPool struct {
	pool chan *client.Client
}

// 创建新的连接池
func NewConnectionPool(size int) (*ConnectionPool, error) {
	pool := make(chan *client.Client, size)

	// 初始化连接池
	for i := 0; i < size; i++ {
		cli, err := client.Dial(":8888")
		if err != nil {
			// 关闭已创建的连接
			closeConnections(pool)
			return nil, fmt.Errorf("创建连接失败: %v", err)
		}
		pool <- cli
	}

	return &ConnectionPool{
		pool: pool,
	}, nil
}

// 从连接池获取连接
func (p *ConnectionPool) Get() *client.Client {
	return <-p.pool
}

// 将连接放回连接池
func (p *ConnectionPool) Put(cli *client.Client) {
	select {
	case p.pool <- cli:
		// 成功放回连接池
	default:
		// 连接池已满，关闭多余连接
		cli.Close()
	}
}

// 关闭所有连接
func (p *ConnectionPool) Close() {
	close(p.pool)
	for cli := range p.pool {
		cli.Close()
	}
}

// 关闭通道中的所有连接
func closeConnections(pool chan *client.Client) {
	close(pool)
	for cli := range pool {
		cli.Close()
	}
}

// 使用复用的连接执行RPC调用
func RPCCall(pool *ConnectionPool) error {
	// 从连接池获取连接
	cli := pool.Get()
	defer pool.Put(cli) // 调用完成后放回连接池

	req := echo.ComplexHelloReq{
		Message:   "Hello",
		ID:        12345,
		Timestamp: time.Now(),
		Metadata:  map[string]string{"key": "value"},
		Tags:      []string{"go", "rpc", "test"},
		Nested: struct {
			Name  string
			Score float64
		}{
			Name:  "nested",
			Score: 9.8,
		},
		Data:       []byte("payload"),
		Attributes: [5]int{1, 2, 3, 4, 5},
		Enabled:    true,
		Options: &struct {
			Retry   int
			Timeout time.Duration
		}{
			Retry:   3,
			Timeout: time.Second,
		},
	}

	// 执行RPC调用
	_, err := cli.Call("EchoService.ComplexHello", req)
	return err
}

func main() {
	// 初始化连接池
	pool, err := NewConnectionPool(connectionPoolSize)
	if err != nil {
		log.Fatalf("初始化连接池失败: %v", err)
	}
	defer pool.Close()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxGoroutines) // 控制并发数量的信号量
	startTime := time.Now()
	taskCount := 0
	successCount := 0
	errorCount := 0
	var mu sync.Mutex // 用于安全更新统计数据

	// 启动统计信息打印goroutine
	statsTicker := time.NewTicker(time.Duration(statsInterval) * time.Second)
	defer statsTicker.Stop()

	go func() {
		for range statsTicker.C {
			mu.Lock()
			elapsed := time.Since(startTime)
			qps := 0.0
			if elapsed.Seconds() > 0 {
				qps = float64(successCount) / elapsed.Seconds()
			}

			fmt.Printf("\n===== 运行统计 =====")
			fmt.Printf("\n当前运行时间: %v", elapsed)
			fmt.Printf("\n总调用次数: %d", taskCount)
			fmt.Printf("\n成功次数: %d", successCount)
			fmt.Printf("\n错误次数: %d", errorCount)
			fmt.Printf("\n平均QPS: %.2f", qps)
			fmt.Printf("\n====================\n\n")
			mu.Unlock()
		}
	}()

	// 持续运行直到达到最小运行时间
	for time.Duration(time.Since(startTime).Seconds()) < time.Duration(runtimeMinSec)*time.Second {
		// 控制任务生成速度，避免过早填满信号量
		semaphore <- struct{}{}
		wg.Add(1)
		taskCount++

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			// 执行RPC调用
			err := RPCCall(pool)

			// 更新统计数据
			mu.Lock()
			if err != nil {
				errorCount++
				log.Printf("rpc call error: %v", err)
			} else {
				successCount++
			}
			mu.Unlock()
		}()

		// 控制任务生成速率，避免瞬间创建过多goroutine
		if taskCount%100 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// 等待所有剩余任务完成
	wg.Wait()
	duration := time.Since(startTime)

	fmt.Printf("\n\n所有RPC调用完成。\n")
	fmt.Printf("总运行时间: %v\n", duration)
	fmt.Printf("总调用次数: %d\n", taskCount)
	fmt.Printf("成功次数: %d\n", successCount)
	fmt.Printf("错误次数: %d\n", errorCount)
	fmt.Printf("平均QPS: %.2f\n", float64(successCount)/duration.Seconds())
}
