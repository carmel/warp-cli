# Zero Trust 支持实现方案

## 概述

基于对 Cloudflare 官方 warp-cli 和 Zero Trust API 的研究，本文档提供了为 warp-cli 添加 Zero Trust 支持的完整实现方案。

## 核心发现

### 官方 warp-cli 的 Zero Trust 命令

Cloudflare 官方的 warp-cli（Linux 版本）已经支持 Zero Trust，主要命令：

```bash
# 方法 1：交互式注册（需要浏览器）
warp-cli registration new <team-name>
# 生成一个 URL，用户在浏览器中访问并认证
# 认证后获得 token

# 方法 2：使用 token 注册
warp-cli teams-enroll-token <token-url>
# 例如：warp-cli teams-enroll-token https://myteam.cloudflareaccess.com/auth?token=xxx

# 方法 3：使用 Service Token（无人值守）
warp-cli teams-enroll --access-client-id <id> --access-client-secret <secret>
```

### Zero Trust 认证流程

```
1. 用户发起注册
   ↓
2. 生成设备注册 URL
   https://<team>.cloudflareaccess.com/warp
   ↓
3. 用户在浏览器中访问 URL
   ↓
4. 通过 IdP 认证（Email OTP、SSO、SAML 等）
   ↓
5. 获得注册 token
   ↓
6. 使用 token 完成设备注册
   ↓
7. 获取 WireGuard 配置
```

---

## 实现方案

### 方案 A：完整实现（推荐）

实现三种注册方式，与官方 warp-cli 功能对等。


#### 1. 交互式注册（Interactive Registration）

**命令：**
```bash
warp-cli register --team <team-name>
# 或
warp-cli register-team <team-name>
```

**实现步骤：**

```go
// cmd/register/register_team.go
func registerTeam(teamName string) error {
    // 1. 生成设备密钥对
    privateKey, _ := util.NewPrivateKey()
    
    // 2. 生成注册 URL
    registrationURL := fmt.Sprintf(
        "https://%s.cloudflareaccess.com/warp?pub=%s",
        teamName,
        url.QueryEscape(privateKey.Public().String()),
    )
    
    // 3. 显示 URL 给用户
    fmt.Println("Please visit the following URL to authenticate:")
    fmt.Println(registrationURL)
    fmt.Println("\nWaiting for authentication...")
    
    // 4. 启动本地 HTTP 服务器接收回调
    token := waitForCallback()
    
    // 5. 使用 token 完成注册
    return completeTeamRegistration(teamName, token, privateKey)
}
```

**复杂度：** 中等
**工作量：** 2-3 天

#### 2. Token 注册（Token-based Registration）

**命令：**
```bash
warp-cli register --team-token <token-url>
```

**实现步骤：**

```go
// cmd/register/register_team_token.go
func registerWithTeamToken(tokenURL string) error {
    // 1. 解析 token URL
    parsedURL, _ := url.Parse(tokenURL)
    token := parsedURL.Query().Get("token")
    team := extractTeamFromURL(tokenURL)
    
    // 2. 生成设备密钥对
    privateKey, _ := util.NewPrivateKey()
    
    // 3. 调用 Zero Trust API 注册
    device, err := cloudflare.RegisterTeamDevice(team, token, privateKey.Public())
    
    // 4. 保存配置
    saveTeamConfig(device, privateKey)
    
    return nil
}
```

**复杂度：** 低
**工作量：** 1-2 天

#### 3. Service Token 注册（Headless/Automated）

**命令：**
```bash
warp-cli register --team <team-name> \
  --client-id <id> \
  --client-secret <secret>
```

**实现步骤：**

```go
// cmd/register/register_service_token.go
func registerWithServiceToken(teamName, clientID, clientSecret string) error {
    // 1. 生成设备密钥对
    privateKey, _ := util.NewPrivateKey()
    
    // 2. 使用 Service Token 认证
    device, err := cloudflare.RegisterTeamDeviceWithServiceToken(
        teamName,
        clientID,
        clientSecret,
        privateKey.Public(),
    )
    
    // 3. 保存配置
    saveTeamConfig(device, privateKey)
    
    return nil
}
```

