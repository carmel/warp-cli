package connect

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/carmel/warp-cli/cloudflare"
	"github.com/carmel/warp-cli/config"
	"github.com/carmel/warp-cli/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	profileFile   string
	interfaceName string
	foreground    bool
	shortMsg      = "Connect to Cloudflare Warp VPN"
)

var Cmd = &cobra.Command{
	Use:   "connect",
	Short: shortMsg,
	Long: util.FormatMessage(shortMsg, `
This command will:
1. Generate a WireGuard profile if it doesn't exist
2. Start wireguard-go to create the interface
3. Apply the configuration using wg setconf

Example:
  warp-cli connect
  warp-cli connect -i wg0 -p custom-profile.conf
  warp-cli connect -f  # run in foreground`),
	Run: func(cmd *cobra.Command, args []string) {
		util.RunCommandFatal(connectVPN)
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&profileFile, "profile", "p", "warp-cli-profile.conf", "WireGuard profile file")
	Cmd.PersistentFlags().StringVarP(&interfaceName, "interface", "i", getDefaultInterface(), "WireGuard interface name")
	Cmd.PersistentFlags().BoolVarP(&foreground, "foreground", "f", false, "Run wireguard-go in foreground")
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

func connectVPN() error {
	// Check if account is valid
	if !config.IsAccountValid() {
		return errors.New("no valid account found. Please run 'warp-cli register' first")
	}

	// Generate profile if it doesn't exist
	if _, err := os.Stat(profileFile); os.IsNotExist(err) {
		log.Printf("Profile file %s not found, generating...", profileFile)
		if err := generateProfile(); err != nil {
			return fmt.Errorf("failed to generate profile: %v", err)
		}
	} else {
		log.Printf("Using existing profile: %s", profileFile)
	}

	// On Linux, prefer wg-quick if available (more reliable with kernel module)
	if runtime.GOOS == "linux" {
		if _, err := exec.LookPath("wg-quick"); err == nil {
			log.Println("Using wg-quick (recommended for Linux)")
			return connectWithWgQuick()
		}
	}

	// Fallback to manual configuration
	log.Println("Using manual WireGuard configuration")
	return connectManual()
}

func connectWithWgQuick() error {
	// Check if interface already exists
	if isInterfaceUp(interfaceName) {
		return fmt.Errorf("interface %s already exists. Please disconnect first", interfaceName)
	}

	// wg-quick requires config file to be named <interface>.conf
	wgQuickConfig := interfaceName + ".conf"

	// // Copy profile to wg-quick compatible name if needed
	// if profileFile != wgQuickConfig {
	// 	input, err := os.ReadFile(profileFile)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to read profile: %v", err)
	// 	}
	// 	if err := os.WriteFile(wgQuickConfig, input, 0600); err != nil {
	// 		return fmt.Errorf("failed to create wg-quick config: %v", err)
	// 	}
	// 	defer os.Remove(wgQuickConfig) // Clean up temporary file
	// }

	// Use wg-quick to bring up the interface
	// wg-quick expects just the interface name (it will look for <name>.conf)

	util.CopyFile(profileFile, wgQuickConfig)

	confPath, _ := filepath.Abs(wgQuickConfig)
	log.Println("profile path: ", confPath)

	cmd := exec.Command("wg-quick", "up", confPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wg-quick up failed: %v\nOutput: %s", err, string(output))
	}

	log.Println("✓ WireGuard interface created")
	log.Println("⏳ Waiting for handshake (this may take up to 30 seconds)...")

	// Wait for handshake with timeout
	handshakeSuccess := false
	maxWait := 30      // seconds
	checkInterval := 2 // seconds

	for i := 0; i < maxWait/checkInterval; i++ {
		time.Sleep(time.Duration(checkInterval) * time.Second)

		// Check if handshake succeeded
		checkCmd := exec.Command("wg", "show", interfaceName)
		checkOutput, checkErr := checkCmd.Output()
		if checkErr == nil {
			outputStr := string(checkOutput)
			if contains(outputStr, "latest handshake:") {
				handshakeSuccess = true
				log.Println("✓ Handshake successful!")
				break
			}

			// Show progress
			if i%3 == 0 {
				log.Printf("  Still waiting... (%d/%d seconds)", (i+1)*checkInterval, maxWait)
			}
		}
	}

	if !handshakeSuccess {
		log.Println("⚠️  Warning: Handshake not completed within timeout")
		log.Println("   Connection may still succeed. Check status with: sudo wg show", interfaceName)
	}

	log.Printf("✓ Successfully connected to Cloudflare Warp!")
	log.Printf("  Interface: %s", interfaceName)
	log.Printf("  Profile: %s", profileFile)
	log.Printf("\nTo disconnect, run: warp-cli disconnect -i %s", interfaceName)
	log.Printf("To check status, run: sudo wg show %s", interfaceName)

	return nil
}

func connectManual() error {
	wireguardPath, err := findWireguardGo()
	if err != nil {
		return fmt.Errorf("wireguard not found: %v\nPlease build it first: cd wireguard && make", err)
	}
	log.Printf("Found wireguard at: %s", wireguardPath)

	// Check if interface already exists
	if isInterfaceUp(interfaceName) {
		return fmt.Errorf("interface %s already exists. Please disconnect first or use a different interface name", interfaceName)
	}

	// Start wireguard-go
	log.Printf("Starting wireguard for interface %s...", interfaceName)
	if err := startWireguard(wireguardPath); err != nil {
		return fmt.Errorf("failed to start wireguard-go: %v", err)
	}

	// Wait for interface to be created
	if !foreground {
		time.Sleep(1 * time.Second)
		if !isInterfaceUp(interfaceName) {
			return fmt.Errorf("interface %s was not created", interfaceName)
		}
	}

	// Apply configuration
	log.Printf("Applying configuration to %s...", interfaceName)
	if err := applyConfig(); err != nil {
		return fmt.Errorf("failed to apply configuration: %v", err)
	}

	log.Printf("✓ Successfully connected to Cloudflare Warp!")
	log.Printf("  Interface: %s", interfaceName)
	log.Printf("  Profile: %s", profileFile)
	log.Printf("\nTo disconnect, run: warp-cli disconnect -i %s", interfaceName)

	return nil
}

func generateProfile() error {
	ctx := config.CreateContext()
	thisDevice, err := cloudflare.GetSourceDevice(ctx)
	if err != nil {
		return fmt.Errorf("GetSourceDevice: %v", err)
	}

	profile, err := util.NewProfile(&util.ProfileData{
		PrivateKey: viper.GetString(config.PrivateKey),
		Address1:   thisDevice.Config.Interface.Addresses.V4,
		Address2:   thisDevice.Config.Interface.Addresses.V6,
		PublicKey:  thisDevice.Config.Peers[0].PublicKey,
		Endpoint:   thisDevice.Config.Peers[0].Endpoint.Host,
	})
	if err != nil {
		return fmt.Errorf("create profile: %v", err)
	}
	if err := profile.Save(profileFile); err != nil {
		return fmt.Errorf("save profile: %v", err)
	}

	log.Printf("Generated profile: %s", profileFile)
	return nil
}

func findWireguardGo() (string, error) {
	// Try in PATH
	path, err := exec.LookPath("wireguard")
	if err == nil {
		return path, nil
	}

	return "", errors.New("wireguard not found in PATH")
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
	// Parse the profile to extract configuration
	profileData, err := parseProfile(profileFile)
	if err != nil {
		return fmt.Errorf("parse profile: %v", err)
	}

	// Create a temporary config file with only WireGuard-compatible fields
	tmpConfig, err := createWgConfig(profileData)
	if err != nil {
		return fmt.Errorf("create wg config: %v", err)
	}
	defer os.Remove(tmpConfig)

	// Apply WireGuard configuration
	cmd := exec.Command("wg", "setconf", interfaceName, tmpConfig)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wg setconf failed: %v\nOutput: %s", err, string(output))
	}

	// Configure IP addresses
	if err := configureAddresses(profileData.Addresses); err != nil {
		return fmt.Errorf("configure addresses: %v", err)
	}

	// Configure MTU if specified
	if profileData.MTU != "" {
		if err := configureMTU(profileData.MTU); err != nil {
			log.Printf("Warning: failed to set MTU: %v", err)
		}
	}

	// Configure routing
	if err := configureRouting(profileData.Addresses); err != nil {
		log.Printf("Warning: failed to configure routing: %v", err)
		log.Println("You may need to manually configure routing:")
		log.Printf("  sudo ip route add default dev %s", interfaceName)
	}

	// Add endpoint route (critical: endpoint must go through physical interface)
	if err := addEndpointRoute(profileData.Peer.Endpoint); err != nil {
		log.Printf("Warning: failed to add endpoint route: %v", err)
	}

	// Configure DNS
	if err := configureDNS(profileData.DNS); err != nil {
		log.Printf("Warning: failed to configure DNS: %v", err)
		log.Println("You may need to manually configure DNS:")
		log.Println("  sudo resolvectl dns", interfaceName, "1.1.1.1")
	}

	// Configure firewall
	if err := configureFirewall(); err != nil {
		log.Printf("Warning: failed to configure firewall: %v", err)
		log.Println("You may need to manually configure firewall:")
		log.Printf("  sudo firewall-cmd --zone=trusted --add-interface=%s\n", interfaceName)
	}

	return nil
}

