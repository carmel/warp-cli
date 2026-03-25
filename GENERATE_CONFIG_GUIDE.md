# WireGuard 配置生成指南

## 概述

`generate-config` 命令帮助你快速创建 WireGuard 配置文件，用于 `connect-custom` 命令。

## 三种生成模式

### 1. 模板模式（Template Mode）

生成带注释的配置模板，适合手动填写。

```bash
warp-cli generate-config --template -o my-server.conf
```

**生成的模板：**
```ini
[Interface]
# Your client's private key (generate with: wg genkey)
PrivateKey = <YOUR_PRIVATE_KEY>

# Your client's IP address in the VPN network
Address = 10.0.0.2/24

# DNS servers (optional)
DNS = 1.1.1.1, 8.8.8.8

[Peer]
# Server's public key
PublicKey = <SERVER_PUBLIC_KEY>

# Server's endpoint (IP:port or domain:port)
Endpoint = <SERVER_IP_OR_DOMAIN>:51820

# Which traffic to route through the VPN
AllowedIPs = 0.0.0.0/0, ::/0

# Keep connection alive (recommended for NAT traversal)
PersistentKeepalive = 25
```

**使用场景：**
- 第一次设置 WireGuard
- 需要了解每个字段的含义
- 想要手动编辑配置

---

### 2. 交互模式（Interactive Mode）

通过问答方式引导创建配置。

```bash
warp-cli generate-config --interactive -o my-server.conf
```

**交互流程：**
```
Interactive WireGuard Configuration Generator
==================================================

? Private Key: 
  ▸ Generate new key
    Enter existing key

? Client IP Address: (10.0.0.2/24) 

? DNS Servers: (1.1.1.1, 8.8.8.8) 

? Server Public Key: 

? Server Endpoint (IP:port or domain:port): (server.example.com:51820) 

? Allowed IPs: (0.0.0.0/0, ::/0) 

? Persistent Keepalive (seconds, 0 to disable): (25) 

✓ Successfully generated WireGuard configuration!
  Output: my-server.conf

To connect using this config:
  warp-cli connect-custom --config my-server.conf
```

**使用场景：**
- 快速创建配置
- 不熟悉配置格式
- 需要自动生成密钥

---

### 3. 直接模式（Direct Mode）

通过命令行参数直接生成配置。

```bash
warp-cli generate-config -o my-server.conf \
  --private-key "YourPrivateKeyHere==" \
  --address "10.0.0.2/24" \
  --public-key "ServerPublicKeyHere==" \
  --endpoint "vpn.example.com:51820"
```

**所有可用参数：**
```bash
--private-key string    # 客户端私钥（必需）
--address string        # 客户端 IP 地址（必需）
--dns string            # DNS 服务器（默认：1.1.1.1, 8.8.8.8）
--public-key string     # 服务器公钥（必需）
--endpoint string       # 服务器端点（必需）
--allowed-ips string    # 允许的 IP（默认：0.0.0.0/0, ::/0）
--keepalive int         # 保持连接间隔（默认：25）
-o, --output string     # 输出文件（默认：wireguard.conf）
```

**使用场景：**
- 自动化脚本
- 批量生成配置
- 已知所有参数

---

## 完整工作流程

### 场景 1：从零开始设置 VPN

```bash
# 1. 生成模板
warp-cli generate-config --template -o my-vpn.conf

# 2. 生成密钥对
wg genkey > private.key
wg pubkey < private.key > public.key

# 3. 查看密钥
cat private.key  # 客户端私钥
cat public.key   # 客户端公钥（发送给服务器管理员）

# 4. 编辑配置文件
# 将 <YOUR_PRIVATE_KEY> 替换为 private.key 的内容
# 将 <SERVER_PUBLIC_KEY> 替换为服务器提供的公钥
# 将 <SERVER_IP_OR_DOMAIN> 替换为服务器地址
nano my-vpn.conf

# 5. 连接
warp-cli connect-custom --config my-vpn.conf
```

### 场景 2：使用交互模式快速设置

```bash
# 1. 交互式生成配置
warp-cli generate-config --interactive -o my-vpn.conf

# 选择 "Generate new key" 自动生成密钥
# 按提示输入其他信息

# 2. 连接
warp-cli connect-custom --config my-vpn.conf
```

