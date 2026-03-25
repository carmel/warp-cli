# 自定义 WireGuard 服务器支持分析

## 当前状态

### ❌ 不支持自定义 WireGuard 服务器

当前 warp-cli 项目**不支持**连接到自己部署的 WireGuard 服务器。

### 原因分析

#### 1. 配置来源固定

查看 `cmd/generate/generate.go` 和 `cmd/connect/connect.go`：

```go
// generate.go
func generateProfile() error {
    // 从 Cloudflare API 获取配置
    thisDevice, err := cloudflare.GetSourceDevice(ctx)

    // 使用 Cloudflare 返回的配置
    profile, err := util.NewProfile(&util.ProfileData{
        PrivateKey: viper.GetString(config.PrivateKey),
        Address1:   thisDevice.Config.Interface.Addresses.V4,  // Cloudflare 分配
        Address2:   thisDevice.Config.Interface.Addresses.V6,  // Cloudflare 分配
        PublicKey:  thisDevice.Config.Peers[0].PublicKey,      // Cloudflare 服务器
        Endpoint:   thisDevice.Config.Peers[0].Endpoint.Host,  // Cloudflare 端点
    })
}
```

**问题：**

- 所有配置都来自 Cloudflare API
- 无法指定自定义服务器
- 无法使用自己的 WireGuard 配置文件

#### 2. 注册流程绑定

```go
// register.go
func registerAccount() error {
    // 必须向 Cloudflare API 注册
    device, err := cloudflare.Register(privateKey.Public(), deviceModel)

    // 获取 Cloudflare 分配的配置
    viper.Set(config.DeviceId, device.Id)
    viper.Set(config.AccessToken, device.Token)
}
```

**问题：**

- 必须先注册到 Cloudflare
- 无法跳过注册直接使用自定义配置

#### 3. Connect 命令依赖

```go
// connect.go
func connectVPN() error {
    // 检查是否有 Cloudflare 账户
    if !config.IsAccountValid() {
        return errors.New("no valid account found")
    }

    // 生成配置（从 Cloudflare API）
    if err := generateProfile(); err != nil {
        return err
    }
}
```

**问题：**

- 必须有有效的 Cloudflare 账户
- 无法使用外部配置文件

---

## 对比：标准 WireGuard 工具

### 标准 WireGuard 使用方式

```bash
# 1. 创建配置文件
cat > wg0.conf <<EOF
[Interface]
PrivateKey = <your-private-key>
Address = 10.0.0.2/24
DNS = 8.8.8.8

[Peer]
PublicKey = <server-public-key>
Endpoint = your-server.com:51820
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
EOF

# 2. 启动连接
wg-quick up wg0

# 或使用 wireguard-go
wireguard-go wg0
wg setconf wg0 wg0.conf
```

### warp-cli 当前方式

```bash
# 1. 必须注册到 Cloudflare
warp-cli register

# 2. 自动生成配置（从 Cloudflare）
warp-cli generate

# 3. 连接（只能连接到 Cloudflare）
warp-cli connect
```

---

## 如何使用自定义 WireGuard 服务器

### 方案 1：直接使用 wireguard-go（推荐）

warp-cli 项目包含了完整的 wireguard-go 实现，你可以直接使用它：

```bash
# 1. 编译 wireguard-go
cd wireguard
make

# 2. 创建你的配置文件
cat > my-server.conf <<EOF
[Interface]
PrivateKey = <your-private-key>
Address = 10.0.0.2/24
DNS = 8.8.8.8

[Peer]
PublicKey = <your-server-public-key>
Endpoint = your-server.com:51820
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
EOF

# 3. 启动 wireguard-go
./wireguard-go wg0

# 4. 应用配置
wg setconf wg0 my-server.conf

# 5. 配置路由（如果需要）
ip addr add 10.0.0.2/24 dev wg0
ip link set wg0 up
ip route add 0.0.0.0/0 dev wg0
```

### 方案 2：使用系统的 WireGuard

```bash
# Linux (使用内核模块)
sudo wg-quick up my-server

# macOS
brew install wireguard-tools
sudo wg-quick up my-server

# Windows
# 使用 WireGuard GUI 客户端
```

### 方案 3：修改 warp-cli 支持自定义配置

如果你想让 warp-cli 支持自定义 WireGuard 服务器，需要添加新功能。

---

## 实现自定义 WireGuard 支持

### 需要的改动

#### 1. 添加 `import` 命令

```go
// cmd/import/import.go
package import

import (
    "fmt"
    "log"
    "os"

    "github.com/carmel/warp-cli/util"
    "github.com/spf13/cobra"
)

var (
    configFile string
    shortMsg   = "Import a custom WireGuard configuration"
)

var Cmd = &cobra.Command{
    Use:   "import",
    Short: shortMsg,
    Long: `Import a custom WireGuard configuration file.

