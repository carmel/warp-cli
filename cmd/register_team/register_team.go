package register_team

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/carmel/warp-cli/cloudflare"
	"github.com/carmel/warp-cli/config"
	"github.com/carmel/warp-cli/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	teamToken    string
	teamName     string
	clientID     string
	clientSecret string
	shortMsg     = "Register device with Cloudflare Zero Trust"
)

var Cmd = &cobra.Command{
	Use:   "register-team",
	Short: shortMsg,
	Long: util.FormatMessage(shortMsg, `
Register your device to a Cloudflare Zero Trust organization.

Three methods are supported:

1. Token URL (recommended for manual registration):
   warp-cli register-team --token "com.cloudflare.warp://myteam.cloudflareaccess.com/auth?token=xxx"

2. Service Token (recommended for automation):
   warp-cli register-team --team mycompany --client-id xxx --client-secret yyy

3. Team name only (generates URL for browser authentication):
   warp-cli register-team --team mycompany

Examples:
  # Method 1: Using token URL (from Cloudflare dashboard)
  warp-cli register-team --token "com.cloudflare.warp://myteam.cloudflareaccess.com/auth?token=abc123"

  # Method 2: Using service token
  warp-cli register-team --team mycompany --client-id xxx --client-secret yyy

  # Method 3: Interactive (generates URL)
  warp-cli register-team --team mycompany
  # Then visit the URL in a browser to authenticate`),
	Run: func(cmd *cobra.Command, args []string) {
		util.RunCommandFatal(registerTeam)
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&teamToken, "token", "t", "", "Team enrollment token")
	Cmd.PersistentFlags().StringVar(&teamName, "team", "", "Team name")
	Cmd.PersistentFlags().StringVar(&clientID, "client-id", "", "Service token client ID")
	Cmd.PersistentFlags().StringVar(&clientSecret, "client-secret", "", "Service token client secret")
}

func registerTeam() error {
	// 检查是否已有账户
	if config.IsAccountValid() {
		return fmt.Errorf("an account already exists. Please remove the config file first")
	}

	// 方法 1: Token URL
	if teamToken != "" {
		return registerWithToken()
	}

	// 方法 2: Service Token
	if clientID != "" && clientSecret != "" {
		if teamName == "" {
			return fmt.Errorf("--team is required when using service token")
		}
		return registerWithServiceToken()
	}

	// 方法 3: Interactive (生成 URL)
	if teamName != "" {
		return registerInteractive()
	}

	return fmt.Errorf("please specify either --token, --team with service credentials, or --team for interactive mode")
}

// registerWithToken 使用 token URL 注册
func registerWithToken() error {
	// 解析 token URL
	// 支持两种格式:
	// 1. com.cloudflare.warp://team.cloudflareaccess.com/auth?token=xxx
	// 2. https://team.cloudflareaccess.com/auth?token=xxx

	// 替换自定义协议为 https 以便解析
	urlStr := strings.Replace(teamToken, "com.cloudflare.warp://", "https://", 1)

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid token URL: %v", err)
	}

	// 提取团队名称（从 hostname 中提取）
	// 例如: carmeltop.cloudflareaccess.com -> carmeltop
	hostname := parsedURL.Hostname()
	if !strings.HasSuffix(hostname, ".cloudflareaccess.com") {
		return fmt.Errorf("invalid token URL: hostname must end with .cloudflareaccess.com")
	}
	extractedTeamName := strings.TrimSuffix(hostname, ".cloudflareaccess.com")

	// 提取 token 参数
	token := parsedURL.Query().Get("token")
	if token == "" {
		return fmt.Errorf("invalid token URL: missing token parameter")
	}

	log.Printf("Registering with team: %s", extractedTeamName)

	// 生成密钥对
	privateKey, err := util.NewPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}

	// 注册设备
	device, err := cloudflare.RegisterTeamDevice(extractedTeamName, token, privateKey.Public())
	if err != nil {
		return fmt.Errorf("registration failed: %v", err)
	}

	// 保存配置
	return saveTeamConfig(extractedTeamName, device, privateKey)
}

// registerWithServiceToken 使用 service token 注册
func registerWithServiceToken() error {
	log.Printf("Registering with service token for team: %s", teamName)

	// 生成密钥对
	privateKey, err := util.NewPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}

	// 使用 Service Token 注册
	device, err := cloudflare.RegisterTeamDeviceWithServiceToken(
		teamName,
		clientID,
		clientSecret,
		privateKey.Public(),
	)
	if err != nil {
		return fmt.Errorf("registration failed: %v", err)
	}

	log.Println("\nNote: This device is enrolled as non_identity@" + teamName + ".cloudflareaccess.com")

	// 保存配置
	return saveTeamConfig(teamName, device, privateKey)
}

// registerInteractive 交互式注册（生成 URL）
func registerInteractive() error {
	log.Printf("Generating registration URL for team: %s", teamName)

	// 生成密钥对
	privateKey, err := util.NewPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}

	// 生成注册 URL
	registrationURL := fmt.Sprintf(
		"https://%s.cloudflareaccess.com/warp?pub=%s",
		teamName,
		url.QueryEscape(privateKey.Public().String()),
	)

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("Please visit the following URL to authenticate:")
	fmt.Println()
	fmt.Println("  " + registrationURL)
	fmt.Println()
	fmt.Println("After authentication, you will receive a token URL.")
	fmt.Println("Run the following command with your token:")
	fmt.Println()
	fmt.Println("  warp-cli register-team --token \"<your-token-url>\"")
	fmt.Println(strings.Repeat("=", 70))

	// 保存私钥以便后续使用
	log.Println("\nPrivate key generated and ready for use.")
	log.Println("Please complete authentication in your browser.")

	return nil
}

// saveTeamConfig 保存 Zero Trust 配置
func saveTeamConfig(team string, device *cloudflare.TeamDevice, privateKey *util.Key) error {
	viper.Set(config.Mode, "team")
	viper.Set(config.TeamName, team)
	viper.Set(config.PrivateKey, privateKey.String())
	viper.Set(config.DeviceId, device.ID)
	viper.Set(config.AccessToken, device.Token)

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	log.Println("\n✓ Successfully registered with Cloudflare Zero Trust!")
	log.Printf("  Team: %s", team)
	log.Printf("  Device ID: %s", device.ID)
	log.Println("\nYou can now connect using: warp-cli connect")

	return nil
}
