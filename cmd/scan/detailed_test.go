package scan

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/carmel/warp-cli/util"
)

// 模拟 Warping 结构，但带详细日志
type DebugWarping struct {
	wg      *sync.WaitGroup
	m       *sync.Mutex
	ips     []string
	results []string
	control chan bool
}

func TestDetailedScan(t *testing.T) {
	util.InitHandshakePacket()

	// 只测试几个已知可用的 IP
	testIPs := []string{
		"162.159.192.1:500",
		"162.159.193.1:854",
		"188.114.96.1:2408",
	}

	dw := &DebugWarping{
		wg:      &sync.WaitGroup{},
		m:       &sync.Mutex{},
		ips:     testIPs,
		results: make([]string, 0),
		control: make(chan bool, 2), // 并发数为 2
	}

	t.Log("Starting detailed scan...")

	for i, ip := range dw.ips {
		t.Logf("Queuing IP %d: %s", i+1, ip)
		dw.wg.Add(1)
		dw.control <- false
		go dw.testIP(ip, t)
	}

	dw.wg.Wait()

	t.Logf("\n=== Results ===")
	t.Logf("Total tested: %d", len(dw.ips))
	t.Logf("Successful: %d", len(dw.results))
	for i, result := range dw.results {
		t.Logf("  %d. %s", i+1, result)
	}
}

func (dw *DebugWarping) testIP(addr string, t *testing.T) {
	defer dw.wg.Done()
	defer func() { <-dw.control }()

	t.Logf("[%s] Starting test...", addr)

	// 连接
	conn, err := net.DialTimeout("udp", addr, 1*time.Second)
	if err != nil {
		t.Logf("[%s] ❌ Connection failed: %v", addr, err)
		return
	}
	defer conn.Close()

	t.Logf("[%s] ✓ Connected", addr)

	// 握手
	success, rtt := testHandshakeWithLog(conn, addr, t)
	if success {
		t.Logf("[%s] ✓ Handshake successful! RTT: %v", addr, rtt)
		dw.m.Lock()
		dw.results = append(dw.results, fmt.Sprintf("%s (RTT: %v)", addr, rtt))
		dw.m.Unlock()
	} else {
		t.Logf("[%s] ❌ Handshake failed", addr)
	}
}

func testHandshakeWithLog(conn net.Conn, addr string, t *testing.T) (bool, time.Duration) {
	handshakePacket := []byte{
		0x01, 0x3c, 0xbd, 0xaf, 0xb4, 0x13, 0x5c, 0xac,
		0x96, 0xa2, 0x94, 0x84, 0xd7, 0xa0, 0x17, 0x5a,
		0xb1, 0x52, 0xdd, 0x3e, 0x59, 0xbe, 0x35, 0x04,
		0x9b, 0xea, 0xdf, 0x75, 0x8b, 0x8d, 0x48, 0xaf,
		0x14, 0xca, 0x65, 0xf2, 0x5a, 0x16, 0x89, 0x34,
		0x74, 0x6f, 0xe8, 0xbc, 0x88, 0x67, 0xb1, 0xc1,
		0x71, 0x13, 0xd7, 0x1c, 0x0f, 0xac, 0x5c, 0x14,
		0x1e, 0xf9, 0xf3, 0x57, 0x83, 0xff, 0xa5, 0x35,
		0x7c, 0x98, 0x71, 0xf4, 0xa0, 0x06, 0x66, 0x2b,
		0x83, 0xad, 0x71, 0x24, 0x5a, 0x86, 0x24, 0x95,
		0x37, 0x6a, 0x5f, 0xe3, 0xb4, 0xf2, 0xe1, 0xf0,
		0x69, 0x74, 0xd7, 0x48, 0x41, 0x66, 0x70, 0xe5,
		0xf9, 0xb0, 0x86, 0x29, 0x7f, 0x65, 0x2e, 0x6d,
		0xfb, 0xf7, 0x42, 0xfb, 0xfc, 0x63, 0xc3, 0xd8,
		0xae, 0xb1, 0x75, 0xa3, 0xe9, 0xb7, 0x58, 0x2f,
		0xbc, 0x67, 0xc7, 0x75, 0x77, 0xe4, 0xc0, 0xb3,
		0x2b, 0x05, 0xf9, 0x29, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	startTime := time.Now()

	// 写入
	n, err := conn.Write(handshakePacket)
	if err != nil {
		t.Logf("[%s]   Write error: %v", addr, err)
		return false, 0
	}
	t.Logf("[%s]   Wrote %d bytes", addr, n)

	// 设置超时
	err = conn.SetDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		t.Logf("[%s]   SetDeadline error: %v", addr, err)
		return false, 0
	}

	// 读取
	revBuff := make([]byte, 1024)
	n, err = conn.Read(revBuff)
	if err != nil {
		t.Logf("[%s]   Read error: %v", addr, err)
		return false, 0
	}
	t.Logf("[%s]   Read %d bytes (expected 92)", addr, n)

	if n != 92 {
		t.Logf("[%s]   Wrong response size", addr)
		return false, 0
	}

	duration := time.Since(startTime)
	return true, duration
}