**复杂度：** 低
**工作量：** 1 天

---

### 方案 B：简化实现（快速上线）

只实现 Token 注册和 Service Token 注册，不实现交互式注册。

**优点：**
- 实现简单
- 适合自动化场景
- 无需处理浏览器回调

**缺点：**
- 用户体验不如交互式
- 需要手动获取 token

**工作量：** 2-3 天

---

## 详细技术实现

### 1. API 端点分析

#### Zero Trust 注册 API

```
端点：https://<team>.cloudflareaccess.com/warp
方法：POST
认证：Bearer Token（从浏览器认证获得）

请求体：
{
  "key": "WireGuard 公钥",
  "install_id": "",
  "fcm_token": "",
  "tos": "时间戳",
  "type": "Linux",
  "model": "PC",
  "locale": "en_US"
}

响应：
{
  "id": "设备ID",
  "token": "访问令牌",
  "account": {...},
  "config": {
    "peers": [...],
    "interface": {...}
  }
}
```

#### Service Token 认证

```
端点：https://<team>.cloudflareaccess.com/warp
方法：POST
Headers：
  CF-Access-Client-Id: <client-id>
  CF-Access-Client-Secret: <client-secret>

请求体：同上
```

### 2. 代码结构

```
cmd/
├── register/
│   ├── register.go              # 现有的消费者注册
│   ├── register_team.go         # Zero Trust 交互式注册
│   ├── register_team_token.go   # Zero Trust Token 注册
│   └── register_service.go      # Service Token 注册
│
cloudflare/
├── api.go                       # 现有的消费者 API
├── api_team.go                  # Zero Trust API
└── auth.go                      # 认证相关
│
config/
├── config.go                    # 配置管理
└── team_config.go               # Zero Trust 配置
│
util/
├── oauth.go                     # OAuth/回调处理
└── browser.go                   # 浏览器打开
```

### 3. 配置文件格式

#### 消费者 WARP 配置（现有）
```toml
# warp-cli-account.toml
device_id = "xxx"
access_token = "xxx"
private_key = "xxx"
license_key = "xxx"
```

#### Zero Trust 配置（新增）
```toml
# warp-cli-team-account.toml
mode = "team"
team_name = "mycompany"
device_id = "xxx"
access_token = "xxx"
private_key = "xxx"
auth_domain = "mycompany.cloudflareaccess.com"
```

### 4. 核心实现代码

#### cloudflare/api_team.go

```go
package cloudflare

import (
    "context"
    "fmt"
    "net/http"
)

const (
    TeamAuthDomain = "cloudflareaccess.com"
)

// RegisterTeamDevice 使用 token 注册 Zero Trust 设备
func RegisterTeamDevice(teamName, token string, publicKey *util.Key) (*TeamDevice, error) {
    url := fmt.Sprintf("https://%s.%s/warp", teamName, TeamAuthDomain)
    
    client := &http.Client{}
    req, _ := http.NewRequest("POST", url, buildRequestBody(publicKey))
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    
    // 解析响应
    var device TeamDevice
    json.NewDecoder(resp.Body).Decode(&device)
    
    return &device, nil
}

// RegisterTeamDeviceWithServiceToken 使用 Service Token 注册
func RegisterTeamDeviceWithServiceToken(
    teamName, clientID, clientSecret string,
    publicKey *util.Key,
) (*TeamDevice, error) {
    url := fmt.Sprintf("https://%s.%s/warp", teamName, TeamAuthDomain)
    
    client := &http.Client{}
    req, _ := http.NewRequest("POST", url, buildRequestBody(publicKey))
    req.Header.Set("CF-Access-Client-Id", clientID)
    req.Header.Set("CF-Access-Client-Secret", clientSecret)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    
    var device TeamDevice
    json.NewDecoder(resp.Body).Decode(&device)
    
    return &device, nil
}

type TeamDevice struct {
    ID     string `json:"id"`
    Token  string `json:"token"`
    Config struct {
        Peers []struct {
            PublicKey string `json:"public_key"`
            Endpoint  struct {
                Host string `json:"host"`
            } `json:"endpoint"`
        } `json:"peers"`
        Interface struct {
            Addresses struct {
                V4 string `json:"v4"`
                V6 string `json:"v6"`
            } `json:"addresses"`
        } `json:"interface"`
    } `json:"config"`
}
```

