package scan

import (
	"testing"

	"github.com/carmel/warp-cli/util"
)

func TestFinalScan(t *testing.T) {
	// 使用和 TestCompareWithActualWarping 相同的设置
	util.Routines = 20
	util.PingTimes = 1
	util.MaxScanCount = 500
	util.PrintNum = 10
	util.Output = "final_result.csv"

	// 不指定 IPText，使用默认 IP 范围
	util.IPText = ""
	util.IPFile = ""

	t.Log("Starting final scan with default IP ranges...")
	err := scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	t.Log("Scan completed - check final_result.csv")
}
