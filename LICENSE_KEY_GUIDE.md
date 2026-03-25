# Cloudflare Warp+ License Key 获取指南

## 什么是 License Key？

License Key（许可证密钥）是 Cloudflare Warp+ 订阅的唯一标识符。通过这个密钥，你可以：
- 将多个设备（最多 5 台）绑定到同一个 Warp+ 账户
- 在不同设备间共享 Warp+ 订阅
- 享受 Warp+ 的更快速度和优先路由

## 重要说明

⚠️ **只支持官方购买的订阅**

> 只有通过官方 1.1.1.1 应用直接购买的 Warp+ 订阅才被支持。
> 通过其他方式获得的密钥（如推荐奖励）将不起作用，也不会得到支持。

## 获取方式

### 方法 1：从 Android 设备获取

1. 在 Android 设备上打开 **1.1.1.1** 应用
2. 点击右上角的 **汉堡菜单按钮**（三条横线图标）
3. 导航到：**Account（账户）** > **Key（密钥）**
4. 你会看到一个类似这样的密钥：`xxxxx-xxxxx-xxxxx`
5. 点击复制或手动记录下来

### 方法 2：从 iOS 设备获取

1. 在 iOS 设备上打开 **1.1.1.1** 应用
2. 点击右上角的 **菜单按钮**
3. 选择 **Account（账户）**
4. 点击 **Key（密钥）**
5. 复制显示的许可证密钥

### 方法 3：从 Windows/macOS 桌面应用获取

1. 打开 Cloudflare WARP 桌面应用
2. 点击设置图标（齿轮图标）
3. 选择 **Preferences（偏好设置）** 或 **Account（账户）**
4. 找到 **License Key** 或 **Account Key** 选项
5. 复制密钥

## License Key 格式

标准的 Warp+ License Key 格式为：
```
xxxxxxxx-xxxxxxxx-xxxxxxxx
```

- 由三组字符组成
- 每组 8 个字符
- 使用连字符（-）分隔
- 包含字母和数字

示例（非真实密钥）：
```
a1b2c3d4-e5f6g7h8-i9j0k1l2
```

## 使用 License Key

### 首次使用（推荐流程）

⚠️ **重要：避免 Cloudflare 的已知 Bug**

由于 Cloudflare 的一个 bug，如果账户曾经连接过 Warp VPN，绑定新的 license key 后可能无法获得 Warp+ 状态。

**正确的操作顺序：**

```bash
# 1. 注册新账户
warp-cli register

# 2. 立即绑定 license key（不要运行其他命令）
warp-cli update --license-key "YOUR_LICENSE_KEY_GOES_HERE"

# 3. 生成配置或连接
warp-cli connect
```

### 更新现有账户的 License Key

如果你已经有账户，想要更换 license key：

```bash
# 更新 license key
warp-cli update --license-key "YOUR_NEW_LICENSE_KEY"

# 重新生成配置
warp-cli generate

# 或者直接连接
warp-cli connect
```

### 验证 Warp+ 是否生效

连接 VPN 后，运行：

```bash
warp-cli trace
```

查看输出的最后一行：
- `warp=on` - 表示使用的是免费版 Warp
- `warp=plus` - 表示使用的是 Warp+ 订阅 ✅

## 设备管理

### 查看已绑定的设备

```bash
warp-cli status
```

这会显示：
- 当前设备信息
- 账户类型（free 或 plus）
- 绑定的设备列表

### 设备数量限制

- 每个 Warp+ 账户最多可以绑定 **5 台设备**
- 如果超过限制，需要先移除旧设备

### 移除设备

#### 方法 1：通过 1.1.1.1 应用移除

1. 打开 1.1.1.1 应用
2. 进入 **Account（账户）** > **Devices（设备）**
3. 选择要移除的设备
4. 点击 **Remove（移除）**

#### 方法 2：通过 warp-cli 移除

```bash
# 首先查看设备列表，获取设备 ID
warp-cli status

# 移除指定设备（使用设备 ID）
warp-cli update --remove "device-id-here"

# 停用设备（不删除）
warp-cli update --deactivate "device-id-here"

# 重新激活设备
warp-cli update --activate "device-id-here"
```

## 常见问题

### Q1: 我没有 Warp+ 订阅，可以使用 warp-cli 吗？

**A:** 可以！即使没有 Warp+ 订阅，你也可以使用免费版的 Warp。只需：
```bash
warp-cli register
warp-cli connect
```

### Q2: 通过推荐获得的 Warp+ 可以用吗？

**A:** 不可以。根据官方说明，只有直接购买的订阅才被支持。推荐奖励获得的 Warp+ 可能无法正常工作。