#### cmd/register/register_team_token.go

```go
package register

import (
    "fmt"
    "log"
    "net/url"
    
    "github.com/carmel/warp-cli/cloudflare"
    "github.com/carmel/warp-cli/config"
    "github.com/carmel/warp-cli/util"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var teamToken string

var TeamTokenCmd = &cobra.Command{
    Use:   "register-team-token",
    Short: "Register device with Zero Trust using a token",
    Long: `Register your device to a Cloudflare Zero Trust organization using a token URL.
    
Example:
  warp-cli register-team-token "https://myteam.cloudflareaccess.com/auth?token=xxx"`,
    Run: func(cmd *cobra.Command, args []string) {
        if len(args) > 0 {
            teamToken = args[0]
        }
        util.RunCommandFatal(registerWithTeamToken)
    },
}

func init() {
    TeamTokenCmd.Flags().StringVarP(&teamToken, "token", "t", "", "Team enrollment token URL")
}

func registerWithTeamToken() error {
    if teamToken == "" {
        return fmt.Errorf("team token URL is required")
    }
    
    // 解析 token URL
    parsedURL, err := url.Parse(teamToken)
    if err != nil {
        return fmt.Errorf("invalid token URL: %v", err)
    }
    
    token := parsedURL.Query().Get("token")
    if token == "" {
        return fmt.Errorf("no token found in URL")
    }
    
    // 提取 team name
    teamName := extractTeamName(parsedURL.Host)
    if teamName == "" {
        return fmt.Errorf("could not extract team name from URL")
    }
    
    log.Printf("Registering device with team: %s", teamName)
    
    // 生成密钥对
    privateKey, err := util.NewPrivateKey()
    if err != nil {
        return fmt.Errorf("failed to generate key: %v", err)
    }
    
    // 注册设备
    device, err := cloudflare.RegisterTeamDevice(teamName, token, privateKey.Public())
    if err != nil {
        return fmt.Errorf("registration failed: %v", err)
    }
    
    // 保存配置
    viper.Set(config.Mode, "team")
    viper.Set(config.TeamName, teamName)
    viper.Set(config.PrivateKey, privateKey.String())
    viper.Set(config.DeviceId, device.ID)
    viper.Set(config.AccessToken, device.Token)
    
    if err := viper.WriteConfig(); err != nil {
        return fmt.Errorf("failed to save config: %v", err)
    }
    
    log.Println("✓ Successfully registered with Zero Trust!")
    log.Printf("  Team: %s", teamName)
    log.Printf("  Device ID: %s", device.ID)
    
    return nil
}

func extractTeamName(host string) string {
    // 从 "myteam.cloudflareaccess.com" 提取 "myteam"
    parts := strings.Split(host, ".")
    if len(parts) > 0 {
        return parts[0]
    }
    return ""
}
```

#### cmd/register/register_service.go

```go
package register

import (
    "fmt"
    "log"
    
    "github.com/carmel/warp-cli/cloudflare"
    "github.com/carmel/warp-cli/config"
    "github.com/carmel/warp-cli/util"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var (
    serviceTeamName    string
    serviceClientID    string
    serviceClientSecret string
)

var ServiceTokenCmd = &cobra.Command{
    Use:   "register-service",
    Short: "Register device with Zero Trust using a service token",
    Long: `Register your device to a Cloudflare Zero Trust organization using a service token.
This is useful for headless servers and automated deployments.

