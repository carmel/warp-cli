# Warp-CLI 文档索引

## 📚 文档导航

### 快速开始

1. **[README.md](README.md)** - 项目主文档
   - 项目介绍和功能概览
   - 基本使用方法
   - 快速开始指南

2. **[WARP_VS_ZERO_TRUST.md](WARP_VS_ZERO_TRUST.md)** - WARP vs Zero Trust 说明 ⭐ 重要
   - warp-cli register 在做什么
   - 消费者 WARP vs 企业 Zero Trust
   - 为什么不需要 email
   - 使用场景对比

### VPN 连接

2. **[CONNECT_GUIDE.md](CONNECT_GUIDE.md)** - VPN 连接详细指南
   - `connect` 和 `disconnect` 命令使用
   - 前置要求和依赖安装
   - 平台特定说明
   - 故障排除
   - 高级用法和示例

### License Key（许可证密钥）

3. **[LICENSE_KEY_QUICK_START.md](LICENSE_KEY_QUICK_START.md)** - 快速获取指南 ⭐ 推荐新手
   - 从哪里获取 license key
   - 基本使用方法
   - 重要提示和注意事项

4. **[LICENSE_KEY_GUIDE.md](LICENSE_KEY_GUIDE.md)** - 完整详细指南
   - License key 详细说明
   - 多平台获取方法
   - 设备管理
   - 常见问题解答
   - 故障排除

5. **[docs/LICENSE_KEY_VISUAL_GUIDE.md](docs/LICENSE_KEY_VISUAL_GUIDE.md)** - 可视化图解指南
   - 图文并茂的获取步骤
   - 流程图和示意图
   - 错误场景和解决方案

### 技术文档

6. **[DIRECT_INTEGRATION_ANALYSIS.md](DIRECT_INTEGRATION_ANALYSIS.md)** - 技术实现分析
   - exec.Command vs 直接集成对比
   - 实现复杂度分析
   - 代码量估算
   - 推荐方案

---

## 🎯 根据需求选择文档

### 我是新用户，想快速开始

```
1. 阅读 README.md（了解项目）
2. 阅读 WARP_VS_ZERO_TRUST.md（理解 register 命令）
3. 阅读 LICENSE_KEY_QUICK_START.md（如果有 Warp+）
4. 阅读 CONNECT_GUIDE.md（连接 VPN）
```

### 我想使用 Warp+

```
1. LICENSE_KEY_QUICK_START.md（快速了解）
2. LICENSE_KEY_VISUAL_GUIDE.md（图文教程）
3. LICENSE_KEY_GUIDE.md（遇到问题时查阅）
```

### 我遇到了连接问题

```
1. CONNECT_GUIDE.md - 故障排除部分
2. LICENSE_KEY_GUIDE.md - 常见问题部分
```

### 我是开发者，想了解技术细节

```
1. DIRECT_INTEGRATION_ANALYSIS.md（实现分析）
2. 查看源码：cmd/connect/connect.go
3. 查看源码：wireguard/ 目录
```

---

## 📖 文档内容概览

### README.md
- ✅ 项目介绍
- ✅ 功能列表
- ✅ 编译和下载
- ✅ 基本命令使用
- ✅ License key 简要说明

### CONNECT_GUIDE.md
- ✅ 前置要求（wireguard-go、wg 工具）
- ✅ 完整工作流程
- ✅ connect 命令详解
- ✅ disconnect 命令详解
- ✅ 平台特定说明（macOS/Linux/Windows）
- ✅ 故障排除
- ✅ 高级用法（多连接、前台调试）

### LICENSE_KEY_QUICK_START.md
- ✅ 快速获取方法（Android/iOS/桌面）
- ✅ License key 格式
- ✅ 推荐使用流程
- ✅ 重要提示
- ✅ 购买信息

### LICENSE_KEY_GUIDE.md
- ✅ License key 详细说明
- ✅ 多平台获取方法
- ✅ 使用方法和最佳实践
- ✅ 设备管理（查看、移除、激活）
- ✅ 常见问题解答（8+ 个问题）
- ✅ 购买指南
- ✅ 安全提示
- ✅ 完整示例流程
- ✅ 故障排除

