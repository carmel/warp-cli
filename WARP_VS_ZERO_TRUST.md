# Cloudflare WARP vs Zero Trust 使用说明

## 核心区别

你提到的是两种**完全不同**的 Cloudflare 服务和使用场景：

### 1️⃣ warp-cli（本项目）- 个人消费者 WARP 服务

**使用场景：** 个人用户的 VPN 服务

**API 端点：** `https://api.cloudflareclient.com`

**注册方式：**
```bash
warp-cli register
```

**这个命令在做什么：**
- 调用 Cloudflare 的**消费者 WARP API**
- 创建一个**匿名的个人账户**
- 生成 WireGuard 密钥对
- 获取设备 ID 和访问令牌
- 获取 WireGuard 配置（服务器公钥、端点、IP 地址等）

**不需要：**
- ❌ Email 地址
- ❌ 组织名称
- ❌ Zero Trust 账户
- ❌ Cloudflare 账户登录

**适用于：**
- ✅ 个人用户使用免费 WARP VPN
- ✅ 个人用户使用 Warp+ 订阅
- ✅ 简单的 VPN 需求
- ✅ 1.1.1.1 应用的替代方案

---

### 2️⃣ Cloudflare Zero Trust（你提到的）- 企业服务

**使用场景：** 企业/组织的安全访问服务

**API 端点：** `https://api.cloudflare.com/client/v4/`

**注册方式：**
- 在 Cloudflare Dashboard 创建 Zero Trust 组织
- 配置访问策略
- 员工通过 email + 组织名登录

**需要：**
- ✅ Email 地址
- ✅ 组织名称（team name）
- ✅ Cloudflare 账户
- ✅ 管理员配置的访问策略

**适用于：**
- ✅ 企业内网访问
- ✅ 应用访问控制
- ✅ 身份验证和授权
- ✅ 安全策略管理

---

## 详细对比

### 注册流程对比

#### warp-cli register（消费者 WARP）

```bash
$ warp-cli register

# 内部流程：
1. 生成 WireGuard 密钥对（私钥 + 公钥）
2. 调用 API: POST https://api.cloudflareclient.com/v0a1922/reg
   {
     "key": "公钥",
     "install_id": "",
     "fcm_token": "",
     "tos": "2024-03-24T10:00:00.000Z",
     "type": "Android",
     "model": "PC",
     "locale": "en_US"
   }
3. 收到响应：
   {
     "id": "设备ID",
     "token": "访问令牌",
     "account": {
       "id": "账户ID",
       "account_type": "free",
       "license": "许可证密钥"
     },
     "config": {
       "peers": [{
         "public_key": "服务器公钥",
         "endpoint": {
           "host": "engage.cloudflareclient.com:2408"
         }
       }],
       "interface": {
         "addresses": {
           "v4": "172.16.0.2",
           "v6": "2606:4700:110:8xxx::xxx"
         }
       }
     }
   }
4. 保存到配置文件：warp-cli-account.toml
```

**特点：**
- 完全匿名，无需任何个人信息
- 自动分配 IP 地址
- 立即可用
- 类似于 1.1.1.1 应用的注册流程

#### Cloudflare Zero Trust（企业 WARP）

```bash
# 用户操作：
1. 打开 Cloudflare WARP 客户端
2. 点击 "Settings" → "Account"
3. 选择 "Login with Cloudflare Zero Trust"
4. 输入组织名称（例如：mycompany）
5. 输入 email 地址
6. 验证 email（收到验证码或链接）
7. 通过管理员配置的身份验证（SSO、SAML 等）

# 内部流程：
1. 连接到组织的 Zero Trust 端点
2. 进行身份验证
3. 获取组织的访问策略
4. 建立安全隧道到企业资源
```

**特点：**
- 需要身份验证
- 基于策略的访问控制
- 可以访问企业内网资源
- 管理员可以管理和监控

---

## 技术实现对比

### warp-cli（消费者 WARP）

```go
// 注册流程
func Register(publicKey *util.Key, deviceModel string) (*openapi.Register200Response, error) {
    timestamp := util.GetTimestamp()
    result, _, err := apiClient.DefaultAPI.
        Register(context.TODO(), ApiVersion).
        RegisterRequest(openapi.RegisterRequest{
            FcmToken:  "",
            InstallId: "",
            Key:       publicKey.String(),  // WireGuard 公钥
            Locale:    "en_US",
            Model:     deviceModel,
            Tos:       timestamp,           // 接受服务条款的时间戳
            Type:      "Android",
        }).Execute()
    return result, err
}
```

**API 端点：**
- 注册：`POST /v0a1922/reg`
- 获取配置：`GET /v0a1922/reg/{deviceId}`
- 更新账户：`PATCH /v0a1922/reg/{deviceId}/account`

**认证方式：**
- Bearer Token（注册时获得）
- 无需用户凭证

### Cloudflare Zero Trust

**API 端点：**
- 认证：`https://<team-name>.cloudflareaccess.com`
- 策略：`https://api.cloudflare.com/client/v4/accounts/{account_id}/access/`

**认证方式：**
- OAuth 2.0
- SAML
- OIDC
- Email OTP
- 等等

---

## 使用场景示例

### 场景 1：个人用户想要 VPN（使用 warp-cli）

```bash
# 小明想要一个简单的 VPN 来保护隐私
warp-cli register
warp-cli connect

# 完成！无需任何账户或 email
```

### 场景 2：企业员工访问公司资源（使用 Zero Trust）

