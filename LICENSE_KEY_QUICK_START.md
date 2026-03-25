# License Key 快速获取指南

## 从哪里获取 License Key？

License Key 从 **Cloudflare 1.1.1.1 官方应用**中获取。

### Android 设备

1. 打开 **1.1.1.1** 应用
2. 点击右上角 **菜单按钮**（☰）
3. 选择 **Account（账户）** → **Key（密钥）**
4. 复制显示的密钥

### iOS 设备

1. 打开 **1.1.1.1** 应用
2. 点击右上角 **菜单按钮**
3. 选择 **Account（账户）** → **Key（密钥）**
4. 复制显示的密钥

### Windows/macOS 桌面应用

1. 打开 **Cloudflare WARP** 应用
2. 点击 **设置图标**（⚙️）
3. 选择 **Account（账户）**
4. 找到并复制 **License Key**

## License Key 格式

```
xxxxxxxx-xxxxxxxx-xxxxxxxx
```

示例（非真实）：`a1b2c3d4-e5f6g7h8-i9j0k1l2`

## 使用方法

### 推荐流程（避免 Bug）

```bash
# 1. 注册新账户
warp-cli register

# 2. 立即绑定 license key（重要：不要先连接！）
warp-cli update --license-key "YOUR_LICENSE_KEY"

# 3. 连接 VPN
warp-cli connect

# 4. 验证 Warp+ 状态
warp-cli trace
# 应该显示：warp=plus
```

## 重要提示

⚠️ **只支持官方购买的订阅**
- 通过 1.1.1.1 应用内购买的订阅才有效
- 推荐奖励获得的 Warp+ 不被支持

⚠️ **设备限制**
- 每个账户最多绑定 5 台设备
- 超过限制需要先移除旧设备

⚠️ **避免 Cloudflare Bug**
- 新账户注册后，先绑定 license key
- 不要先连接 VPN 再绑定 key
- 否则可能无法获得 Warp+ 状态

## 没有 Warp+ 订阅？

免费版 Warp 也可以使用：

```bash
warp-cli register
warp-cli connect
```

## 购买 Warp+ 订阅

在 1.1.1.1 应用中：
- 点击 **Warp+** 升级选项
- 价格通常为 $4.99/月 或 $49.99/年
- 通过 App Store 或 Google Play 购买

## 更多信息

详细指南请查看：[LICENSE_KEY_GUIDE.md](LICENSE_KEY_GUIDE.md)