type ProfileData struct {
	PrivateKey string
	Addresses  []string
	DNS        []string
	MTU        string
	Peer       PeerData
}

type PeerData struct {
	PublicKey  string
	Endpoint   string
	AllowedIPs []string
}

func parseProfile(path string) (*ProfileData, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data := &ProfileData{}
	lines := string(content)

	// Simple parser for WireGuard config
	for _, line := range splitLines(lines) {
		line = trimSpace(line)
		if line == "" || line[0] == '#' || line[0] == '[' {
			continue
		}

		parts := splitKeyValue(line)
		if len(parts) != 2 {
			continue
		}

		key := trimSpace(parts[0])
		value := trimSpace(parts[1])

		switch key {
		case "PrivateKey":
			data.PrivateKey = value
		case "Address":
			data.Addresses = splitComma(value)
		case "DNS":
			data.DNS = splitComma(value)
		case "MTU":
			data.MTU = value
		case "PublicKey":
			data.Peer.PublicKey = value
		case "Endpoint":
			data.Peer.Endpoint = value
		case "AllowedIPs":
			data.Peer.AllowedIPs = splitComma(value)
		}
	}

	return data, nil
}

func createWgConfig(data *ProfileData) (string, error) {
	tmpFile, err := os.CreateTemp("", "wg-*.conf")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// Write only WireGuard-compatible fields
	content := fmt.Sprintf(`[Interface]
PrivateKey = %s

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = %s
`, data.PrivateKey, data.Peer.PublicKey, data.Peer.Endpoint, joinComma(data.Peer.AllowedIPs))

	if _, err := tmpFile.WriteString(content); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func configureAddresses(addresses []string) error {
	for _, addr := range addresses {
		addr = trimSpace(addr)
		if addr == "" {
			continue
		}

		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "linux":
			cmd = exec.Command("ip", "addr", "add", addr, "dev", interfaceName)
		case "darwin", "freebsd", "openbsd":
			// On BSD systems, use ifconfig
			// Split CIDR notation
			parts := splitSlash(addr)
			if len(parts) == 2 {
				cmd = exec.Command("ifconfig", interfaceName, "inet", parts[0], parts[0], "netmask", cidrToNetmask(parts[1]))
			} else {
				cmd = exec.Command("ifconfig", interfaceName, "inet", addr, addr)
			}
		default:
			return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to add address %s: %v\nOutput: %s", addr, err, string(output))
		}
	}

	// Bring interface up
	if err := bringInterfaceUp(); err != nil {
		return fmt.Errorf("bring interface up: %v", err)
	}

	return nil
}