This allows you to use warp-cli with your own WireGuard server
instead of Cloudflare WARP.

Example:
  warp-cli import --config my-server.conf`,
    Run: func(cmd *cobra.Command, args []string) {
        util.RunCommandFatal(importConfig)
    },
}

func init() {
    Cmd.Flags().StringVarP(&configFile, "config", "c", "", "WireGuard config file to import")
    Cmd.MarkFlagRequired("config")
}

func importConfig() error {
    // 读取配置文件
    data, err := os.ReadFile(configFile)
    if err != nil {
        return fmt.Errorf("failed to read config: %v", err)
    }

    // 验证配置格式
    if err := validateWireGuardConfig(string(data)); err != nil {
        return fmt.Errorf("invalid config: %v", err)
    }

    // 复制到 warp-cli-profile.conf
    if err := os.WriteFile("warp-cli-profile.conf", data, 0600); err != nil {
        return fmt.Errorf("failed to save config: %v", err)
    }

    // 创建一个标记文件，表示使用自定义配置
    if err := os.WriteFile("warp-cli-custom.flag", []byte("custom"), 0600); err != nil {
        return fmt.Errorf("failed to create flag: %v", err)
    }

    log.Println("✓ Successfully imported custom WireGuard configuration!")
    log.Println("  Config file: warp-cli-profile.conf")
    log.Println("\nYou can now connect using: warp-cli connect")

    return nil
}

func validateWireGuardConfig(config string) error {
    // 简单验证：检查必需的字段
    required := []string{"[Interface]", "PrivateKey", "[Peer]", "PublicKey", "Endpoint"}
    for _, field := range required {
        if !strings.Contains(config, field) {
            return fmt.Errorf("missing required field: %s", field)
        }
    }
    return nil
}
```

#### 2. 修改 `connect` 命令

```go
// cmd/connect/connect.go
func connectVPN() error {
    // 检查是否使用自定义配置
    if isCustomConfig() {
        log.Println("Using custom WireGuard configuration")
        return connectCustom()
    }

    // 原有的 Cloudflare 逻辑
    if !config.IsAccountValid() {
        return errors.New("no valid account found. Please run 'warp-cli register' or 'warp-cli import' first")
    }

    // ... 原有代码
}

func isCustomConfig() bool {
    _, err := os.Stat("warp-cli-custom.flag")
    return err == nil
}

func connectCustom() error {
    // 检查配置文件是否存在
    if _, err := os.Stat(profileFile); os.IsNotExist(err) {
        return fmt.Errorf("profile file not found. Please run 'warp-cli import' first")
    }

    // 启动 wireguard-go
    wireguardPath, err := findWireguardGo()
    if err != nil {
        return fmt.Errorf("wireguard not found: %v", err)
    }

    if err := startWireguardGo(wireguardPath); err != nil {
        return fmt.Errorf("failed to start wireguard-go: %v", err)
    }

    // 应用配置
    if err := applyConfig(); err != nil {
        return fmt.Errorf("failed to apply configuration: %v", err)
    }

    log.Println("✓ Successfully connected to custom WireGuard server!")
    return nil
}
```

#### 3. 添加 `connect-custom` 命令（更简单的方式）

```go
// cmd/connect_custom/connect_custom.go
package connect_custom

import (
    "fmt"
    "log"
    "os"
    "os/exec"

    "github.com/carmel/warp-cli/util"
    "github.com/spf13/cobra"
)

var (
    configFile    string
    interfaceName string
    shortMsg      = "Connect to a custom WireGuard server"
)

var Cmd = &cobra.Command{
    Use:   "connect-custom",
    Short: shortMsg,
    Long: `Connect to your own WireGuard server using a config file.

This bypasses Cloudflare WARP and connects directly to your server.

Example:
  warp-cli connect-custom --config my-server.conf
  warp-cli connect-custom --config my-server.conf --interface wg1`,
    Run: func(cmd *cobra.Command, args []string) {
        util.RunCommandFatal(connectCustom)
    },
}

func init() {
    Cmd.Flags().StringVarP(&configFile, "config", "c", "", "WireGuard config file")
    Cmd.Flags().StringVarP(&interfaceName, "interface", "i", "wg0", "Interface name")
    Cmd.MarkFlagRequired("config")
}

func connectCustom() error {
    // 检查配置文件
    if _, err := os.Stat(configFile); os.IsNotExist(err) {
        return fmt.Errorf("config file not found: %s", configFile)
    }

    log.Printf("Connecting to custom WireGuard server...")
    log.Printf("  Config: %s", configFile)
    log.Printf("  Interface: %s", interfaceName)

    // 查找 wireguard-go
    wireguardPath, err := findWireguardGo()
    if err != nil {
        return fmt.Errorf("wireguard not found: %v", err)
    }

    // 启动 wireguard-go
    cmd := exec.Command(wireguardPath, interfaceName)
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start wireguard-go: %v", err)
    }

    // 等待接口创建
    time.Sleep(1 * time.Second)

    // 应用配置
    cmd = exec.Command("wg", "setconf", interfaceName, configFile)
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("failed to apply config: %v\n%s", err, output)
    }

    log.Println("✓ Successfully connected to custom WireGuard server!")
    log.Printf("  Interface: %s", interfaceName)
    log.Printf("\nTo disconnect, run: warp-cli disconnect -i %s", interfaceName)

    return nil
}

func findWireguardGo() (string, error) {
    // 同 connect.go 的实现
    // ...
}
```

