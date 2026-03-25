# Zero Trust 快速开始指南

## 🎉 warp-cli 现在支持 Cloudflare Zero Trust！

warp-cli 现在可以连接到 Cloudflare Zero Trust 组织，让你可以：
- 访问企业内网资源
- 使用企业安全策略
- 通过命令行管理 Zero Trust 连接

---

## 三种注册方式

### 方法 1：Token URL（推荐用于手动注册）

**步骤：**

1. 在浏览器中访问你的团队注册页面：
   ```
   https://<your-team>.cloudflareaccess.com/warp
   ```

2. 使用你的企业账户登录（Email、SSO 等）

3. 认证成功后，复制获得的 token URL

4. 使用 token 注册：
   ```bash
   warp-cli register-team --token "https://myteam.cloudflareaccess.com/auth?token=xxx"
   ```

5. 连接：
   ```bash
   warp-cli connect
   ```

**示例：**
```bash
$ warp-cli register-team --token "https://acme.cloudflareaccess.com/auth?token=abc123def456"
Registering with token URL...
Team: acme

✓ Successfully registered with Cloudflare Zero Trust!
  Team: acme
  Device ID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

You can now connect using: warp-cli connect
```

---

### 方法 2：Service Token（推荐用于自动化）

**适用场景：**
- 服务器自动化部署
- CI/CD 流程
- 无人值守环境

**步骤：**

1. 管理员在 Cloudflare Dashboard 创建 Service Token：
   - 进入 Zero Trust Dashboard
   - Access → Service Auth → Service Tokens
   - 创建新的 Service Token
   - 复制 Client ID 和 Client Secret

2. 在设备上注册：
   ```bash
   warp-cli register-team \
     --team mycompany \
     --client-id "xxx" \
     --client-secret "yyy"
   ```

3. 连接：
   ```bash
   warp-cli connect
   ```

**示例：**
```bash
$ warp-cli register-team \
    --team acme \
    --client-id "a1b2c3d4e5f6" \
    --client-secret "secret123456"

Registering with service token for team: acme

Note: This device is enrolled as non_identity@acme.cloudflareaccess.com

✓ Successfully registered with Cloudflare Zero Trust!
  Team: acme
  Device ID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

You can now connect using: warp-cli connect
```

---

### 方法 3：交互式（生成 URL）

**步骤：**

1. 生成注册 URL：
   ```bash
   warp-cli register-team --team mycompany
   ```

2. 在浏览器中访问显示的 URL

3. 认证后获得 token URL

4. 使用 token URL 完成注册（回到方法 1）

**示例：**
```bash
$ warp-cli register-team --team acme

======================================================================
Please visit the following URL to authenticate:

  https://acme.cloudflareaccess.com/warp?pub=xxxxxxxxx

After authentication, you will receive a token URL.
Run the following command with your token:

  warp-cli register-team --token "<your-token-url>"
======================================================================

Private key generated and ready for use.
Please complete authentication in your browser.
```

---

## 完整使用流程

### 场景 1：企业员工首次使用

```bash
# 1. 访问团队注册页面（在浏览器中）
https://mycompany.cloudflareaccess.com/warp

# 2. 使用公司邮箱登录

# 3. 复制 token URL

# 4. 注册设备
warp-cli register-team --token "https://mycompany.cloudflareaccess.com/auth?token=xxx"

# 5. 连接
warp-cli connect

# 6. 验证连接
warp-cli status

# 7. 访问内网资源
curl http://internal-app.company.local
```

### 场景 2：服务器自动化部署

```bash
#!/bin/bash
# deploy-warp.sh

# 环境变量
TEAM_NAME="mycompany"
CLIENT_ID="your-client-id"
CLIENT_SECRET="your-client-secret"

# 安装 warp-cli
# ...

# 注册到 Zero Trust
warp-cli register-team \
  --team "$TEAM_NAME" \
  --client-id "$CLIENT_ID" \
  --client-secret "$CLIENT_SECRET"

# 连接
warp-cli connect

# 验证
warp-cli status

echo "Zero Trust connection established!"
```

### 场景 3：开发环境设置

```bash
# 1. 注册（使用 token）
warp-cli register-team --token "$TEAM_TOKEN_URL"

# 2. 连接
warp-cli connect

# 3. 验证可以访问内网服务
ping internal-db.company.local

# 4. 开始开发
npm run dev
```