func configureMTU(mtu string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("ip", "link", "set", "mtu", mtu, "dev", interfaceName)
	case "darwin", "freebsd", "openbsd":
		cmd = exec.Command("ifconfig", interfaceName, "mtu", mtu)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set MTU: %v\nOutput: %s", err, string(output))
	}

	return nil
}

func bringInterfaceUp() error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("ip", "link", "set", "up", "dev", interfaceName)
	case "darwin", "freebsd", "openbsd":
		cmd = exec.Command("ifconfig", interfaceName, "up")
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to bring interface up: %v\nOutput: %s", err, string(output))
	}

	return nil
}

// Helper functions
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitKeyValue(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func splitComma(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			parts = append(parts, trimSpace(s[start:i]))
			start = i + 1
		}
	}
	if start < len(s) {
		parts = append(parts, trimSpace(s[start:]))
	}
	return parts
}

func splitSlash(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func joinComma(parts []string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += ", "
		}
		result += part
	}
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func cidrToNetmask(cidr string) string {
	// Simple CIDR to netmask conversion for common cases
	masks := map[string]string{
		"8":  "255.0.0.0",
		"16": "255.255.0.0",
		"24": "255.255.255.0",
		"32": "255.255.255.255",
	}
	if mask, ok := masks[cidr]; ok {
		return mask
	}
	return "255.255.255.255"
}

