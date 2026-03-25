package generate_config

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/carmel/warp-cli/util"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	outputFile  string
	interactive bool
	template    bool
	autoKey     bool
	privateKey  string
	address     string
	dns         string
	publicKey   string
	endpoint    string
	allowedIPs  string
	keepalive   int
	shortMsg    = "Generate a WireGuard configuration file"
)

var Cmd = &cobra.Command{
	Use:   "generate-config",
	Short: shortMsg,
	Long: util.FormatMessage(shortMsg, `
Generate a WireGuard configuration file for use with connect-custom.

Three modes:
1. Template mode: Generate a template with placeholders
2. Interactive mode: Guided configuration creation
3. Direct mode: Specify all parameters via flags

Examples:
  # Generate a template
  warp-cli generate-config --template -o my-server.conf

  # Interactive mode
  warp-cli generate-config --interactive -o my-server.conf

  # Direct mode with flags
  warp-cli generate-config -o my-server.conf \
    --private-key "xxx" \
    --address "10.0.0.2/24" \
    --public-key "yyy" \
    --endpoint "server.com:51820"`),
	Run: func(cmd *cobra.Command, args []string) {
		util.RunCommandFatal(generateConfig)
	},
}

func init() {
	Cmd.Flags().StringVarP(&outputFile, "output", "o", "wireguard.conf", "Output file path")
	Cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode")
	Cmd.Flags().BoolVarP(&template, "template", "t", false, "Generate template with placeholders")
	Cmd.Flags().BoolVar(&autoKey, "auto-key", true, "Auto-generate private key in template mode")

	// Direct mode flags
	Cmd.Flags().StringVar(&privateKey, "private-key", "", "Client private key")
	Cmd.Flags().StringVar(&address, "address", "", "Client IP address (e.g., 10.0.0.2/24)")
	Cmd.Flags().StringVar(&dns, "dns", "1.1.1.1, 8.8.8.8", "DNS servers")
	Cmd.Flags().StringVar(&publicKey, "public-key", "", "Server public key")
	Cmd.Flags().StringVar(&endpoint, "endpoint", "", "Server endpoint (e.g., server.com:51820)")
	Cmd.Flags().StringVar(&allowedIPs, "allowed-ips", "0.0.0.0/0, ::/0", "Allowed IPs")
	Cmd.Flags().IntVar(&keepalive, "keepalive", 25, "Persistent keepalive interval (seconds)")
}

func generateConfig() error {
	// Check if file already exists
	if _, err := os.Stat(outputFile); err == nil {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("File %s already exists. Overwrite", outputFile),
			IsConfirm: true,
		}
		if _, err := prompt.Run(); err != nil {
			return fmt.Errorf("operation cancelled")
		}
	}

	var config string
	var err error
	var generatedPrivKey, generatedPubKey string

	if template {
		config, generatedPrivKey, generatedPubKey, err = generateTemplate()
		if err != nil {
			return err
		}
	} else if interactive {
		config, err = generateInteractive()
		if err != nil {
			return err
		}
	} else {
		config, err = generateDirect()
		if err != nil {
			return err
		}
	}

	// Write to file
	if err := os.WriteFile(outputFile, []byte(config), 0600); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	log.Printf("✓ Successfully generated WireGuard configuration!")
	log.Printf("  Output: %s", outputFile)

	// Show generated keys if in template mode with auto-key and keys were actually generated
	if template && autoKey && generatedPrivKey != "" && generatedPrivKey != "<YOUR_PRIVATE_KEY>" {
		log.Println("\n" + strings.Repeat("=", 70))
		log.Println("Generated Keys:")
		log.Println(strings.Repeat("=", 70))
		log.Printf("Private Key: %s", generatedPrivKey)
		log.Printf("Public Key:  %s", generatedPubKey)
		log.Println(strings.Repeat("=", 70))
		log.Println("\n⚠️  IMPORTANT:")
		log.Println("  - Keep your private key secure!")
		log.Println("  - Share your PUBLIC KEY with the server administrator")
		log.Println("  - The private key has been saved to the config file")
	}

	log.Printf("\nTo connect using this config:")
	log.Printf("  warp-cli connect-custom --config %s", outputFile)

	return nil
}

