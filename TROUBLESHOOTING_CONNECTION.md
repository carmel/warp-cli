# VPN 连接故障排查指南

## 问题：连接成功但无法访问外网

当 `warp-cli connect` 显示成功，但浏览器无法打开外网（如 Google）时，通常是以下几个问题：

### 1. 检查接口状态

```bash
# 查看 WireGuard 接口
sudo wg show wg0

# 查看接口 IP 配置
ip addr show wg0

# 查看接口是否 UP
ip link show wg0
```

期望看到：
- 接口状态为 `UP`
- 有 IP 地址配置
- 有 peer 连接信息
- 有数据传输（rx/tx 不为 0）

### 2. 检查路由表

```bash
# 查看路由表
ip route show

# 查看默认路由
ip route show default
```

**问题**：如果没有通过 wg0 的路由，流量不会走 VPN。

**解决方案**：手动添加路由

```bash
# 添加默认路由（所有流量走 VPN）
sudo ip route add default dev wg0 table 200
sudo ip rule add from 172.16.0.2 table 200

# 或者只路由特定网段
sudo ip route add 0.0.0.0/1 dev wg0
sudo ip route add 128.0.0.0/1 dev wg0
```

### 3. 检查 DNS 配置

```bash
# 查看当前 DNS
cat /etc/resolv.conf

# 测试 DNS 解析
nslookup google.com
dig google.com
```

**问题**：DNS 没有配置为 Cloudflare 的 DNS 服务器。

**解决方案**：配置 DNS

```bash
# 临时修改 DNS
sudo bash -c 'echo "nameserver 1.1.1.1" > /etc/resolv.conf'
sudo bash -c 'echo "nameserver 1.0.0.1" >> /etc/resolv.conf'

# 或使用 systemd-resolved
sudo resolvectl dns wg0 1.1.1.1 1.0.0.1
```

### 4. 检查防火墙

```bash
# 查看 firewalld 状态（Fedora 默认）
sudo firewall-cmd --state

# 查看当前规则
sudo firewall-cmd --list-all

# 允许 wg0 接口
sudo firewall-cmd --zone=trusted --add-interface=wg0
sudo firewall-cmd --zone=trusted --add-interface=wg0 --permanent
```

### 5. 检查 IP 转发和 NAT

```bash
# 检查 IP 转发是否启用
sysctl net.ipv4.ip_forward

# 启用 IP 转发（如果需要）
sudo sysctl -w net.ipv4.ip_forward=1
```

### 6. 测试连接

```bash
# 测试 Cloudflare 连接
curl https://www.cloudflare.com/cdn-cgi/trace/

# 测试 DNS
ping -c 4 1.1.1.1

# 测试外网（通过 IP）
ping -c 4 8.8.8.8

# 测试外网（通过域名）
ping -c 4 google.com

# 查看数据包传输
sudo wg show wg0 transfer
```

### 7. 查看 WireGuard 日志

```bash
# 查看系统日志
sudo journalctl -u wg-quick@wg0 -f

# 或查看 dmesg
sudo dmesg | grep -i wireguard
```

## 完整的连接脚本

创建一个自动配置脚本 `setup-vpn.sh`：

```bash
#!/bin/bash

# 检查是否为 root
if [ "$EUID" -ne 0 ]; then 
    echo "请使用 sudo 运行"
    exit 1
fi

INTERFACE="wg0"
VPN_IP="172.16.0.2"  # 从配置文件中获取

echo "=== 配置 VPN 路由和 DNS ==="

# 1. 检查接口状态
echo "1. 检查接口..."
if ! ip link show $INTERFACE &> /dev/null; then
    echo "错误: 接口 $INTERFACE 不存在"
    exit 1
fi

# 2. 确保接口 UP
echo "2. 启动接口..."
ip link set $INTERFACE up

# 3. 配置路由
echo "3. 配置路由..."
# 删除旧路由（如果存在）
ip route del default dev $INTERFACE 2>/dev/null || true

# 添加路由表
if ! grep -q "200 warp" /etc/iproute2/rt_tables; then
    echo "200 warp" >> /etc/iproute2/rt_tables
fi

# 添加路由规则
ip route add default dev $INTERFACE table warp
ip rule add from $VPN_IP table warp

# 或者使用简单的默认路由（可能会覆盖现有路由）
# ip route add 0.0.0.0/1 dev $INTERFACE
# ip route add 128.0.0.0/1 dev $INTERFACE

# 4. 配置 DNS
echo "4. 配置 DNS..."
if command -v resolvectl &> /dev/null; then
    # 使用 systemd-resolved
    resolvectl dns $INTERFACE 1.1.1.1 1.0.0.1
    resolvectl domain $INTERFACE "~."
else
    # 备份并修改 resolv.conf
    cp /etc/resolv.conf /etc/resolv.conf.backup
    cat > /etc/resolv.conf << EOF
nameserver 1.1.1.1
nameserver 1.0.0.1
nameserver 2606:4700:4700::1111
nameserver 2606:4700:4700::1001
EOF
fi

# 5. 配置防火墙
echo "5. 配置防火墙..."
if command -v firewall-cmd &> /dev/null; then
    firewall-cmd --zone=trusted --add-interface=$INTERFACE 2>/dev/null || true
fi

# 6. 测试连接
echo "6. 测试连接..."
sleep 2

echo ""
echo "=== 连接测试 ==="
echo -n "Ping Cloudflare DNS: "
if ping -c 1 -W 2 1.1.1.1 &> /dev/null; then
    echo "✓"
else
    echo "✗"
fi

echo -n "DNS 解析: "
if nslookup google.com 1.1.1.1 &> /dev/null; then
    echo "✓"
else
    echo "✗"
fi

echo -n "访问外网: "
if curl -s --max-time 5 https://www.google.com &> /dev/null; then
    echo "✓"
else
    echo "✗"
fi

echo ""
echo "=== 当前状态 ==="
echo "接口信息:"
ip addr show $INTERFACE | grep -E "inet|state"

echo ""
echo "路由信息:"
ip route show | grep $INTERFACE

echo ""
echo "DNS 信息:"
if command -v resolvectl &> /dev/null; then
    resolvectl status $INTERFACE
else
    cat /etc/resolv.conf | grep nameserver
fi

echo ""
echo "WireGuard 状态:"
wg show $INTERFACE

echo ""
echo "=== 完成 ==="
echo "如果仍然无法访问外网，请检查:"
echo "1. Zero Trust 策略是否允许访问"
echo "2. 本地防火墙规则"
echo "3. WireGuard peer 是否有数据传输"
```