func configureRouting(addresses []string) error {
	// Get the first IPv4 address for routing rules
	var vpnIP string
	for _, addr := range addresses {
		addr = trimSpace(addr)
		if addr != "" && !contains(addr, ":") {
			// IPv4 address
			parts := splitSlash(addr)
			if len(parts) > 0 {
				vpnIP = parts[0]
				break
			}
		}
	}

	switch runtime.GOOS {
	case "linux":
		// Create routing table if not exists
		tableName := "warp"
		tableNum := "200"

		// Check if table exists in rt_tables
		content, err := os.ReadFile("/etc/iproute2/rt_tables")
		if err == nil && !contains(string(content), tableName) {
			// Add table
			f, err := os.OpenFile("/etc/iproute2/rt_tables", os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				f.WriteString(fmt.Sprintf("\n%s %s\n", tableNum, tableName))
				f.Close()
			}
		}

		// Add default route to warp table
		cmd := exec.Command("ip", "route", "add", "default", "dev", interfaceName, "table", tableName)
		cmd.Run() // Ignore error if route already exists

		// Add rule to use warp table for VPN traffic
		if vpnIP != "" {
			cmd = exec.Command("ip", "rule", "add", "from", vpnIP, "table", tableName)
			cmd.Run() // Ignore error if rule already exists
		}

		// Add split routes (0.0.0.0/1 and 128.0.0.0/1 covers all IPs)
		// This is less invasive than replacing default route
		cmd = exec.Command("ip", "route", "add", "0.0.0.0/1", "dev", interfaceName, "metric", "100")
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("Note: Could not add route 0.0.0.0/1: %v", err)
			log.Printf("Output: %s", string(output))
		}

		cmd = exec.Command("ip", "route", "add", "128.0.0.0/1", "dev", interfaceName, "metric", "100")
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("Note: Could not add route 128.0.0.0/1: %v", err)
			log.Printf("Output: %s", string(output))
		}

		log.Println("✓ Routing configured")
		return nil

	case "darwin":
		// On macOS, add routes for all traffic
		cmd := exec.Command("route", "add", "-net", "0.0.0.0/1", "-interface", interfaceName)
		cmd.Run()

		cmd = exec.Command("route", "add", "-net", "128.0.0.0/1", "-interface", interfaceName)
		cmd.Run()

		log.Println("✓ Routing configured")
		return nil

	default:
		return fmt.Errorf("automatic routing not supported on %s", runtime.GOOS)
	}
}