---

## 配置文件

### 消费者 WARP 配置
```toml
# warp-cli-account.toml
device_id = "xxx"
access_token = "xxx"
private_key = "xxx"
license_key = "xxx"
mode = "consumer"
```

### Zero Trust 配置
```toml
# warp-cli-account.toml
device_id = "xxx"
access_token = "xxx"
private_key = "xxx"
mode = "team"
team_name = "mycompany"
```

---

## 常用命令

```bash
# 查看帮助
warp-cli register-team --help

# Token 注册
warp-cli register-team --token "https://team.cloudflareaccess.com/auth?token=xxx"

# Service Token 注册
warp-cli register-team --team mycompany --client-id xxx --client-secret yyy

# 生成注册 URL
warp-cli register-team --team mycompany

# 连接
warp-cli connect

# 断开
warp-cli disconnect

# 查看状态
warp-cli status

# 验证连接
warp-cli trace
```

---

## 常见问题

### Q1: 如何获取 Service Token？

**A:** 管理员需要在 Cloudflare Zero Trust Dashboard 中创建：
1. 登录 Cloudflare Dashboard
2. 进入 Zero Trust
3. Access → Service Auth → Service Tokens
4. 点击 "Create Service Token"
5. 复制 Client ID 和 Client Secret

### Q2: Token URL 在哪里获取？

**A:** 有两种方式：
1. 访问 `https://<team>.cloudflareaccess.com/warp` 并认证
2. 使用 `warp-cli register-team --team <name>` 生成 URL，然后在浏览器中认证

### Q3: 可以同时使用消费者 WARP 和 Zero Trust 吗？

**A:** 不能同时使用，但可以切换：
- 使用不同的配置文件：`--config team-account.toml`
- 或者重新注册切换模式

### Q4: Service Token 注册的设备显示什么用户？

**A:** 显示为 `non_identity@<team>.cloudflareaccess.com`，这是正常的。

### Q5: 注册失败怎么办？

**A:** 检查：
- Team name 是否正确
- Token 是否过期
- Service Token 是否有效
- 网络连接是否正常
- 管理员是否配置了设备注册策略

### Q6: 如何切换回消费者 WARP？

**A:** 删除配置文件并重新注册：
```bash
rm warp-cli-account.toml
warp-cli register  # 消费者 WARP
```

---

## 与消费者 WARP 的区别

| 特性 | 消费者 WARP | Zero Trust |
|------|------------|------------|
| 注册命令 | `register` | `register-team` |
| 需要认证 | ❌ | ✅ |
| 访问内网 | ❌ | ✅ |
| 企业策略 | ❌ | ✅ |
| 设备管理 | 个人 | 企业管理员 |
| 适用场景 | 个人 VPN | 企业访问 |

---

## 故障排除

### 问题：Token 无效

```
Error: registration failed with status 401: Unauthorized
```

**解决方案：**
- Token 可能已过期，重新获取
- 检查 team name 是否正确
- 确认你有权限注册设备

### 问题：Service Token 认证失败

```
Error: registration failed with status 403: Forbidden
```

**解决方案：**
- 检查 Client ID 和 Client Secret 是否正确
- 确认 Service Token 未被禁用
- 联系管理员检查设备注册策略

### 问题：无法访问内网资源

**解决方案：**
1. 检查连接状态：`warp-cli status`
2. 验证路由：`warp-cli trace`
3. 确认管理员已配置正确的网络策略
4. 检查防火墙规则

---

## 下一步

- 查看完整实现计划：[ZERO_TRUST_IMPLEMENTATION_PLAN.md](ZERO_TRUST_IMPLEMENTATION_PLAN.md)
- 了解 WARP vs Zero Trust：[WARP_VS_ZERO_TRUST.md](WARP_VS_ZERO_TRUST.md)
- 连接指南：[CONNECT_GUIDE.md](CONNECT_GUIDE.md)
- 文档索引：[DOCUMENTATION_INDEX.md](DOCUMENTATION_INDEX.md)

---

**注意：** 当前实现支持 Token 注册和 Service Token 注册。交互式注册（自动回调）将在后续版本中添加。