func generateTemplate() (string, string, string, error) {
	var privKey, pubKey string
	var err error

	// Auto-generate key if enabled
	if autoKey {
		privKey, pubKey, err = generateKeyPair()
		if err != nil {
			log.Printf("⚠️  Warning: Failed to auto-generate key: %v", err)
			log.Println("   Falling back to placeholder. You can generate keys manually with: wg genkey")
			privKey = "<YOUR_PRIVATE_KEY>"
			pubKey = ""
		}
	} else {
		privKey = "<YOUR_PRIVATE_KEY>"
	}

	config := fmt.Sprintf(`[Interface]
# Your client's private key%s
PrivateKey = %s

# Your client's IP address in the VPN network
Address = 10.0.0.2/24

# DNS servers (optional)
DNS = 1.1.1.1, 8.8.8.8

# MTU (optional, default is 1420)
# MTU = 1420

[Peer]
# Server's public key
PublicKey = <SERVER_PUBLIC_KEY>

# Optional: Pre-shared key for additional security (generate with: wg genpsk)
# PresharedKey = <PRESHARED_KEY>

# Server's endpoint (IP:port or domain:port)
Endpoint = <SERVER_IP_OR_DOMAIN>:51820

# Which traffic to route through the VPN
# 0.0.0.0/0, ::/0 = all traffic (full tunnel)
# 10.0.0.0/24 = only specific subnet (split tunnel)
AllowedIPs = 0.0.0.0/0, ::/0

# Keep connection alive (recommended for NAT traversal)
PersistentKeepalive = 25
`, func() string {
		if autoKey && privKey != "<YOUR_PRIVATE_KEY>" {
			return " (auto-generated)"
		}
		return " (generate with: wg genkey)"
	}(), privKey)

	return config, privKey, pubKey, nil
}

func generateInteractive() (string, error) {
	log.Println("Interactive WireGuard Configuration Generator")
	log.Println(strings.Repeat("=", 50))

	// Generate or input private key
	privKey, err := promptPrivateKey()
	if err != nil {
		return "", err
	}

	// Client address
	addr, err := promptString("Client IP Address", "10.0.0.2/24", true)
	if err != nil {
		return "", err
	}

	// DNS servers
	dnsServers, err := promptString("DNS Servers", "1.1.1.1, 8.8.8.8", false)
	if err != nil {
		return "", err
	}

	// Server public key
	pubKey, err := promptString("Server Public Key", "", true)
	if err != nil {
		return "", err
	}

	// Server endpoint
	ep, err := promptString("Server Endpoint (IP:port or domain:port)", "server.example.com:51820", true)
	if err != nil {
		return "", err
	}

	// Allowed IPs
	allowed, err := promptString("Allowed IPs", "0.0.0.0/0, ::/0", false)
	if err != nil {
		return "", err
	}

	// Keepalive
	ka, err := promptInt("Persistent Keepalive (seconds, 0 to disable)", 25)
	if err != nil {
		return "", err
	}

	// Build config
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = %s

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = %s
`, privKey, addr, dnsServers, pubKey, ep, allowed)

	if ka > 0 {
		config += fmt.Sprintf("PersistentKeepalive = %d\n", ka)
	}

	return config, nil
}

func generateDirect() (string, error) {
	// Validate required fields
	if privateKey == "" {
		return "", fmt.Errorf("--private-key is required (or use --interactive/--template)")
	}
	if address == "" {
		return "", fmt.Errorf("--address is required")
	}
	if publicKey == "" {
		return "", fmt.Errorf("--public-key is required")
	}
	if endpoint == "" {
		return "", fmt.Errorf("--endpoint is required")
	}

	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = %s

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = %s
`, privateKey, address, dns, publicKey, endpoint, allowedIPs)

	if keepalive > 0 {
		config += fmt.Sprintf("PersistentKeepalive = %d\n", keepalive)
	}

	return config, nil
}

func promptPrivateKey() (string, error) {
	prompt := promptui.Select{
		Label: "Private Key",
		Items: []string{"Generate new key", "Enter existing key"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	if result == "Generate new key" {
		// Generate new key pair
		privKey, pubKey, err := generateKeyPair()
		if err != nil {
			return "", err
		}

		// Show the generated keys
		log.Printf("Generated private key: %s", privKey)
		log.Printf("Corresponding public key: %s", pubKey)
		log.Println("(Share the public key with your server administrator)")

		return privKey, nil
	}

	// Enter existing key
	return promptString("Private Key", "", true)
}

func promptString(label, defaultValue string, required bool) (string, error) {
	prompt := promptui.Prompt{
		Label:   label,
		Default: defaultValue,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	if required && strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("%s is required", label)
	}

	return result, nil
}

func promptInt(label string, defaultValue int) (int, error) {
	prompt := promptui.Prompt{
		Label:   label,
		Default: fmt.Sprintf("%d", defaultValue),
		Validate: func(input string) error {
			_, err := fmt.Sscanf(input, "%d", new(int))
			if err != nil {
				return fmt.Errorf("invalid number")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	var value int
	fmt.Sscanf(result, "%d", &value)
	return value, nil
}

// generateKeyPair generates a new WireGuard private/public key pair
func generateKeyPair() (privateKey, publicKey string, err error) {
	// Generate private key
	cmd := exec.Command("wg", "genkey")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key (is 'wg' installed?): %v", err)
	}
	privateKey = strings.TrimSpace(string(output))

	// Generate public key from private key
	cmd = exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	output, err = cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %v", err)
	}
	publicKey = strings.TrimSpace(string(output))

	return privateKey, publicKey, nil
}