### LICENSE_KEY_VISUAL_GUIDE.md
- ✅ 移动设备获取步骤图
- ✅ 桌面应用获取步骤图
- ✅ License key 格式说明图
- ✅ 完整使用流程图
- ✅ 常见错误和解决方案图
- ✅ 账户类型对比表
- ✅ 设备管理示意图

### DIRECT_INTEGRATION_ANALYSIS.md
- ✅ 当前实现方式分析
- ✅ 直接集成实现方式
- ✅ 详细复杂度分析
- ✅ 代码量估算
- ✅ 依赖关系分析
- ✅ 功能对比
- ✅ 性能对比
- ✅ 推荐方案

---

## 🔍 快速查找

### 命令参考

| 命令 | 文档位置 |
|------|---------|
| `warp-cli register` | README.md |
| `warp-cli connect` | CONNECT_GUIDE.md |
| `warp-cli disconnect` | CONNECT_GUIDE.md |
| `warp-cli update --license-key` | LICENSE_KEY_GUIDE.md |
| `warp-cli generate` | README.md |
| `warp-cli status` | README.md |
| `warp-cli trace` | README.md |

### 常见问题

| 问题 | 文档位置 |
|------|---------|
| warp-cli register 在做什么？ | WARP_VS_ZERO_TRUST.md |
| 为什么不需要 email？ | WARP_VS_ZERO_TRUST.md |
| WARP 和 Zero Trust 有什么区别？ | WARP_VS_ZERO_TRUST.md |
| 如何获取 license key？ | LICENSE_KEY_QUICK_START.md |
| 如何连接 VPN？ | CONNECT_GUIDE.md |
| 为什么显示 warp=on 而不是 warp=plus？ | LICENSE_KEY_GUIDE.md - Q3 |
| 如何移除设备？ | LICENSE_KEY_GUIDE.md - 设备管理 |
| wireguard-go 未找到怎么办？ | CONNECT_GUIDE.md - 故障排除 |
| 接口已存在怎么办？ | CONNECT_GUIDE.md - 故障排除 |
| 设备数量超限怎么办？ | LICENSE_KEY_GUIDE.md - Q5 |
| 可以使用推荐获得的 Warp+ 吗？ | LICENSE_KEY_GUIDE.md - Q2 |

### 平台特定信息

| 平台 | 文档位置 |
|------|---------|
| macOS | CONNECT_GUIDE.md - 平台特定说明 |
| Linux | CONNECT_GUIDE.md - 平台特定说明 |
| Windows | CONNECT_GUIDE.md - 平台特定说明 |

---

## 📝 文档更新日志

### 2024-03-24
- ✅ 创建 CONNECT_GUIDE.md
- ✅ 创建 LICENSE_KEY_QUICK_START.md
- ✅ 创建 LICENSE_KEY_GUIDE.md
- ✅ 创建 LICENSE_KEY_VISUAL_GUIDE.md
- ✅ 创建 DIRECT_INTEGRATION_ANALYSIS.md
- ✅ 更新 README.md
- ✅ 创建 DOCUMENTATION_INDEX.md

---

## 🤝 贡献

如果你发现文档有错误或需要改进，欢迎：
1. 提交 Issue
2. 提交 Pull Request
3. 在讨论区提出建议

---

## 📞 获取帮助

1. **查看文档** - 首先查看相关文档
2. **搜索 Issues** - 查看是否有类似问题
3. **提交 Issue** - 如果问题未解决，创建新 Issue
4. **社区讨论** - 在讨论区寻求帮助

---

## 🔗 外部资源

- [Cloudflare WARP 官网](https://1.1.1.1/)
- [WireGuard 官网](https://www.wireguard.com/)
- [WireGuard Quick Start](https://www.wireguard.com/quickstart/)
- [WireGuard UAPI 协议](https://www.wireguard.com/xplatform/#configuration-protocol)

---

## 📄 许可证

本项目遵循 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

---

**提示：** 建议将此文档添加到浏览器书签，方便快速查找！