---

## 使用示例

### 场景 1：使用项目自带的 wireguard-go

```bash
# 1. 编译 wireguard-go
cd wireguard && make && cd ..

# 2. 创建配置文件
cat > my-vpn.conf <<EOF
[Interface]
PrivateKey = YourPrivateKeyHere
Address = 10.0.0.2/24

[Peer]
PublicKey = ServerPublicKeyHere
Endpoint = vpn.example.com:51820
AllowedIPs = 0.0.0.0/0
EOF

# 3. 使用 wireguard-go 连接
./wireguard/wireguard-go wg0
wg setconf wg0 my-vpn.conf

# 4. 断开
wg-quick down wg0
# 或
ip link del wg0
```

### 场景 2：如果实现了 connect-custom 命令

```bash
# 1. 编译 warp-cli（包含新功能）
go build -o warp-cli .

# 2. 连接到自定义服务器
warp-cli connect-custom --config my-vpn.conf

# 3. 断开
warp-cli disconnect
```

---

## 实现计划

### 阶段 1：最小实现（1-2 天）

添加 `connect-custom` 命令：

- [x] 读取自定义配置文件
- [x] 启动 wireguard-go
- [x] 应用配置
- [x] 基本错误处理

### 阶段 2：完整实现（2-3 天）

添加 `import` 命令和增强功能：

- [ ] 配置文件验证
- [ ] 配置文件管理
- [ ] 与现有命令集成
- [ ] 自动检测配置类型

### 阶段 3：高级功能（3-5 天）

- [ ] 多配置文件管理
- [ ] 配置文件编辑器
- [ ] 配置文件模板
- [ ] 服务器连接测试

---

## 对比总结

| 功能            | 当前 warp-cli | 添加自定义支持后 | 标准 WireGuard |
| --------------- | ------------- | ---------------- | -------------- |
| Cloudflare WARP | ✅            | ✅               | ❌             |
| Zero Trust      | ✅            | ✅               | ❌             |
| 自定义服务器    | ❌            | ✅               | ✅             |
| 配置管理        | 自动          | 自动 + 手动      | 手动           |
| 易用性          | 高            | 高               | 中             |

---

## 推荐方案

### 如果你只需要连接自定义 WireGuard 服务器

**推荐：直接使用项目自带的 wireguard-go**

```bash
cd wireguard && make
./wireguard-go wg0
wg setconf wg0 your-config.conf
```

**优点：**

- 无需修改代码
- 立即可用
- 完全兼容标准 WireGuard

### 如果你想要统一的命令行体验

**推荐：实现 `connect-custom` 命令**

这样你可以：

- 使用 `warp-cli connect` 连接 Cloudflare
- 使用 `warp-cli connect-custom` 连接自定义服务器
- 统一的命令行接口

---

## 结论

### 当前状态

- ❌ warp-cli **不支持**连接自定义 WireGuard 服务器
- ✅ 项目包含完整的 wireguard-go 实现
- ✅ 可以直接使用 wireguard-go 连接任何 WireGuard 服务器

### 解决方案

1. **立即可用**：直接使用 `wireguard/wireguard-go`
2. **未来增强**：添加 `connect-custom` 命令
3. **完整集成**：实现配置管理系统

### 工作量估算

- 最小实现（connect-custom）：1-2 天
- 完整实现（import + 管理）：3-5 天
- 高级功能（多配置、模板）：5-7 天

---

## 快速开始（使用 wireguard-go）

```bash
# 1. 编译
cd wireguard
make

# 2. 创建配置
cat > my-server.conf <<'EOF'
[Interface]
PrivateKey = <your-private-key>
Address = 10.0.0.2/24
DNS = 8.8.8.8

[Peer]
PublicKey = <server-public-key>
Endpoint = your-server.com:51820
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
EOF

# 3. 连接
sudo ./wireguard-go wg0
sudo wg setconf wg0 my-server.conf

# 4. 验证
wg show wg0

# 5. 断开
sudo ip link del wg0
```

这样你就可以使用项目自带的 wireguard-go 连接任何 WireGuard 服务器了！