### Q3: 绑定 license key 后仍显示 warp=on 而不是 warp=plus

**A:** 这是 Cloudflare 的已知 bug。解决方法：
1. 注册一个全新的账户：`warp-cli register`
2. 立即绑定 license key（不要先连接 VPN）
3. 然后再连接

### Q4: License key 在哪里存储？

**A:** License key 存储在配置文件中：
- 默认位置：`warp-cli-account.toml`
- 可以通过 `--config` 参数指定其他位置

查看配置文件：
```bash
cat warp-cli-account.toml
```

### Q5: 可以在多台电脑上使用同一个 license key 吗？

**A:** 可以，但有限制：
- 最多 5 台设备同时绑定
- 每台设备都需要运行 `warp-cli update --license-key "YOUR_KEY"`
- 超过 5 台需要先移除旧设备

### Q6: License key 会过期吗？

**A:** 取决于你的订阅类型：
- 如果是按月/年订阅，到期后需要续费
- 如果是一次性购买，通常不会过期
- 可以在 1.1.1.1 应用中查看订阅状态

### Q7: 忘记了 license key 怎么办？

**A:** 在任何已登录的设备上：
1. 打开 1.1.1.1 应用
2. 进入 Account > Key
3. 重新查看或复制密钥

### Q8: 更换 license key 会影响现有连接吗？

**A:** 会的。更换 license key 后需要：
1. 断开当前连接：`warp-cli disconnect`
2. 重新生成配置：`warp-cli generate`
3. 重新连接：`warp-cli connect`

## 购买 Warp+ 订阅

如果你还没有 Warp+ 订阅，可以通过以下方式购买：

1. **通过 1.1.1.1 应用购买**（推荐）
   - 打开应用
   - 点击 Warp+ 升级选项
   - 按照提示完成购买

2. **价格**（可能因地区而异）
   - 通常为 $4.99/月 或 $49.99/年
   - 某些地区可能有不同定价

3. **支付方式**
   - iOS: 通过 Apple App Store
   - Android: 通过 Google Play Store
   - 桌面应用: 通过应用内购买

## 安全提示

⚠️ **保护你的 License Key**

- 不要在公共场合分享你的 license key
- 不要将 license key 提交到公共代码仓库
- 如果 license key 泄露，可以在 1.1.1.1 应用中重置

## 示例：完整的 Warp+ 设置流程

```bash
# 1. 注册新账户
./warp-cli register
# 输出：Successfully registered new Cloudflare Warp account

# 2. 立即绑定 Warp+ license key
./warp-cli update --license-key "a1b2c3d4-e5f6g7h8-i9j0k1l2"
# 输出：Updated license key detected, re-binding device to new account

# 3. 设置设备名称（可选）
./warp-cli update --name "My-Laptop"

# 4. 连接 VPN
./warp-cli connect
# 输出：✓ Successfully connected to Cloudflare Warp!

# 5. 验证 Warp+ 状态
./warp-cli trace
# 查看最后一行应该显示：warp=plus

# 6. 查看账户状态
./warp-cli status
# 输出会显示账户类型为 "plus"
```

## 相关命令参考

```bash
# 查看 update 命令的所有选项
warp-cli update --help

# 更新 license key
warp-cli update --license-key "YOUR_KEY"

# 更新设备名称
warp-cli update --name "My-Device"

# 移除设备
warp-cli update --remove "device-id"

# 停用设备
warp-cli update --deactivate "device-id"

# 激活设备
warp-cli update --activate "device-id"

# 组合使用
warp-cli update --license-key "YOUR_KEY" --name "My-Laptop"
```

## 故障排除

### 问题：提示 "invalid license key"

**可能原因：**
- License key 格式错误（检查是否有多余的空格或字符）
- License key 已过期
- License key 不是官方购买的

**解决方法：**
```bash
# 确保 license key 格式正确，使用引号包裹
warp-cli update --license-key "xxxxxxxx-xxxxxxxx-xxxxxxxx"
```

### 问题：提示 "maximum devices reached"

**解决方法：**
```bash
# 查看已绑定的设备
warp-cli status

# 移除不需要的设备
warp-cli update --remove "old-device-id"
```

### 问题：绑定后仍然是免费版

**解决方法：**
1. 注册全新账户
2. 在首次连接前绑定 license key
3. 参考上面的"避免 Cloudflare Bug"部分

## 总结

- License Key 从官方 1.1.1.1 应用获取
- 只支持官方购买的订阅
- 最多绑定 5 台设备
- 新账户应该先绑定 key 再连接
- 使用 `warp-cli trace` 验证 Warp+ 状态