func configureDNS(dnsServers []string) error {
	if len(dnsServers) == 0 {
		return nil
	}

	switch runtime.GOOS {
	case "linux":
		// Try systemd-resolved first
		if _, err := exec.LookPath("resolvectl"); err == nil {
			// Use resolvectl
			args := []string{"dns", interfaceName}
			args = append(args, dnsServers...)
			cmd := exec.Command("resolvectl", args...)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("resolvectl dns failed: %v", err)
			}

			// Set as default DNS
			cmd = exec.Command("resolvectl", "domain", interfaceName, "~.")
			cmd.Run() // Ignore error

			log.Println("✓ DNS configured via systemd-resolved")
			return nil
		}

		// Fallback: modify /etc/resolv.conf
		// Backup original
		if _, err := os.Stat("/etc/resolv.conf.warp-backup"); os.IsNotExist(err) {
			input, err := os.ReadFile("/etc/resolv.conf")
			if err == nil {
				os.WriteFile("/etc/resolv.conf.warp-backup", input, 0644)
			}
		}

		// Write new resolv.conf
		content := ""
		for _, dns := range dnsServers {
			dns = trimSpace(dns)
			if dns != "" {
				content += fmt.Sprintf("nameserver %s\n", dns)
			}
		}

		if err := os.WriteFile("/etc/resolv.conf", []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write resolv.conf: %v", err)
		}

		log.Println("✓ DNS configured via /etc/resolv.conf")
		return nil

	case "darwin":
		// On macOS, use networksetup
		// This is more complex and requires knowing the network service name
		log.Println("Note: DNS configuration on macOS requires manual setup")
		log.Println("Run: sudo networksetup -setdnsservers Wi-Fi 1.1.1.1 1.0.0.1")
		return nil

	default:
		return fmt.Errorf("automatic DNS configuration not supported on %s", runtime.GOOS)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)+1 && containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func configureFirewall() error {
	switch runtime.GOOS {
	case "linux":
		// Check if firewalld is running
		if _, err := exec.LookPath("firewall-cmd"); err != nil {
			// firewalld not installed
			return nil
		}

		// Check if firewalld is running
		cmd := exec.Command("firewall-cmd", "--state")
		if err := cmd.Run(); err != nil {
			// firewalld not running
			return nil
		}

		// Add interface to trusted zone
		cmd = exec.Command("firewall-cmd", "--zone=trusted", "--add-interface="+interfaceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add interface to trusted zone: %v", err)
		}

		// Make it permanent
		cmd = exec.Command("firewall-cmd", "--permanent", "--zone=trusted", "--add-interface="+interfaceName)
		cmd.Run() // Ignore error for permanent rule

		log.Println("✓ Firewall configured (interface added to trusted zone)")
		return nil

	default:
		// No automatic firewall configuration for other OS
		return nil
	}
}

func addEndpointRoute(endpoint string) error {
	// Extract IP from endpoint (format: host:port or ip:port)
	parts := splitColon(endpoint)
	if len(parts) < 2 {
		return fmt.Errorf("invalid endpoint format: %s", endpoint)
	}

	endpointHost := parts[0]

	// Resolve if it's a hostname
	var endpointIP string
	cmd := exec.Command("getent", "hosts", endpointHost)
	output, err := cmd.Output()
	if err == nil {
		// Parse getent output
		fields := splitSpace(string(output))
		if len(fields) > 0 {
			endpointIP = fields[0]
		}
	} else {
		// Assume it's already an IP
		endpointIP = endpointHost
	}

	if endpointIP == "" {
		return fmt.Errorf("failed to resolve endpoint: %s", endpointHost)
	}

	switch runtime.GOOS {
	case "linux":
		// Get default gateway and interface
		cmd := exec.Command("ip", "route", "show", "default")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get default route: %v", err)
		}

		// Parse: default via 192.168.1.1 dev eth0
		routeStr := string(output)
		var gateway, iface string

		fields := splitSpace(routeStr)
		for i, field := range fields {
			if field == "via" && i+1 < len(fields) {
				gateway = fields[i+1]
			}
			if field == "dev" && i+1 < len(fields) {
				iface = fields[i+1]
			}
		}

		if gateway == "" || iface == "" {
			return fmt.Errorf("failed to parse default route")
		}

		// Add route for endpoint through physical interface
		cmd = exec.Command("ip", "route", "add", endpointIP, "via", gateway, "dev", iface)
		if err := cmd.Run(); err != nil {
			// Route might already exist, ignore error
			log.Printf("Note: Could not add endpoint route (may already exist)")
		} else {
			log.Printf("✓ Added endpoint route: %s via %s dev %s", endpointIP, gateway, iface)
		}

		return nil

	case "darwin":
		// On macOS
		cmd := exec.Command("route", "add", "-host", endpointIP, "-interface", "en0")
		if err := cmd.Run(); err != nil {
			log.Printf("Note: Could not add endpoint route")
		}
		return nil

	default:
		return fmt.Errorf("automatic endpoint routing not supported on %s", runtime.GOOS)
	}
}

func splitColon(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	return parts
}

func splitSpace(s string) []string {
	var parts []string
	var current string
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(s[i])
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
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
		// On Windows, check with netsh or wg show
		cmd := exec.Command("wg", "show", iface)
		return cmd.Run() == nil
	}
	return false
}
