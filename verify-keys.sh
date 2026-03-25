#!/bin/bash

echo "=== 验证 WireGuard 密钥配置 ==="
echo ""

# 1. 从配置文件读取密钥
echo "1. 配置文件中的密钥:"
PRIVATE_KEY=$(grep "^PrivateKey" wg0.conf | awk '{print $3}')
PEER_PUBLIC_KEY=$(grep "^PublicKey" wg0.conf | awk '{print $3}')

echo "  Private Key: $PRIVATE_KEY"
echo "  Peer Public Key: $PEER_PUBLIC_KEY"
echo ""

# 2. 计算本地公钥
echo "2. 从私钥计算的公钥:"
LOCAL_PUBLIC_KEY=$(echo "$PRIVATE_KEY" | wg pubkey)
echo "  Local Public Key: $LOCAL_PUBLIC_KEY"
echo ""

# 3. 从账户配置读取
echo "3. 账户配置中的信息:"
if [ -f "warp-cli-account.toml" ]; then
    echo "  Device ID: $(grep device_id warp-cli-account.toml | cut -d'"' -f2)"
    echo "  Private Key: $(grep private_key warp-cli-account.toml | cut -d'"' -f2)"
    
    ACCOUNT_PRIVATE=$(grep private_key warp-cli-account.toml | cut -d'"' -f2)
    ACCOUNT_PUBLIC=$(echo "$ACCOUNT_PRIVATE" | wg pubkey)
    echo "  Calculated Public: $ACCOUNT_PUBLIC"
else
    echo "  warp-cli-account.toml 不存在"
fi
echo ""

# 4. 检查密钥是否匹配
echo "4. 密钥匹配检查:"
if [ "$PRIVATE_KEY" = "$ACCOUNT_PRIVATE" ]; then
    echo "  ✓ 配置文件和账户文件的私钥匹配"
else
    echo "  ✗ 配置文件和账户文件的私钥不匹配！"
    echo "    这可能是问题所在"
fi
echo ""

# 5. 测试密钥格式
echo "5. 密钥格式验证:"
if echo "$PRIVATE_KEY" | base64 -d &>/dev/null && [ ${#PRIVATE_KEY} -eq 44 ]; then
    echo "  ✓ 私钥格式正确（44字符 base64）"
else
    echo "  ✗ 私钥格式可能有问题"
fi

if echo "$PEER_PUBLIC_KEY" | base64 -d &>/dev/null && [ ${#PEER_PUBLIC_KEY} -eq 44 ]; then
    echo "  ✓ 对端公钥格式正确"
else
    echo "  ✗ 对端公钥格式可能有问题"
fi
echo ""

# 6. 建议
echo "=== 建议 ==="
if [ "$PRIVATE_KEY" != "$ACCOUNT_PRIVATE" ]; then
    echo "❌ 密钥不匹配！需要重新生成配置文件:"
    echo "   rm warp-cli-profile.conf wg0.conf"
    echo "   sudo ./warp-cli connect"
else
    echo "✓ 密钥配置正确"
    echo ""
    echo "问题可能在于:"
    echo "  1. Cloudflare 服务器拒绝连接（设备未授权）"
    echo "  2. Zero Trust 策略阻止"
    echo "  3. Endpoint 不可达"
    echo ""
    echo "建议:"
    echo "  1. 检查 Zero Trust Dashboard 中的设备状态"
    echo "  2. 尝试重新注册设备"
    echo "  3. 联系 Cloudflare 支持"
fi
