# 自定义 WireGuard 服务器快速开始

## ✅ 现在支持自定义 WireGuard 服务器！

warp-cli 现在可以连接到任何 WireGuard 服务器，不仅限于 Cloudflare WARP。

---

## 快速开始

### 1. 准备配置文件

创建标准的 WireGuard 配置文件：

```ini
# my-server.conf
[Interface]
PrivateKey = YourPrivateKeyHere==
Address = 10.0.0.2/24
DNS = 8.8.8.8

[Peer]
PublicKey = ServerPublicKeyHere==
Endpoint = vpn.example.com:51820
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
```

### 2. 连接

```bash
warp-cli connect-custom --config my-server.conf
```

### 3. 断开

```bash
warp-cli disconnect
```

---

## 完整示例

### 场景 1：连接到个人 VPS

```bash
# 1. 在 VPS 上设置 WireGuard 服务器
# (假设已完成)

# 2. 创建客户端配置
cat > my-vps.conf <<'EOF'
[Interface]
PrivateKey = cGxlYXNlIHJlcGxhY2UgdGhpcyB3aXRoIHlvdXIga2V5
Address = 10.0.0.2/24
DNS = 1.1.1.1

[Peer]
PublicKey = c2VydmVyIHB1YmxpYyBrZXkgaGVyZQ==
Endpoint = my-vps.example.com:51820
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
EOF

# 3. 连接
warp-cli connect-custom --config my-vps.conf

# 输出：
# Connecting to custom WireGuard server...
#   Config: my-vps.conf
#   Interface: wg0
#   Using: /path/to/wireguard-go
# ✓ Config file validation passed
# Starting wireguard-go for interface wg0...
# Applying configuration to wg0...
#
# ✓ Successfully connected to custom WireGuard server!
#   Interface: wg0
#   Config: my-vps.conf
#
# To disconnect, run: warp-cli disconnect -i wg0
```

### 场景 2：使用不同的接口名

```bash
# 连接到第一个服务器
warp-cli connect-custom --config server1.conf --interface wg0

# 同时连接到第二个服务器
warp-cli connect-custom --config server2.conf --interface wg1

# 断开特定接口
warp-cli disconnect -i wg0
warp-cli disconnect -i wg1
```

### 场景 3：前台运行（调试）

```bash
# 前台运行，查看详细日志
warp-cli connect-custom --config my-server.conf --foreground

# Ctrl+C 断开
```

---

## 配置文件格式

### 最小配置

```ini
[Interface]
PrivateKey = <your-private-key>
Address = 10.0.0.2/24

[Peer]
PublicKey = <server-public-key>
Endpoint = server.example.com:51820
AllowedIPs = 0.0.0.0/0
```

### 完整配置

```ini
[Interface]
PrivateKey = <your-private-key>
Address = 10.0.0.2/24, fd00::2/64
DNS = 1.1.1.1, 8.8.8.8
MTU = 1420

[Peer]
PublicKey = <server-public-key>
PresharedKey = <optional-preshared-key>
Endpoint = server.example.com:51820
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
```

### 配置说明

**[Interface] 部分：**
- `PrivateKey`: 客户端私钥（必需）
- `Address`: 客户端 IP 地址（必需）
- `DNS`: DNS 服务器（可选）
- `MTU`: 最大传输单元（可选，默认 1420）

**[Peer] 部分：**
- `PublicKey`: 服务器公钥（必需）
- `Endpoint`: 服务器地址和端口（必需）
- `AllowedIPs`: 允许的 IP 范围（必需）
- `PresharedKey`: 预共享密钥（可选，增强安全性）
- `PersistentKeepalive`: 保持连接（可选，推荐 25 秒）

---

## 生成密钥对

如果你需要生成新的密钥对：

```bash
# 生成私钥
wg genkey > private.key

# 从私钥生成公钥
wg pubkey < private.key > public.key

# 查看密钥
cat private.key
cat public.key
```

---

## 与其他功能对比

### warp-cli 支持三种连接方式

| 命令 | 用途 | 配置来源 |
|------|------|---------|
| `connect` | Cloudflare WARP | 自动（从 Cloudflare API） |
| `connect` | Zero Trust | 自动（从 Zero Trust API） |
| `connect-custom` | 自定义服务器 | 手动（配置文件） |

### 使用场景

**使用 `connect`（Cloudflare）：**
- 需要 Cloudflare WARP 服务
- 需要 Zero Trust 企业访问
- 想要自动配置管理

**使用 `connect-custom`（自定义）：**
- 有自己的 WireGuard 服务器
- 需要连接到第三方 VPN
- 需要完全控制配置

---

## 常见问题

### Q1: 配置文件格式错误怎么办？

**A:** warp-cli 会验证配置文件，提示缺少的字段：

