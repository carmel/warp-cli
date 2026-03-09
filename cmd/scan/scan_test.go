package scan

import (
	"os"
	"testing"

	"github.com/carmel/warp-cli/util"
)

func TestRun(t *testing.T) {
	// 设置测试参数
	util.Routines = 50
	util.PingTimes = 1
	util.MaxScanCount = 1000 // 增加到 1000
	util.PrintNum = 10
	util.Output = "test_result.csv"

	t.Log("Starting scan test...")
	err := scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	t.Log("Scan completed")

	// 检查 CSV 文件
	t.Log("Checking CSV file...")
	data, err := os.ReadFile("test_result.csv")
	if err != nil {
		// 文件可能在 cmd/scan 目录下
		data, err = os.ReadFile("cmd/scan/test_result.csv")
		if err != nil {
			t.Logf("CSV file not found (this is OK if no IPs were found)")
			return
		}
	}
	t.Logf("CSV content:\n%s", string(data))
}
