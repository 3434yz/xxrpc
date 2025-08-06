package benchmark

import (
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"xxrpc/client"
	"xxrpc/examples/simple/echo"
)

func setupEchoServer(tb *testing.B) *exec.Cmd {
	tb.Helper()

	cmd := exec.Command("go", "run", "xxrpc/examples/simple/server/.")
	if err := cmd.Start(); err != nil {
		tb.Fatalf("failed to start echo server: %v", err)
	}

	time.Sleep(time.Second)
	return cmd
}

func stopEchoServer(cmd *exec.Cmd) {
	if cmd != nil {
		if err := cmd.Process.Kill(); err != nil {
			cmd.Process.Release()
		}
	}
}

func BenchmarkEcho(b *testing.B) {
	cmd := setupEchoServer(b)
	defer stopEchoServer(cmd)

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

	var (
		successCount uint64
		failCount    uint64
		totalLatency int64 // 纳秒
	)

	b.ResetTimer()

	start := time.Now()

	b.RunParallel(func(p *testing.PB) {
		cli, err := client.Dial(":8888")
		if err != nil {
			b.Fatalf("client dial error: %v", err)
		}

		for p.Next() {
			begin := time.Now()
			_, err := cli.Call("EchoService.ComplexHello", req)
			elapsed := time.Since(begin).Nanoseconds()

			atomic.AddInt64(&totalLatency, elapsed)

			if err != nil {
				atomic.AddUint64(&failCount, 1)
			} else {
				atomic.AddUint64(&successCount, 1)
			}
		}
	})

	b.StopTimer()

	duration := time.Since(start)
	total := successCount + failCount
	avgLatency := time.Duration(totalLatency / int64(max(1, total)))
	qps := float64(successCount) / duration.Seconds()

	b.Logf("Total Requests: %d", total)
	b.Logf("Successful Count: %d", successCount)
	b.Logf("Failed: Count: %d", failCount)
	b.Logf("Average Latency: %v", avgLatency)
	b.Logf("QPS: %.2f", qps)
}