```
Error: invalid config file: missing required field: Endpoint (Peer endpoint)
```

检查配置文件是否包含所有必需字段。

### Q2: 可以同时连接 Cloudflare 和自定义服务器吗？

**A:** 可以，使用不同的接口名：

```bash
# Cloudflare WARP
warp-cli connect --interface wg0

# 自定义服务器
warp-cli connect-custom --config my-server.conf --interface wg1
```

### Q3: wireguard-go 未找到怎么办？

**A:** 编译 wireguard-go：

```bash
cd wireguard
make
cd ..
```

### Q4: 需要 root 权限吗？

**A:** 是的，创建网络接口需要 root 权限：

```bash
sudo warp-cli connect-custom --config my-server.conf
```

### Q5: 如何验证连接？

**A:** 使用 `wg` 命令：

```bash
# 查看接口状态
sudo wg show wg0

# 查看详细信息
sudo wg show wg0 dump
```

### Q6: 配置文件可以放在哪里？

**A:** 任何位置，使用绝对或相对路径：

```bash
# 相对路径
warp-cli connect-custom --config ./configs/server1.conf

# 绝对路径
warp-cli connect-custom --config /etc/wireguard/wg0.conf

# 当前目录
warp-cli connect-custom --config my-server.conf
```

---

## 故障排除

### 问题：接口已存在

```
Error: interface wg0 already exists
```

**解决方案：**
```bash
# 断开现有连接
warp-cli disconnect -i wg0

# 或使用不同的接口名
warp-cli connect-custom --config my-server.conf --interface wg1
```

### 问题：无法连接到服务器

**检查清单：**
1. 服务器是否运行？
2. 端口是否开放？（默认 51820）
3. 防火墙是否允许 UDP 流量？
4. Endpoint 地址是否正确？
5. 密钥是否匹配？

**测试连接：**
```bash
# 测试 UDP 端口
nc -u -v server.example.com 51820

# 查看 wireguard-go 日志
LOG_LEVEL=debug warp-cli connect-custom --config my-server.conf --foreground
```

### 问题：连接成功但无法访问网络

**检查路由：**
```bash
# 查看路由表
ip route show

# 查看接口状态
ip addr show wg0

# 测试连接
ping -I wg0 8.8.8.8
```

---

## 高级用法

### 分隧道（Split Tunneling）

只路由特定流量通过 VPN：

```ini
[Interface]
PrivateKey = <your-private-key>
Address = 10.0.0.2/24

[Peer]
PublicKey = <server-public-key>
Endpoint = server.example.com:51820
# 只路由 10.0.0.0/8 网段
AllowedIPs = 10.0.0.0/8
PersistentKeepalive = 25
```

### 多个 Peer

连接到多个服务器：

```ini
[Interface]
PrivateKey = <your-private-key>
Address = 10.0.0.2/24

[Peer]
PublicKey = <server1-public-key>
Endpoint = server1.example.com:51820
AllowedIPs = 10.0.0.0/24

[Peer]
PublicKey = <server2-public-key>
Endpoint = server2.example.com:51820
AllowedIPs = 10.1.0.0/24
```

### IPv6 支持

```ini
[Interface]
PrivateKey = <your-private-key>
Address = 10.0.0.2/24, fd00::2/64
DNS = 2606:4700:4700::1111

[Peer]
PublicKey = <server-public-key>
Endpoint = [2001:db8::1]:51820
AllowedIPs = 0.0.0.0/0, ::/0
```

---

## 与标准 WireGuard 工具对比

| 功能 | warp-cli | wg-quick | wireguard-go |
|------|----------|----------|--------------|
| 配置文件 | ✅ | ✅ | ✅ |
| 自动路由 | ❌ | ✅ | ❌ |
| 跨平台 | ✅ | ✅ | ✅ |
| Cloudflare WARP | ✅ | ❌ | ❌ |
| 统一命令 | ✅ | ❌ | ❌ |

---

## 总结

### 现在 warp-cli 支持

1. ✅ Cloudflare WARP（消费者）
2. ✅ Cloudflare Zero Trust（企业）
3. ✅ 自定义 WireGuard 服务器（新增）

### 一个工具，三种用途

```bash
# Cloudflare WARP
warp-cli register
warp-cli connect

# Zero Trust
warp-cli register-team --token "..."
warp-cli connect

# 自定义服务器
warp-cli connect-custom --config my-server.conf
```

---

## 相关文档

- 详细分析：[CUSTOM_WIREGUARD_SUPPORT.md](CUSTOM_WIREGUARD_SUPPORT.md)
- Cloudflare 连接：[CONNECT_GUIDE.md](CONNECT_GUIDE.md)
- Zero Trust：[ZERO_TRUST_QUICK_START.md](ZERO_TRUST_QUICK_START.md)
- 主文档：[README.md](README.md)