Example:
  warp-cli register-service --team mycompany \
    --client-id xxx --client-secret yyy`,
    Run: func(cmd *cobra.Command, args []string) {
        util.RunCommandFatal(registerWithServiceToken)
    },
}

func init() {
    ServiceTokenCmd.Flags().StringVar(&serviceTeamName, "team", "", "Team name (required)")
    ServiceTokenCmd.Flags().StringVar(&serviceClientID, "client-id", "", "Service token client ID (required)")
    ServiceTokenCmd.Flags().StringVar(&serviceClientSecret, "client-secret", "", "Service token client secret (required)")
    ServiceTokenCmd.MarkFlagRequired("team")
    ServiceTokenCmd.MarkFlagRequired("client-id")
    ServiceTokenCmd.MarkFlagRequired("client-secret")
}

func registerWithServiceToken() error {
    log.Printf("Registering device with team: %s (using service token)", serviceTeamName)
    
    // 生成密钥对
    privateKey, err := util.NewPrivateKey()
    if err != nil {
        return fmt.Errorf("failed to generate key: %v", err)
    }
    
    // 使用 Service Token 注册
    device, err := cloudflare.RegisterTeamDeviceWithServiceToken(
        serviceTeamName,
        serviceClientID,
        serviceClientSecret,
        privateKey.Public(),
    )
    if err != nil {
        return fmt.Errorf("registration failed: %v", err)
    }
    
    // 保存配置
    viper.Set(config.Mode, "team")
    viper.Set(config.TeamName, serviceTeamName)
    viper.Set(config.PrivateKey, privateKey.String())
    viper.Set(config.DeviceId, device.ID)
    viper.Set(config.AccessToken, device.Token)
    
    if err := viper.WriteConfig(); err != nil {
        return fmt.Errorf("failed to save config: %v", err)
    }
    
    log.Println("✓ Successfully registered with Zero Trust!")
    log.Printf("  Team: %s", serviceTeamName)
    log.Printf("  Device ID: %s", device.ID)
    log.Println("\nNote: This device is enrolled as non_identity@<team>.cloudflareaccess.com")
    
    return nil
}
```

---

## 实施计划

### 阶段 1：基础架构（1-2 天）

- [ ] 创建 `cloudflare/api_team.go`
- [ ] 添加 Zero Trust API 客户端
- [ ] 实现基本的 HTTP 请求封装
- [ ] 添加配置文件支持（team mode）

### 阶段 2：Token 注册（1-2 天）

- [ ] 实现 `register-team-token` 命令
- [ ] URL 解析和验证
- [ ] Token 提取逻辑
- [ ] 设备注册流程
- [ ] 配置保存

### 阶段 3：Service Token 注册（1 天）

- [ ] 实现 `register-service` 命令
- [ ] Service Token 认证
- [ ] 无人值守注册流程

### 阶段 4：交互式注册（2-3 天，可选）

- [ ] 实现 `register-team` 命令
- [ ] 生成注册 URL
- [ ] 本地 HTTP 服务器（接收回调）
- [ ] 浏览器自动打开
- [ ] Token 接收和处理

### 阶段 5：集成和测试（2-3 天）

- [ ] 更新 `connect` 命令支持 team mode
- [ ] 更新 `generate` 命令
- [ ] 更新 `status` 命令
- [ ] 编写测试用例
- [ ] 文档更新

### 阶段 6：文档和示例（1-2 天）

- [ ] 使用指南
- [ ] API 文档
- [ ] 示例脚本
- [ ] 故障排除

**总工期：** 8-13 天（取决于是否实现交互式注册）

---

## 使用示例

### 场景 1：个人用户（交互式）

```bash
# 1. 注册到 Zero Trust
warp-cli register-team mycompany
# 输出：Please visit: https://mycompany.cloudflareaccess.com/warp?pub=xxx
# 用户在浏览器中认证

# 2. 连接
warp-cli connect

# 3. 验证
warp-cli trace
```

### 场景 2：使用 Token

```bash
# 1. 在浏览器中访问
# https://mycompany.cloudflareaccess.com/warp
# 认证后复制 token URL

# 2. 使用 token 注册
warp-cli register-team-token "https://mycompany.cloudflareaccess.com/auth?token=xxx"

# 3. 连接
warp-cli connect
```

### 场景 3：服务器自动化部署

```bash
# 1. 管理员创建 Service Token
# 在 Cloudflare Dashboard 中创建

# 2. 在服务器上注册
warp-cli register-service \
  --team mycompany \
  --client-id "xxx" \
  --client-secret "yyy"

# 3. 连接
warp-cli connect

# 4. 设置开机自启
systemctl enable warp-cli-connect
```

---

## 兼容性考虑

