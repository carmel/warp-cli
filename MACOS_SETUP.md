# macOS 设置指南

## 权限要求

在 macOS 上，创建 VPN 网络接口需要 root 权限。有两种方式运行 warp-cli：

### 方式 1: 使用 sudo（推荐）

```bash
sudo ./warp-cli connect
```

这是最简单的方式，但每次连接都需要输入密码。

### 方式 2: 设置 SUID 权限

如果你想避免每次都输入密码，可以给 wireguard-go 设置 SUID 权限：

```bash
# 给 wireguard-go 设置 root 所有者和 SUID 位
sudo chown root:wheel wireguard/wireguard-go
sudo chmod u+s wireguard/wireguard-go

# 现在可以不用 sudo 运行
./warp-cli connect
```

⚠️ 注意：设置 SUID 位有安全风险，只在你信任这个程序时使用。

## 完整使用流程

### 1. 注册账户

**消费者 WARP（免费）：**

```bash
./warp-cli register
```

**Zero Trust（企业）：**

```bash
# 使用 token URL
./warp-cli register-team --token "com.cloudflare.warp://yourteam.cloudflareaccess.com/auth?token=xxx"

# 或使用 Service Token
./warp-cli register-team --team yourteam --client-id xxx --client-secret yyy
```

### 2. 编译 wireguard-go

```bash
cd wireguard && make && cd ..
```

### 3. 连接 VPN

```bash
# 使用 sudo
sudo ./warp-cli connect

# 或者如果已设置 SUID
./warp-cli connect
```

### 4. 检查连接状态

```bash
# 查看接口状态
ifconfig utun

# 查看 WireGuard 状态
sudo wg show utun

# 测试连接
curl https://www.cloudflare.com/cdn-cgi/trace/
```

### 5. 断开连接

```bash
sudo ./warp-cli disconnect -i utun
```

## 常见问题

### Q: "interface utun was not created"

**原因**: 没有 root 权限

**解决**: 使用 `sudo ./warp-cli connect`

### Q: "wireguard not found"

**原因**: 没有编译 wireguard

**解决**:

```bash
cd wireguard && make && cd ..
```

### Q: "interface utun already exists"

**原因**: 接口已经在运行

**解决**:

```bash
# 先断开
sudo ./warp-cli disconnect -i utun

# 或者使用不同的接口名
sudo ./warp-cli connect -i utun2
```

### Q: "wg: command not found"

**原因**: 系统没有安装 WireGuard 工具

**解决**:

```bash
# 使用 Homebrew 安装
brew install wireguard-tools
```

## macOS 特定说明

### 接口命名

macOS 使用 `utun` 作为默认接口名称：

- `utun` - 系统自动分配编号（utun0, utun1, 等）
- 可以手动指定：`-i utun5`

### 网络配置

连接后，你可能需要配置 DNS 和路由：

```bash
# 查看当前 DNS
scutil --dns

# 查看路由表
netstat -rn

# 添加路由（如果需要）
sudo route add -net 10.0.0.0/8 -interface utun0
```

### 防火墙

如果启用了 macOS 防火墙，确保允许 wireguard-go：

1. 打开 系统偏好设置 → 安全性与隐私 → 防火墙
2. 点击 防火墙选项
3. 添加 wireguard-go 到允许列表

## 自动化脚本

创建一个便捷脚本 `connect.sh`：

```bash
#!/bin/bash

# 检查是否为 root
if [ "$EUID" -ne 0 ]; then
    echo "请使用 sudo 运行此脚本"
    exit 1
fi

# 连接
./warp-cli connect

# 配置 DNS（可选）
# networksetup -setdnsservers Wi-Fi 1.1.1.1 1.0.0.1

echo "VPN 已连接！"
```

使用：

```bash
chmod +x connect.sh
sudo ./connect.sh
```

## 性能优化

### MTU 设置

如果遇到连接问题，可以调整 MTU：

```bash
# 查看当前 MTU
ifconfig utun0 | grep mtu

# 设置 MTU
sudo ifconfig utun0 mtu 1420
```

### 保持连接

创建 LaunchDaemon 自动重连（高级）：

```xml
<!-- /Library/LaunchDaemons/com.warp-cli.plist -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.warp-cli</string>
    <key>ProgramArguments</key>
    <array>
        <string>/path/to/warp-cli</string>
        <string>connect</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

加载：

```bash
sudo launchctl load /Library/LaunchDaemons/com.warp-cli.plist
```

## 安全建议

1. 不要在不信任的网络上使用 SUID 权限
2. 定期更新 wireguard-go
3. 使用 Zero Trust 时启用设备认证
4. 定期检查连接日志
5. 使用强密码保护配置文件

## 故障排查

### 启用详细日志

```bash
# 前台运行查看详细输出
sudo ./warp-cli connect -f

# 查看 wireguard-go 日志
sudo ./wireguard/wireguard-go -f utun
```

### 重置配置

```bash
# 删除配置文件
rm warp-cli-account.toml warp-cli-profile.conf

# 重新注册
./warp-cli register
```

### 网络诊断

```bash
# 测试 DNS
nslookup cloudflare.com

# 测试连接
ping -c 4 1.1.1.1

# 追踪路由
traceroute 1.1.1.1

# 检查 Cloudflare 连接
curl https://www.cloudflare.com/cdn-cgi/trace/
```