```bash
# 小红是 ABC 公司的员工，需要访问公司内网

1. 打开 Cloudflare WARP 客户端
2. 输入组织名：abc-company
3. 输入公司 email：xiaohong@abc.com
4. 验证身份（可能需要 SSO 登录）
5. 连接后可以访问公司内网资源
```

### 场景 3：个人用户想要 Warp+（使用 warp-cli）

```bash
# 小明想要更快的速度，购买了 Warp+
warp-cli register
warp-cli update --license-key "从1.1.1.1应用获取的密钥"
warp-cli connect

# 现在享受 Warp+ 的速度
```

---

## 为什么 warp-cli 不支持 Zero Trust？

### 技术原因

1. **不同的 API**
   - 消费者 WARP：`api.cloudflareclient.com`
   - Zero Trust：`api.cloudflare.com` + 组织特定端点

2. **不同的认证机制**
   - 消费者 WARP：简单的 Bearer Token
   - Zero Trust：复杂的 OAuth/SAML/OIDC 流程

3. **不同的配置模型**
   - 消费者 WARP：标准 WireGuard 配置
   - Zero Trust：动态策略 + 隧道配置

### 设计目标不同

**warp-cli 的目标：**
- 提供简单的 VPN 功能
- 替代 1.1.1.1 应用
- 跨平台命令行工具
- 无需账户即可使用

**Zero Trust 的目标：**
- 企业安全访问
- 身份验证和授权
- 策略管理
- 审计和监控

---

## 如何使用 Zero Trust？

如果你需要使用 Cloudflare Zero Trust，应该：

### 方法 1：使用官方客户端（推荐）

1. 下载 Cloudflare WARP 官方客户端
   - Windows: https://install.appcenter.ms/orgs/cloudflare/apps/1.1.1.1-windows-1/distribution_groups/release
   - macOS: https://install.appcenter.ms/orgs/cloudflare/apps/1.1.1.1-macos-1/distribution_groups/release
   - Linux: https://pkg.cloudflareclient.com/

2. 安装并打开客户端

3. 点击设置 → 账户 → Login with Cloudflare Zero Trust

4. 输入组织名称和 email

### 方法 2：使用 warp-connector（服务器端）

如果你是管理员，想在服务器上部署：

```bash
# 安装 cloudflared
curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o cloudflared
chmod +x cloudflared

# 认证
cloudflared tunnel login

# 创建隧道
cloudflared tunnel create my-tunnel

# 配置和运行
cloudflared tunnel run my-tunnel
```

---

## 常见问题

### Q1: 我可以用 warp-cli 连接到 Zero Trust 吗？

**A:** 不可以。warp-cli 只支持消费者 WARP 服务，不支持 Zero Trust。

### Q2: 我有 Zero Trust 账户，可以用 warp-cli 吗？

**A:** 可以，但是是两个独立的服务：
- 你可以用 warp-cli 注册一个**消费者 WARP 账户**（与 Zero Trust 无关）
- 你的 Zero Trust 账户仍然需要通过官方客户端使用

### Q3: warp-cli register 需要 email 吗？

**A:** 不需要。这是完全匿名的注册，只需要生成密钥对。

### Q4: 为什么官方客户端需要 email，warp-cli 不需要？

**A:** 因为你使用的是官方客户端的 **Zero Trust 模式**，而 warp-cli 使用的是**消费者 WARP 模式**。

官方客户端支持两种模式：
- **消费者模式**：无需 email（类似 warp-cli）
- **Zero Trust 模式**：需要 email 和组织名

### Q5: warp-cli 注册的账户可以在 1.1.1.1 应用中使用吗？

**A:** 不能直接使用，但可以通过 license key 关联：
1. 在 1.1.1.1 应用中查看你的 license key
2. 在 warp-cli 中绑定这个 key
3. 这样两个设备共享同一个账户

### Q6: 我应该使用哪个？

**选择 warp-cli（消费者 WARP）如果：**
- ✅ 你是个人用户
- ✅ 只需要简单的 VPN 功能
- ✅ 想要命令行工具
- ✅ 不需要企业功能

**选择 Zero Trust 如果：**
- ✅ 你是企业用户
- ✅ 需要访问公司内网
- ✅ 需要身份验证和访问控制
- ✅ 管理员要求使用

---

## 总结

### warp-cli register 在做什么？

```
简单来说：
1. 生成一对 WireGuard 密钥
2. 向 Cloudflare 的消费者 API 注册一个匿名设备
3. 获取 VPN 配置（IP 地址、服务器信息等）
4. 保存到本地配置文件

就像你第一次打开 1.1.1.1 应用时，它自动做的事情。
```

### 与 Zero Trust 的区别

| 特性 | warp-cli（消费者 WARP） | Zero Trust |
|------|------------------------|------------|
| 需要 Email | ❌ | ✅ |
| 需要组织名 | ❌ | ✅ |
| 身份验证 | ❌ | ✅ |
| 访问企业资源 | ❌ | ✅ |
| 个人 VPN | ✅ | ✅ |
| 命令行工具 | ✅ | ❌（需要官方客户端） |
| 完全匿名 | ✅ | ❌ |
| API 端点 | api.cloudflareclient.com | api.cloudflare.com |

---

## 相关资源

- **消费者 WARP**: https://1.1.1.1/
- **Zero Trust**: https://www.cloudflare.com/zero-trust/
- **Zero Trust 文档**: https://developers.cloudflare.com/cloudflare-one/
- **warp-cli 文档**: [README.md](README.md)

---

**结论：** warp-cli 是为个人消费者 WARP 服务设计的，与企业的 Zero Trust 服务是完全不同的产品。如果你需要使用 Zero Trust，请使用官方的 Cloudflare WARP 客户端。
