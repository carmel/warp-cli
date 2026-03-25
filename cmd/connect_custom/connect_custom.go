package connect_custom

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/carmel/warp-cli/util"
	"github.com/spf13/cobra"
)

var (
	configFile    string
	interfaceName string
	foreground    bool
	shortMsg      = "Connect to a custom WireGuard server"
)

var Cmd = &cobra.Command{
	Use:   "connect-custom",
	Short: shortMsg,
	Long: util.FormatMessage(shortMsg, `
Connect to your own WireGuard server using a standard WireGuard config file.

This bypasses Cloudflare WARP and connects directly to your server.
You can use any standard WireGuard configuration file.

Example:
  warp-cli connect-custom --config my-server.conf
  warp-cli connect-custom --config my-server.conf --interface wg1
  warp-cli connect-custom --config my-server.conf --foreground

Config file format:
  [Interface]
  PrivateKey = <your-private-key>
  Address = 10.0.0.2/24
  DNS = 8.8.8.8

  [Peer]
  PublicKey = <server-public-key>
  Endpoint = your-server.com:51820
  AllowedIPs = 0.0.0.0/0
  PersistentKeepalive = 25`),
	Run: func(cmd *cobra.Command, args []string) {
		util.RunCommandFatal(connectCustom)
	},
}

func init() {
	Cmd.Flags().StringVarP(&configFile, "config", "c", "", "WireGuard config file (required)")
	Cmd.Flags().StringVarP(&interfaceName, "interface", "i", getDefaultInterface(), "Interface name")
	Cmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run wireguard-go in foreground")
	Cmd.MarkFlagRequired("config")
}

func getDefaultInterface() string {
	switch runtime.GOOS {
	case "darwin":
		return "utun"
	case "openbsd":
		return "tun"
	default:
		return "wg0"
	}
}

func connectCustom() error {
	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configFile)
	}

	// 验证配置文件格式
	if err := validateConfig(); err != nil {
		return fmt.Errorf("invalid config file: %v", err)
	}

	log.Printf("Connecting to custom WireGuard server...")
	log.Printf("  Config: %s", configFile)
	log.Printf("  Interface: %s", interfaceName)

	// 查找 wireguard-go
	wireguardPath, err := findWireguard()
	if err != nil {
		return fmt.Errorf("wireguard not found: %v\nPlease build it first: cd wireguard && make", err)
	}
	log.Printf("  Using: %s", wireguardPath)

	// 检查接口是否已存在
	if isInterfaceUp(interfaceName) {
		return fmt.Errorf("interface %s already exists. Please disconnect first or use a different interface name", interfaceName)
	}

	// 启动 wireguard-go
	log.Printf("Starting wireguard for interface %s...", interfaceName)
	if err := startWireguard(wireguardPath); err != nil {
		return fmt.Errorf("failed to start wireguard-go: %v", err)
	}

	// 等待接口创建
	if !foreground {
		time.Sleep(1 * time.Second)
		if !isInterfaceUp(interfaceName) {
			return fmt.Errorf("interface %s was not created", interfaceName)
		}
	}

	// 应用配置
	log.Printf("Applying configuration to %s...", interfaceName)
	if err := applyConfig(); err != nil {
		return fmt.Errorf("failed to apply configuration: %v", err)
	}

	log.Println("\n✓ Successfully connected to custom WireGuard server!")
	log.Printf("  Interface: %s", interfaceName)
	log.Printf("  Config: %s", configFile)
	log.Printf("\nTo disconnect, run: warp-cli disconnect -i %s", interfaceName)

	return nil
}

func validateConfig() error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config: %v", err)
	}

	content := string(data)

	// 检查必需的字段
	required := map[string]string{
		"[Interface]": "Interface section",
		"PrivateKey":  "Private key",
		"[Peer]":      "Peer section",
		"PublicKey":   "Peer public key",
		"Endpoint":    "Peer endpoint",
	}

	for field, description := range required {
		if !strings.Contains(content, field) {
			return fmt.Errorf("missing required field: %s (%s)", field, description)
		}
	}

	log.Println("✓ Config file validation passed")
	return nil
}

func findWireguard() (string, error) {
	// Try in PATH
	path, err := exec.LookPath("wireguard")
	if err == nil {
		return path, nil
	}

	return "", fmt.Errorf("wireguard not found in PATH")
}

func startWireguard(wireguardPath string) error {
	args := []string{interfaceName}
	if foreground {
		args = append([]string{"-f"}, args...)
	}

	cmd := exec.Command(wireguardPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if foreground {
		// Run in foreground (blocking)
		return cmd.Run()
	}

	// Start in background
	if err := cmd.Start(); err != nil {
		return err
	}

	// Don't wait for the process
	go cmd.Wait()

	return nil
}

func applyConfig() error {
	// Get absolute path to config
	absConfig, err := filepath.Abs(configFile)
	if err != nil {
		return fmt.Errorf("get absolute path: %v", err)
	}

	// Use wg setconf
	cmd := exec.Command("wg", "setconf", interfaceName, absConfig)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wg setconf failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}

func isInterfaceUp(iface string) bool {
	switch runtime.GOOS {
	case "darwin", "linux", "freebsd", "openbsd":
		cmd := exec.Command("ip", "link", "show", iface)
		if err := cmd.Run(); err == nil {
			return true
		}
		// Try ifconfig as fallback
		cmd = exec.Command("ifconfig", iface)
		return cmd.Run() == nil
	case "windows":
		cmd := exec.Command("wg", "show", iface)
		return cmd.Run() == nil
	}
	return false
}
