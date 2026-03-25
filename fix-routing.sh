#!/bin/bash

echo "=== 修复 WireGuard 路由 ==="
echo ""

# 获取默认网关和接口
DEFAULT_GW=$(ip route | grep default | awk '{print $3}' | head -1)
DEFAULT_IF=$(ip route | grep default | awk '{print $5}' | head -1)

echo "默认网关: $DEFAULT_GW"
echo "默认接口: $DEFAULT_IF"
echo ""

# 获取 WireGuard endpoint
if [ ! -f "warp-cli-profile.conf" ]; then
    echo "错误: warp-cli-profile.conf 不存在"
    exit 1
fi

ENDPOINT=$(grep "^Endpoint" warp-cli-profile.conf | awk '{print $3}')
ENDPOINT_IP=$(echo $ENDPOINT | cut -d: -f1)

echo "Endpoint: $ENDPOINT"
echo "Endpoint IP: $ENDPOINT_IP"
echo ""

# 解析 hostname（如果需要）
if ! echo $ENDPOINT_IP | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$'; then
    echo "解析 hostname: $ENDPOINT_IP"
    RESOLVED_IP=$(getent hosts $ENDPOINT_IP | awk '{print $1}' | head -1)
    if [ -n "$RESOLVED_IP" ]; then
        ENDPOINT_IP=$RESOLVED_IP
        echo "解析结果: $ENDPOINT_IP"
    else
        echo "警告: 无法解析 hostname"
    fi
    echo ""
fi

# 删除旧路由（如果存在）
echo "删除旧路由..."
sudo ip route del $ENDPOINT_IP 2>/dev/null || true

# 添加新路由
echo "添加 endpoint 路由..."
sudo ip route add $ENDPOINT_IP via $DEFAULT_GW dev $DEFAULT_IF

if [ $? -eq 0 ]; then
    echo "✓ 路由添加成功"
else
    echo "✗ 路由添加失败"
    exit 1
fi

echo ""
echo "验证路由:"
ip route | grep $ENDPOINT_IP

echo ""
echo "测试 ping endpoint:"
ping -c 3 $ENDPOINT_IP

echo ""
echo "=== 完成 ==="
echo ""
echo "现在重新连接 VPN:"
echo "  sudo ./warp-cli disconnect -i wg0"
echo "  sudo ./warp-cli connect"
echo ""
echo "然后测试:"
echo "  sudo wg show wg0"
echo "  ping -c 4 1.1.1.1"