### 场景 3：自动化部署

```bash
#!/bin/bash
# deploy-vpn.sh

# 生成密钥
PRIVATE_KEY=$(wg genkey)
PUBLIC_KEY=$(echo "$PRIVATE_KEY" | wg pubkey)

echo "Client public key (share with server): $PUBLIC_KEY"

# 从服务器获取配置信息
SERVER_PUBLIC_KEY="xxx"
SERVER_ENDPOINT="vpn.example.com:51820"
CLIENT_ADDRESS="10.0.0.2/24"

# 生成配置
warp-cli generate-config -o client.conf \
  --private-key "$PRIVATE_KEY" \
  --address "$CLIENT_ADDRESS" \
  --public-key "$SERVER_PUBLIC_KEY" \
  --endpoint "$SERVER_ENDPOINT"

# 连接
warp-cli connect-custom --config client.conf
```

---

## 配置字段说明

### [Interface] 部分

| 字段 | 必需 | 说明 | 示例 |
|------|------|------|------|
| PrivateKey | ✅ | 客户端私钥 | `cGxlYXNlIHJlcGxhY2U=` |
| Address | ✅ | 客户端 VPN IP | `10.0.0.2/24` |
| DNS | ❌ | DNS 服务器 | `1.1.1.1, 8.8.8.8` |
| MTU | ❌ | 最大传输单元 | `1420` |

### [Peer] 部分

| 字段 | 必需 | 说明 | 示例 |
|------|------|------|------|
| PublicKey | ✅ | 服务器公钥 | `c2VydmVyIHB1YmxpYw=` |
| Endpoint | ✅ | 服务器地址:端口 | `vpn.example.com:51820` |
| AllowedIPs | ✅ | 路由的 IP 范围 | `0.0.0.0/0, ::/0` |
| PresharedKey | ❌ | 预共享密钥 | `cHJlc2hhcmVkIGtleQ=` |
| PersistentKeepalive | ❌ | 保持连接间隔 | `25` |

---

## 高级配置示例

### 分隧道（Split Tunneling）

只路由特定流量通过 VPN：

```bash
warp-cli generate-config -o split-tunnel.conf \
  --private-key "xxx" \
  --address "10.0.0.2/24" \
  --public-key "yyy" \
  --endpoint "vpn.example.com:51820" \
  --allowed-ips "10.0.0.0/8, 192.168.0.0/16"
```

### IPv6 支持

```bash
warp-cli generate-config -o ipv6.conf \
  --private-key "xxx" \
  --address "10.0.0.2/24, fd00::2/64" \
  --dns "2606:4700:4700::1111" \
  --public-key "yyy" \
  --endpoint "[2001:db8::1]:51820" \
  --allowed-ips "0.0.0.0/0, ::/0"
```

### 禁用 Keepalive

```bash
warp-cli generate-config -o no-keepalive.conf \
  --private-key "xxx" \
  --address "10.0.0.2/24" \
  --public-key "yyy" \
  --endpoint "vpn.example.com:51820" \
  --keepalive 0
```

---

## 密钥生成

### 使用 wg 命令

```bash
# 生成私钥
wg genkey > private.key

# 从私钥生成公钥
wg pubkey < private.key > public.key

# 生成预共享密钥（可选）
wg genpsk > preshared.key

# 查看密钥
cat private.key
cat public.key
```

### 在交互模式中自动生成

选择 "Generate new key" 选项，warp-cli 会自动：
1. 生成私钥
2. 生成对应的公钥
3. 显示公钥（用于发送给服务器管理员）

---

## 配置验证

生成配置后，可以验证格式：

```bash
# 使用 wg 命令验证
wg show < my-server.conf

# 或使用 connect-custom 的验证功能
warp-cli connect-custom --config my-server.conf
# 如果配置有问题，会显示错误信息
```

---

## 常见问题

### Q1: 如何获取服务器公钥？

**A:** 联系服务器管理员，或在服务器上运行：
```bash
sudo wg show wg0 public-key
```

### Q2: 客户端公钥在哪里？

**A:** 
- 交互模式会自动显示
- 或手动生成：`wg pubkey < private.key`

