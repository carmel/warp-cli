# 自动密钥生成功能

## 概述

`warp-cli generate-config` 命令的模板模式现在支持自动生成 WireGuard 密钥对。

## 功能特性

### 默认行为（自动生成密钥）

当使用 `--template` 模式时，默认会自动调用 `wg genkey` 生成私钥，并自动计算对应的公钥：

```bash
warp-cli generate-config --template -o my-server.conf
```

输出示例：
```
✓ Successfully generated WireGuard configuration!
  Output: my-server.conf

======================================================================
Generated Keys:
======================================================================
Private Key: aBcDeFgHiJkLmNoPqRsTuVwXyZ1234567890ABCDEFGH=
Public Key:  XyZ9876543210ZyXwVuTsRqPoNmLkJiHgFeDcBa=
======================================================================

⚠️  IMPORTANT:
  - Keep your private key secure!
  - Share your PUBLIC KEY with the server administrator
  - The private key has been saved to the config file
```

生成的配置文件会自动填充私钥：
```ini
[Interface]
# Your client's private key (auto-generated)
PrivateKey = aBcDeFgHiJkLmNoPqRsTuVwXyZ1234567890ABCDEFGH=
...
```

### 禁用自动生成

如果你想使用占位符而不是自动生成密钥，可以使用 `--auto-key=false`：

```bash
warp-cli generate-config --template --auto-key=false -o my-server.conf
```

这将生成带有占位符的配置：
```ini
[Interface]
# Your client's private key (generate with: wg genkey)
PrivateKey = <YOUR_PRIVATE_KEY>
...
```

### 降级处理

如果系统中没有安装 `wg` 命令，会自动降级到占位符模式，并显示警告：

```
⚠️  Warning: Failed to auto-generate key: failed to generate private key (is 'wg' installed?): exec: "wg": executable file not found in $PATH
   Falling back to placeholder. You can generate keys manually with: wg genkey
```

## 技术实现

### 新增功能

1. **新增变量**: `autoKey bool` - 控制是否自动生成密钥（默认 true）
2. **新增标志**: `--auto-key` - 允许用户禁用自动生成
3. **新增函数**: `generateKeyPair()` - 封装密钥生成逻辑

### 密钥生成流程

```go
func generateKeyPair() (privateKey, publicKey string, err error) {
    // 1. 生成私钥
    cmd := exec.Command("wg", "genkey")
    output, err := cmd.Output()
    privateKey = strings.TrimSpace(string(output))
    
    // 2. 从私钥生成公钥
    cmd = exec.Command("wg", "pubkey")
    cmd.Stdin = strings.NewReader(privateKey)
    output, err = cmd.Output()
    publicKey = strings.TrimSpace(string(output))
    
    return privateKey, publicKey, nil
}
```

## 使用场景

### 场景 1: 快速生成可用配置（推荐）
```bash
# 自动生成密钥，只需填写服务器信息
warp-cli generate-config --template -o server.conf
# 编辑 server.conf，填写 PublicKey 和 Endpoint
# 将显示的公钥发送给服务器管理员
```

### 场景 2: 生成模板供后续编辑
```bash
# 生成纯模板，手动填写所有信息
warp-cli generate-config --template --auto-key=false -o template.conf
```

### 场景 3: 交互式模式（不受影响）
```bash
# 交互式模式仍然会询问是否生成密钥
warp-cli generate-config --interactive -o server.conf
```

## 安全提示

1. **私钥安全**: 生成的私钥会保存在配置文件中，请确保文件权限为 600
2. **公钥分享**: 只需要将公钥发送给服务器管理员，不要分享私钥
3. **密钥备份**: 建议备份生成的密钥对，以防配置文件丢失

## 兼容性

- 需要系统中安装 `wg` 命令（WireGuard 工具）
- 如果未安装，会自动降级到占位符模式
- 不影响其他模式（interactive、direct）的使用
