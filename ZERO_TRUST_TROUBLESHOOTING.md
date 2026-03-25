# Zero Trust 注册故障排查指南

## Token URL 格式

Cloudflare Zero Trust 使用自定义协议的 token URL：

```
com.cloudflare.warp://<team-name>.cloudflareaccess.com/auth?token=<jwt-token>
```

示例：
```
com.cloudflare.warp://carmeltop.cloudflareaccess.com/auth?token=eyJhbGciOiJSUzI1NiI...
```

warp-cli 会自动解析这个 URL 并提取：
- 团队名称（例如：carmeltop）
- JWT token

## 错误: "failed to parse response: invalid character '<' looking for beginning of value"

### 问题原因

这个错误表明服务器返回的是 HTML 页面而不是 JSON 响应，通常由以下原因导致：

### 1. 团队名称（Team Name）错误

**症状**: API 端点不存在或返回 404 页面

**解决方案**:
- 确认你的 Zero Trust 团队名称正确
- 团队名称应该是你的组织在 Cloudflare Zero Trust 中的唯一标识符
- 格式: `https://<team-name>.cloudflareaccess.com`

**如何查找团队名称**:
1. 登录 [Cloudflare Zero Trust Dashboard](https://one.dash.cloudflare.com/)
2. 进入 Settings → General
3. 查看 "Team domain" 或 "Auth domain"
4. 例如: `mycompany.cloudflareaccess.com` → 团队名称是 `mycompany`

### 2. Token 格式或权限错误

**症状**: 认证失败，返回登录页面

**Token URL 方式**:
```bash
# 正确的 token URL 格式
warp-cli register-team --token-url "https://mycompany.cloudflareaccess.com/warp?token=xxx"
```

Token URL 应该包含:
- 完整的团队域名
- `/warp` 路径
- `?token=` 参数

**Service Token 方式**:
```bash
# 需要同时提供 Client ID 和 Client Secret
warp-cli register-team \
  --team mycompany \
  --client-id "xxx.access" \
  --client-secret "yyy"
```

Service Token 要求:
- Client ID 通常以 `.access` 结尾
- Client Secret 是一个长字符串
- Service Token 必须有访问 WARP 设备注册的权限

### 3. API 端点变更

**症状**: Cloudflare 可能更改了 API 端点

**当前使用的端点**:
```
POST https://<team-name>.cloudflareaccess.com/warp
```

**检查方法**:
1. 使用浏览器访问 `https://<team-name>.cloudflareaccess.com`
2. 确认域名可以访问且不返回错误
3. 检查 Cloudflare 官方文档是否有 API 变更

### 4. 网络问题

**症状**: 请求被代理或防火墙拦截

**检查方法**:
```bash
# 测试网络连接
curl -v https://<team-name>.cloudflareaccess.com

# 检查是否有代理
echo $HTTP_PROXY
echo $HTTPS_PROXY
```

**解决方案**:
- 确保可以访问 `*.cloudflareaccess.com`
- 检查公司防火墙设置
- 尝试使用 VPN 或更换网络环境

## 调试步骤

### 步骤 1: 查看完整错误信息

重新编译后的版本会显示服务器返回的完整内容：

```bash
./warp-cli register-team --token-url "https://..."
```

错误信息现在会包含:
```
registration failed: failed to parse response: invalid character '<' looking for beginning of value
Response body: <html>...</html>
```

### 步骤 2: 分析响应内容

**如果看到 HTML 登录页面**:
- Token 无效或已过期
- 需要重新生成 token

**如果看到 404 页面**:
- 团队名称错误
- API 端点不存在

**如果看到 403 页面**:
- Service Token 权限不足
- 需要配置正确的访问策略

### 步骤 3: 验证配置

使用交互式模式重新尝试：

```bash
./warp-cli register-team --interactive
```

系统会引导你输入:
1. 团队名称
2. 认证方式（Token URL 或 Service Token）
3. 相应的认证凭据

### 步骤 4: 手动测试 API

使用 curl 直接测试 API：

```bash
# 测试 Token 认证
curl -X POST \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"key":"test","tos":"2024-01-01T00:00:00Z","type":"Linux","model":"PC","locale":"en_US"}' \
  https://YOUR_TEAM.cloudflareaccess.com/warp

# 测试 Service Token 认证
curl -X POST \
  -H "CF-Access-Client-Id: YOUR_CLIENT_ID" \
  -H "CF-Access-Client-Secret: YOUR_CLIENT_SECRET" \
  -H "Content-Type: application/json" \
  -d '{"key":"test","tos":"2024-01-01T00:00:00Z","type":"Linux","model":"PC","locale":"en_US"}' \
  https://YOUR_TEAM.cloudflareaccess.com/warp
```

## 常见解决方案

### 方案 1: 重新获取 Token URL

1. 登录 Cloudflare Zero Trust Dashboard
2. 进入 Settings → WARP Client
3. 点击 "Manage" → "Add a device"
4. 选择 "Enroll a device"
5. 复制生成的 enrollment URL（包含 token）

### 方案 2: 创建 Service Token

1. 登录 Cloudflare Zero Trust Dashboard
2. 进入 Access → Service Auth → Service Tokens
3. 点击 "Create Service Token"
4. 命名并保存 Client ID 和 Client Secret
5. 配置访问策略允许该 Service Token 访问 WARP 注册

### 方案 3: 检查 Zero Trust 配置

确保 Zero Trust 已正确配置:
1. WARP Client 功能已启用
2. 设备注册策略已配置
3. 网络路由已设置（如果需要）

## 获取帮助

如果以上方法都无法解决问题，请提供以下信息：

1. 完整的错误消息（包括 Response body）
2. 使用的命令
3. 团队名称（可以脱敏）
4. 认证方式（Token URL 或 Service Token）
5. curl 测试的结果

## 相关文档

- [Cloudflare Zero Trust 文档](https://developers.cloudflare.com/cloudflare-one/)
- [WARP Client 配置](https://developers.cloudflare.com/cloudflare-one/connections/connect-devices/warp/)
- [Service Tokens](https://developers.cloudflare.com/cloudflare-one/identity/service-tokens/)