### Q3: AllowedIPs 应该设置什么？

**A:** 
- `0.0.0.0/0, ::/0` - 全部流量（全隧道）
- `10.0.0.0/8` - 只路由特定网段（分隧道）
- `192.168.1.0/24` - 只路由单个子网

### Q4: 为什么需要 PersistentKeepalive？

**A:** 
- 保持 NAT 映射活跃
- 防止连接超时
- 推荐值：25 秒
- 如果服务器在公网且客户端有固定 IP，可以设为 0

### Q5: 生成的配置文件可以直接用吗？

**A:** 
- 模板模式：需要手动填写占位符
- 交互模式：可以直接使用
- 直接模式：可以直接使用

### Q6: 如何修改已生成的配置？

**A:** 
```bash
# 方法 1：重新生成
warp-cli generate-config --template -o my-server.conf

# 方法 2：直接编辑
nano my-server.conf

# 方法 3：使用交互模式重新创建
warp-cli generate-config --interactive -o my-server.conf
```

---

## 与其他命令配合使用

### 完整流程

```bash
# 1. 生成配置
warp-cli generate-config --interactive -o my-vpn.conf

# 2. 连接
warp-cli connect-custom --config my-vpn.conf

# 3. 验证连接
wg show wg0

# 4. 断开
warp-cli disconnect -i wg0
```

### 多配置管理

```bash
# 生成多个配置
warp-cli generate-config --template -o home-vpn.conf
warp-cli generate-config --template -o office-vpn.conf
warp-cli generate-config --template -o mobile-vpn.conf

# 根据需要连接不同的 VPN
warp-cli connect-custom --config home-vpn.conf --interface wg0
warp-cli connect-custom --config office-vpn.conf --interface wg1
```

---

## 命令参考

### 基本用法

```bash
# 生成模板
warp-cli generate-config --template

# 交互模式
warp-cli generate-config --interactive

# 直接模式
warp-cli generate-config \
  --private-key "xxx" \
  --address "10.0.0.2/24" \
  --public-key "yyy" \
  --endpoint "server.com:51820"
```

### 完整参数

```bash
warp-cli generate-config [flags]

Flags:
  -o, --output string        输出文件路径
  -t, --template             生成模板
  -i, --interactive          交互模式
      --private-key string   客户端私钥
      --address string       客户端 IP 地址
      --dns string           DNS 服务器
      --public-key string    服务器公钥
      --endpoint string      服务器端点
      --allowed-ips string   允许的 IP
      --keepalive int        保持连接间隔
  -h, --help                 帮助信息
```

---

## 最佳实践

1. **使用交互模式快速开始**
   - 适合新手
   - 自动生成密钥
   - 有提示和默认值

2. **保存密钥**
   ```bash
   # 生成配置时保存私钥
   warp-cli generate-config --interactive -o my-vpn.conf
   # 从配置中提取私钥备份
   grep PrivateKey my-vpn.conf > private.key.backup
   ```

3. **使用有意义的文件名**
   ```bash
   warp-cli generate-config -o home-server.conf
   warp-cli generate-config -o office-vpn.conf
   warp-cli generate-config -o mobile-hotspot.conf
   ```

4. **设置正确的文件权限**
   ```bash
   chmod 600 my-vpn.conf  # 只有所有者可读写
   ```

5. **版本控制**
   ```bash
   # 不要将私钥提交到 git
   echo "*.conf" >> .gitignore
   echo "*.key" >> .gitignore
   ```

---

## 总结

`generate-config` 命令提供三种方式生成 WireGuard 配置：

| 模式 | 命令 | 适用场景 |
|------|------|---------|
| 模板 | `--template` | 学习、手动配置 |
| 交互 | `--interactive` | 快速设置、新手 |
| 直接 | 参数指定 | 自动化、脚本 |

选择适合你的方式，快速创建配置并连接到 VPN！

---

## 相关文档

- 连接自定义服务器：[CUSTOM_WIREGUARD_QUICK_START.md](CUSTOM_WIREGUARD_QUICK_START.md)
- 详细分析：[CUSTOM_WIREGUARD_SUPPORT.md](CUSTOM_WIREGUARD_SUPPORT.md)
- 主文档：[README.md](README.md)
