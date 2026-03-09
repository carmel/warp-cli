package scan

import (
	"fmt"
	"time"

	"github.com/carmel/warp-cli/util"
	"github.com/spf13/cobra"
)

var shortMsg = "Scan and test Cloudflare Warp endpoints for optimal performance"

var (
	minDelay    int
	maxDelay    int
	maxLossRate float64
)

var Cmd = &cobra.Command{
	Use:   "scan",
	Short: shortMsg,
	Long:  util.FormatMessage(shortMsg, ``),
	Run: func(cmd *cobra.Command, args []string) {
		util.RunCommandFatal(scan)
	},
}

func init() {
	Cmd.Flags().IntVarP(&util.Routines, "threads", "n", 200, "Number of concurrent test threads")
	Cmd.Flags().IntVarP(&util.PingTimes, "times", "t", 1, "Number of latency test times per endpoint")
	Cmd.Flags().IntVarP(&util.MaxScanCount, "count", "c", 5000, "Maximum number of addresses to scan")

	Cmd.Flags().IntVar(&maxDelay, "max-delay", 5000, "Maximum acceptable latency in milliseconds")
	Cmd.Flags().IntVar(&minDelay, "min-delay", 0, "Minimum acceptable latency in milliseconds")
	Cmd.Flags().Float64Var(&maxLossRate, "max-loss", 1.0, "Maximum acceptable packet loss rate (0.0-1.0)")

	Cmd.Flags().BoolVar(&util.AllMode, "all", false, "Test all IP and port combinations")
	Cmd.Flags().BoolVar(&util.IPv6Mode, "ipv6", false, "Scan IPv6 addresses only")
	Cmd.Flags().IntVarP(&util.PrintNum, "print", "p", 10, "Number of results to display")
	Cmd.Flags().StringVarP(&util.IPFile, "file", "f", "", "Path to IP data file")
	Cmd.Flags().StringVar(&util.IPText, "ip", "", "Specify IP addresses directly (comma-separated)")
	Cmd.Flags().StringVarP(&util.Output, "output", "o", "result.csv", "Output result file path")
	Cmd.Flags().StringVar(&util.PrivateKey, "private-key", "", "Custom WireGuard private key")
	Cmd.Flags().StringVar(&util.PublicKey, "public-key", "", "Custom WireGuard public key")
	Cmd.Flags().StringVar(&util.ReservedString, "reserved", "", "Custom reserved field for WireGuard")
}

func scan() error {
	// Apply delay and loss rate settings
	util.InputMaxDelay = time.Duration(maxDelay) * time.Millisecond
	util.InputMinDelay = time.Duration(minDelay) * time.Millisecond
	util.InputMaxLossRate = float32(maxLossRate)

	util.InitHandshakePacket()

	fmt.Printf("CloudflareWarpSpeedTest\n\n")

	pingData := util.NewWarping().Run().FilterDelay().FilterLossRate()
	util.ExportCsv(pingData)
	pingData.Print()

	return nil
}
