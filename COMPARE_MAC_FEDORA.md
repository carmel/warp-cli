# Mac vs Fedora WARP 连接对比

## 当前状况

- **Mac**: Cloudflare WARP 官方客户端能成功连接 Zero Trust
- **Fedora**: 使用相同账户和配置无法连接（0 B received，无握手）

## 需要对比的信息

### 在 Mac 上收集

```bash
# 1. WireGuard 状态
sudo wg show

# 2. 网络接口配置
ifconfig utun

# 3. 路由表
netstat -rn | grep utun

# 4. DNS 配置
scutil --dns | grep -A 5 utun

# 5. 配置文件（如果能找到）
sudo find /opt/homebrew /usr/local /Library -name "*.conf" -path "*cloudflare*" 2>/dev/null
```

### 在 Fedora 上对比

```bash
# 1. WireGuard 状态
sudo wg show wg0

# 2. 网络接口配置
ip addr show wg0

# 3. 路由表
ip route | grep wg0

# 4. DNS 配置
resolvectl status wg0

# 5. 配置文件
cat wg0.conf
```

## 可能的差异点

### 1. Endpoint 地址

Mac 和 Fedora 可能使用不同的 Cloudflare endpoint：
- Mac 可能使用地理位置更近的服务器
- Fedora 可能被分配到不可达的服务器

**检查方法**：
```bash
# Mac
sudo wg show | grep endpoint

# Fedora
sudo wg show wg0 | grep endpoint
```

### 2. 密钥

虽然使用同一账户，但可能是不同的设备注册：
- Mac 有自己的设备 ID 和密钥对
- Fedora 有自己的设备 ID 和密钥对

**检查方法**：
```bash
# Mac
sudo wg show | grep "public key"

# Fedora
sudo wg show wg0 | grep "public key"
```

### 3. 网络环境

虽然在同一局域网，但可能有细微差异：
- 防火墙规则
- NAT 配置
- IPv6 支持

### 4. WireGuard 实现

- Mac: 可能使用内核扩展或用户空间实现
- Fedora: 使用内核模块

## 测试步骤

### 步骤 1: 在 Mac 上导出配置

如果 Mac 上的 WARP 使用标准 WireGuard 配置，尝试导出：

```bash
# 查找配置文件
sudo find / -name "*.conf" -path "*warp*" 2>/dev/null

# 或查看 wg 配置
sudo wg showconf utun0 > mac-warp.conf
```

### 步骤 2: 在 Fedora 上使用 Mac 的配置

```bash
# 复制 Mac 的配置到 Fedora
# 然后测试
sudo wg-quick up mac-warp
```

### 步骤 3: 检查 endpoint 可达性

```bash
# 在 Mac 上
ENDPOINT=$(sudo wg show | grep endpoint | awk '{print $2}')
echo "Mac endpoint: $ENDPOINT"
nc -vzu $(echo $ENDPOINT | cut -d: -f1) $(echo $ENDPOINT | cut -d: -f2)

# 在 Fedora 上测试相同的 endpoint
nc -vzu $(echo $ENDPOINT | cut -d: -f1) $(echo $ENDPOINT | cut -d: -f2)
```

## 可能的解决方案

### 方案 1: 使用 Mac 的 endpoint

如果 Mac 使用不同的 endpoint，修改 Fedora 配置：

```bash
# 编辑 wg0.conf
# 将 Endpoint 改为 Mac 使用的地址
sudo nano wg0.conf

# 重新连接
sudo wg-quick down wg0
sudo wg-quick up wg0
```

### 方案 2: 在 Fedora 上重新注册

可能需要在 Fedora 上使用不同的设备注册：

```bash
# 删除现有配置
rm warp-cli-account.toml warp-cli-profile.conf

# 重新注册（使用新的 token）
./warp-cli register-team --token "NEW_TOKEN"

# 连接
sudo ./warp-cli connect
```

### 方案 3: 使用 Cloudflare Tunnel

如果 WARP 完全无法工作，使用 Cloudflare Tunnel 作为替代：

```bash
# 安装 cloudflared
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64
chmod +x cloudflared-linux-amd64
sudo mv cloudflared-linux-amd64 /usr/local/bin/cloudflared

# 登录
cloudflared tunnel login

# 创建隧道
cloudflared tunnel create my-tunnel

# 运行隧道
cloudflared tunnel run my-tunnel
```

### 方案 4: 检查 ISP 限制

在 Fedora 上测试是否是 ISP 限制：

```bash
# 测试不同的 Cloudflare IP
for ip in 162.159.192.1 162.159.193.1 162.159.194.1 162.159.195.1; do
    echo "Testing $ip..."
    timeout 2 bash -c "echo > /dev/udp/$ip/2408" 2>/dev/null && echo "  OK" || echo "  FAIL"
done

# 使用 traceroute 查看路由
traceroute -n 162.159.192.1
```

## 调试信息收集

请收集以下信息以便进一步分析：

### Mac 信息

```bash
echo "=== Mac WireGuard 配置 ==="
sudo wg show

echo ""
echo "=== Mac 网络接口 ==="
ifconfig | grep -A 10 utun

echo ""
echo "=== Mac 路由 ==="
netstat -rn | grep -E "Destination|utun"

echo ""
echo "=== Mac Cloudflare 测试 ==="
curl https://www.cloudflare.com/cdn-cgi/trace/
```

### Fedora 信息

```bash
echo "=== Fedora WireGuard 配置 ==="
sudo wg show wg0

echo ""
echo "=== Fedora 网络接口 ==="
ip addr show wg0

echo ""
echo "=== Fedora 路由 ==="
ip route | grep wg0

echo ""
echo "=== Fedora 配置文件 ==="
cat wg0.conf

echo ""
echo "=== Fedora 网络测试 ==="
ping -c 2 162.159.192.1
nc -vzu 162.159.192.1 2408
```

## 下一步

1. 收集 Mac 和 Fedora 的对比信息
2. 检查 endpoint 是否不同
3. 测试 Mac 的 endpoint 在 Fedora 上是否可达
4. 如果可达，使用 Mac 的配置
5. 如果不可达，说明是网络环境问题，需要联系网络管理员或 ISP
