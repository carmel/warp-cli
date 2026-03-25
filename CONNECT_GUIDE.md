# Warp-CLI 连接指南

## 新增功能

warp-cli 现在支持直接建立 VPN 连接，无需手动操作 wireguard-go 和 wg 命令。

## 前置要求

1. 编译 wireguard-go：

```bash
cd wireguard
make
cd ..
```

2. 确保系统已安装 `wg` 工具（WireGuard 命令行工具）：

```bash
# macOS
brew install wireguard-tools

# Linux (Ubuntu/Debian)
sudo apt install wireguard-tools

# Linux (Fedora/RHEL)
sudo dnf install wireguard-tools
```

## 使用方法

### 完整工作流程

```bash
# 1. 注册账户（首次使用）
./warp-cli register

# 2. （可选）绑定 Warp+ 许可证
./warp-cli update --license-key "YOUR_LICENSE_KEY"

# 3. 连接到 VPN
./warp-cli connect

# 4. 验证连接
./warp-cli trace

# 5. 断开连接
./warp-cli disconnect
```

### connect 命令详解

基本用法：

```bash
./warp-cli connect
```

自定义选项：

```bash
# 指定接口名称
./warp-cli connect -i wg0

# 指定配置文件
./warp-cli connect -p custom-profile.conf

# 前台运行（不后台化）
./warp-cli connect -f

# 组合使用
./warp-cli connect -i wg1 -p my-profile.conf
```

### disconnect 命令详解

基本用法：

```bash
./warp-cli disconnect
```

指定接口：

```bash
./warp-cli disconnect -i wg0
```

## 工作原理

`connect` 命令会自动执行以下步骤：

1. 检查账户是否有效
2. 如果配置文件不存在，自动生成
3. 查找 wireguard-go 可执行文件（优先使用 `./wireguard/wireguard-go`）
4. 启动 wireguard-go 创建网络接口
5. 使用 `wg setconf` 应用配置
6. 显示连接状态

`disconnect` 命令会：

1. 检查接口是否存在
2. 根据操作系统使用适当的方法删除接口：
   - Linux: `ip link del`
   - macOS/BSD: 删除控制套接字
   - Windows: 终止 wireguard-go 进程

## 平台特定说明

### macOS

- 默认接口名称：`utun`（系统会自动分配编号，如 utun3）
- 需要 root 权限：`sudo ./warp-cli connect`

### Linux

- 默认接口名称：`wg0`
- 需要 root 权限：`sudo ./warp-cli connect`
- 建议使用内核模块而非 wireguard-go

### Windows

- 默认接口名称：`wg0`
- 需要管理员权限

## 故障排除

### 接口已存在

```
Error: interface wg0 already exists
```

解决方法：

```bash
./warp-cli disconnect -i wg0
# 或使用不同的接口名称
./warp-cli connect -i wg1
```

### wireguard-go 未找到

```
Error: wireguard not found
```

解决方法：

```bash
cd wireguard && make && cd ..
```

### 权限不足

```
Error: Operation not permitted
```

解决方法：

```bash
sudo ./warp-cli connect
```

### wg 命令未找到

```
Error: wg setconf failed
```

解决方法：安装 wireguard-tools（见前置要求）

## 高级用法

### 多个连接

可以同时运行多个 VPN 连接（使用不同的接口和配置）：

```bash
# 连接 1
./warp-cli connect -i wg0 -p profile1.conf

# 连接 2（需要另一个账户）
./warp-cli connect -i wg1 -p profile2.conf --config account2.toml
```

### 前台调试

如果遇到问题，可以在前台运行查看详细日志：

```bash
LOG_LEVEL=debug ./warp-cli connect -f
```

### 手动操作（传统方式）

如果需要更多控制，仍可以使用传统方式：

```bash
# 生成配置
./warp-cli generate

# 手动启动
./wireguard/wireguard-go wg0
wg setconf wg0 warp-cli-profile.conf

# 手动停止
ip link del wg0
```

## 与原有命令的关系

- `generate` 命令仍然可用，用于仅生成配置文件
- `connect` 命令会在需要时自动调用 generate 的功能
- 所有其他命令（register, update, status, trace）保持不变

## 示例场景

### 场景 1：快速开始

```bash
./warp-cli register
./warp-cli connect
```

### 场景 2：使用 Warp+

```bash
./warp-cli register
./warp-cli update --license-key "xxxxx-xxxxx-xxxxx"
./warp-cli connect
./warp-cli trace  # 应该显示 warp=plus
```

### 场景 3：临时连接

```bash
# 前台运行，Ctrl+C 即可断开
./warp-cli connect -f
```