### 向后兼容

- 保持现有的 `register` 命令不变（消费者 WARP）
- 新增独立的 Zero Trust 命令
- 配置文件通过 `mode` 字段区分

### 命令对比

| 功能 | 消费者 WARP | Zero Trust |
|------|------------|------------|
| 注册 | `register` | `register-team-token` |
| 连接 | `connect` | `connect`（自动检测 mode） |
| 状态 | `status` | `status`（显示 team 信息） |
| 配置 | `generate` | `generate`（自动检测 mode） |

---

## 潜在问题和解决方案

### 问题 1：Token 过期

**问题：** 注册 token 可能有时效性

**解决方案：**
- 检测 token 过期错误
- 提示用户重新获取 token
- 实现 token 刷新机制（如果 API 支持）

### 问题 2：多种认证方式

**问题：** Zero Trust 支持多种 IdP（Email OTP、SSO、SAML 等）

**解决方案：**
- 使用通用的 token-based 流程
- 认证由浏览器处理
- CLI 只负责接收最终的 token

### 问题 3：设备策略

**问题：** 管理员可能配置了设备策略（如 mTLS、设备态检查）

**解决方案：**
- 在文档中说明策略要求
- 提供错误信息指导
- 支持 mTLS 证书配置（高级功能）

### 问题 4：网络隔离环境

**问题：** 某些服务器可能无法访问外网

**解决方案：**
- 支持代理配置
- 提供离线注册方式（通过配置文件导入）

---

## 测试计划

### 单元测试

```go
// cloudflare/api_team_test.go
func TestRegisterTeamDevice(t *testing.T) {
    // 测试 token 注册
}

func TestRegisterTeamDeviceWithServiceToken(t *testing.T) {
    // 测试 service token 注册
}

func TestExtractTeamName(t *testing.T) {
    // 测试 team name 提取
}
```

### 集成测试

1. 创建测试 Zero Trust 组织
2. 生成测试 Service Token
3. 自动化注册流程测试
4. 连接和断开测试

### 手动测试清单

- [ ] Token 注册流程
- [ ] Service Token 注册流程
- [ ] 交互式注册流程（如果实现）
- [ ] 连接到 Zero Trust 网络
- [ ] 访问内网资源
- [ ] 设备策略验证
- [ ] 错误处理和提示

---

## 文档更新

### 需要更新的文档

1. **README.md**
   - 添加 Zero Trust 支持说明
   - 更新功能列表

2. **ZERO_TRUST_GUIDE.md**（新建）
   - 详细的 Zero Trust 使用指南
   - 三种注册方式的说明
   - 常见问题解答

3. **DOCUMENTATION_INDEX.md**
   - 添加 Zero Trust 相关文档索引

4. **WARP_VS_ZERO_TRUST.md**
   - 更新：说明现在两种模式都支持

---

## 总结

### 推荐实施方案

**阶段 1（MVP）：** 实现 Token 注册 + Service Token 注册
- 工作量：3-4 天
- 覆盖大部分使用场景
- 适合自动化部署

**阶段 2（完整版）：** 添加交互式注册
- 额外工作量：2-3 天
- 提升用户体验
- 与官方 warp-cli 功能对等

### 技术难点

1. **Token 获取流程**（中等）
   - 需要处理浏览器回调
   - 本地 HTTP 服务器

2. **API 兼容性**（低）
   - Zero Trust API 相对简单
   - 与消费者 API 类似

3. **配置管理**（低）
   - 需要区分两种模式
   - 配置文件格式扩展

### 预期效果

实现后，warp-cli 将成为：
- ✅ 支持消费者 WARP
- ✅ 支持 Zero Trust
- ✅ 跨平台命令行工具
- ✅ 适合自动化部署
- ✅ 功能完整的 WARP 客户端替代方案

---

## 下一步行动

1. **确认需求**
   - 是否需要交互式注册？
   - 优先级如何？

2. **准备测试环境**
   - 创建 Zero Trust 组织
   - 生成 Service Token

3. **开始实施**
   - 从 Token 注册开始
   - 逐步添加功能

4. **持续迭代**
   - 收集用户反馈
   - 优化用户体验