使用方法：

```bash
# 1. 连接 VPN
sudo ./warp-cli connect

# 2. 运行配置脚本
sudo bash setup-vpn.sh

# 3. 测试
curl https://www.cloudflare.com/cdn-cgi/trace/
```

## 自动化：修改 warp-cli connect

如果你想让 `warp-cli connect` 自动配置路由和 DNS，可以在代码中添加这些步骤。

## Fedora 特定问题

### NetworkManager 冲突

Fedora 使用 NetworkManager，可能会干扰手动配置：

```bash
# 让 NetworkManager 忽略 wg0
sudo nmcli device set wg0 managed no

# 或者完全禁用 NetworkManager（不推荐）
# sudo systemctl stop NetworkManager
```

### SELinux

如果启用了 SELinux，可能会阻止某些操作：

```bash
# 临时禁用 SELinux
sudo setenforce 0

# 查看 SELinux 日志
sudo ausearch -m avc -ts recent
```

### systemd-resolved

Fedora 使用 systemd-resolved 管理 DNS：

```bash
# 配置 DNS
sudo resolvectl dns wg0 1.1.1.1 1.0.0.1

# 设置为默认 DNS
sudo resolvectl domain wg0 "~."

# 查看状态
resolvectl status wg0
```

## 快速诊断命令

```bash
# 一键诊断
sudo bash -c '
echo "=== 接口状态 ==="
ip addr show wg0
echo ""
echo "=== WireGuard 状态 ==="
wg show wg0
echo ""
echo "=== 路由表 ==="
ip route | grep wg0
echo ""
echo "=== DNS 配置 ==="
cat /etc/resolv.conf
echo ""
echo "=== 连接测试 ==="
ping -c 2 1.1.1.1
echo ""
curl https://www.cloudflare.com/cdn-cgi/trace/
'
```

## 常见错误和解决方案

### 错误 1: "Network is unreachable"

**原因**: 没有配置路由

**解决**: 
```bash
sudo ip route add default dev wg0 metric 100
```

### 错误 2: DNS 解析失败

**原因**: DNS 没有配置

**解决**:
```bash
sudo resolvectl dns wg0 1.1.1.1
```

### 错误 3: 可以 ping 通 IP 但无法访问域名

**原因**: DNS 问题

**解决**:
```bash
# 测试 DNS
nslookup google.com 1.1.1.1

# 手动配置 DNS
sudo bash -c 'echo "nameserver 1.1.1.1" > /etc/resolv.conf'
```

### 错误 4: WireGuard 显示 0 数据传输

**原因**: 路由配置错误，流量没有走 VPN

**解决**:
```bash
# 强制所有流量走 VPN
sudo ip route add 0.0.0.0/1 dev wg0
sudo ip route add 128.0.0.0/1 dev wg0
```

## 持久化配置

要让配置在重启后保持，需要创建 systemd 服务或使用 NetworkManager 配置。

### 方法 1: systemd 服务

创建 `/etc/systemd/system/warp-vpn.service`:

```ini
[Unit]
Description=Cloudflare WARP VPN
After=network.target

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/path/to/warp-cli connect
ExecStartPost=/path/to/setup-vpn.sh
ExecStop=/path/to/warp-cli disconnect

[Install]
WantedBy=multi-user.target
```

启用：
```bash
sudo systemctl enable warp-vpn
sudo systemctl start warp-vpn
```

### 方法 2: NetworkManager dispatcher

创建 `/etc/NetworkManager/dispatcher.d/99-warp`:

```bash
#!/bin/bash
if [ "$1" = "wg0" ] && [ "$2" = "up" ]; then
    /path/to/setup-vpn.sh
fi
```

```bash
sudo chmod +x /etc/NetworkManager/dispatcher.d/99-warp
```