func TestCompareWithActualWarping(t *testing.T) {
	// 设置参数
	util.Routines = 2
	util.PingTimes = 1
	util.MaxScanCount = 3
	util.IPText = "162.159.192.1,162.159.193.1,188.114.96.1" // 指定 IP
	util.Output = ""                                         // 不输出文件

	util.InitHandshakePacket()

	t.Log("Testing with actual Warping implementation...")

	// 使用实际的 Warping
	warping := util.NewWarping()
	results := warping.Run()

	t.Logf("Found %d results", len(results))

	if len(results) == 0 {
		t.Log("❌ No results found - there's a bug in the Warping implementation")

		// 让我们手动测试看看是否是网络问题
		t.Log("\nManual test:")
		conn, err := net.DialTimeout("udp", "162.159.192.1:500", 2*time.Second)
		if err != nil {
			t.Logf("Manual connection failed: %v", err)
		} else {
			t.Log("Manual connection succeeded!")
			conn.Close()
		}
	} else {
		t.Log("✓ Results found!")
		for i, r := range results {
			t.Logf("  %d. %s - %v", i+1, r.IP, r.Delay)
		}
	}
}

// 测试进度条是否影响结果
func TestWithoutProgressBar(t *testing.T) {
	util.InitHandshakePacket()
	util.Routines = 2
	util.PingTimes = 1

	testIPs := []string{
		"162.159.192.1:500",
		"162.159.193.1:854",
	}

	successCount := 0

	for _, addr := range testIPs {
		conn, err := net.DialTimeout("udp", addr, 1*time.Second)
		if err != nil {
			continue
		}

		// 简单握手测试
		handshakePacket := []byte{
			0x01, 0x3c, 0xbd, 0xaf, 0xb4, 0x13, 0x5c, 0xac,
			0x96, 0xa2, 0x94, 0x84, 0xd7, 0xa0, 0x17, 0x5a,
			0xb1, 0x52, 0xdd, 0x3e, 0x59, 0xbe, 0x35, 0x04,
			0x9b, 0xea, 0xdf, 0x75, 0x8b, 0x8d, 0x48, 0xaf,
			0x14, 0xca, 0x65, 0xf2, 0x5a, 0x16, 0x89, 0x34,
			0x74, 0x6f, 0xe8, 0xbc, 0x88, 0x67, 0xb1, 0xc1,
			0x71, 0x13, 0xd7, 0x1c, 0x0f, 0xac, 0x5c, 0x14,
			0x1e, 0xf9, 0xf3, 0x57, 0x83, 0xff, 0xa5, 0x35,
			0x7c, 0x98, 0x71, 0xf4, 0xa0, 0x06, 0x66, 0x2b,
			0x83, 0xad, 0x71, 0x24, 0x5a, 0x86, 0x24, 0x95,
			0x37, 0x6a, 0x5f, 0xe3, 0xb4, 0xf2, 0xe1, 0xf0,
			0x69, 0x74, 0xd7, 0x48, 0x41, 0x66, 0x70, 0xe5,
			0xf9, 0xb0, 0x86, 0x29, 0x7f, 0x65, 0x2e, 0x6d,
			0xfb, 0xf7, 0x42, 0xfb, 0xfc, 0x63, 0xc3, 0xd8,
			0xae, 0xb1, 0x75, 0xa3, 0xe9, 0xb7, 0x58, 0x2f,
			0xbc, 0x67, 0xc7, 0x75, 0x77, 0xe4, 0xc0, 0xb3,
			0x2b, 0x05, 0xf9, 0x29, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
		}

		conn.Write(handshakePacket)
		conn.SetDeadline(time.Now().Add(1 * time.Second))

		revBuff := make([]byte, 1024)
		n, err := conn.Read(revBuff)
		conn.Close()

		if err == nil && n == 92 {
			successCount++
			t.Logf("✓ %s succeeded", addr)
		} else {
			t.Logf("❌ %s failed: n=%d, err=%v", addr, n, err)
		}
	}

	t.Logf("\nSuccess rate: %d/%d", successCount, len(testIPs))
}
