package disconnect

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/carmel/warp-cli/util"
	"github.com/spf13/cobra"
)

var (
	interfaceName string
	shortMsg      = "Disconnect from Cloudflare Warp VPN"
)

var Cmd = &cobra.Command{
	Use:   "disconnect",
	Short: shortMsg,
	Long: util.FormatMessage(shortMsg, `
This command will stop the WireGuard interface and clean up.

Example:
  warp-cli disconnect
  warp-cli disconnect -i wg0`),
	Run: func(cmd *cobra.Command, args []string) {
		util.RunCommandFatal(disconnectVPN)
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&interfaceName, "interface", "i", getDefaultInterface(), "WireGuard interface name")
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

func disconnectVPN() error {
	// Check if interface exists
	if !isInterfaceUp(interfaceName) {
		return fmt.Errorf("interface %s is not up or does not exist", interfaceName)
	}

	log.Printf("Disconnecting interface %s...", interfaceName)

	// Try to remove interface using ip link del
	if err := removeInterface(); err != nil {
		log.Printf("Failed to remove interface with ip/ifconfig: %v", err)
		log.Printf("Trying to remove control socket...")

		// Try removing control socket as fallback
		if err := removeControlSocket(); err != nil {
			return fmt.Errorf("failed to disconnect: %v", err)
		}
	}

	log.Printf("✓ Successfully disconnected from Cloudflare Warp!")
	return nil
}

func removeInterface() error {
	switch runtime.GOOS {
	case "linux":
		// On Linux, use ip link del
		cmd := exec.Command("ip", "link", "del", interfaceName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("ip link del failed: %v\nOutput: %s", err, string(output))
		}
		return nil

	case "darwin", "freebsd", "openbsd":
		// On BSD systems, try removing control socket
		return removeControlSocket()

	case "windows":
		// On Windows, the interface should be removed when wireguard-go exits
		// Try to find and kill the process
		return killWireguardProcess()

	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func removeControlSocket() error {
	socketPath := filepath.Join("/var/run/wireguard", interfaceName+".sock")

	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return fmt.Errorf("control socket not found: %s", socketPath)
	}

	if err := os.Remove(socketPath); err != nil {
		return fmt.Errorf("failed to remove control socket: %v", err)
	}

	log.Printf("Removed control socket: %s", socketPath)
	return nil
}

func killWireguardProcess() error {
	// Try to find wireguard-go process for this interface
	// This is a simple implementation, might need improvement
	cmd := exec.Command("pkill", "-f", "wireguard-go.*"+interfaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill wireguard-go process: %v", err)
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
