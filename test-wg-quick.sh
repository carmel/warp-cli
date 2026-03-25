#!/bin/bash

echo "=== 使用 wg-quick 测试 ==="
echo ""

CONFIG_FILE="warp-cli-profile.conf"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: $CONFIG_FILE 不存在"
    exit 1
fi

# wg-quick 要求配置文件名为 接口名.conf
WG_CONFIG="wg0.conf"

# 1. 复制配置文件
echo "1. 准备配置文件..."
cp $CONFIG_FILE $WG_CONFIG
if [ ! -f "$WG_CONFIG" ]; then
    echo "✗ 配置文件复制失败"
    exit 1
fi
echo "✓ 配置文件已复制: $WG_CONFIG"

# 2. 断开现有连接
echo "2. 断开现有连接..."
sudo wg-quick down wg0 2>/dev/null || true
sudo ip link del wg0 2>/dev/null || true

# 3. 使用 wg-quick 启动
echo "3. 使用 wg-quick 启动..."
echo "   配置文件: $(pwd)/$WG_CONFIG"
sudo wg-quick up $(pwd)/$WG_CONFIG

if [ $? -ne 0 ]; then
    echo "✗ wg-quick 启动失败"
    rm -f $WG_CONFIG
    exit 1
fi

echo "✓ wg-quick 启动成功"
echo ""

# 4. 等待握手
echo "4. 等待握手..."
sleep 5

# 5. 检查状态
echo "5. 检查 WireGuard 状态:"
sudo wg show wg0
echo ""

# 6. 测试连接
echo "6. 测试连接:"
echo -n "  Ping 1.1.1.1: "
if ping -c 2 -W 3 1.1.1.1 &> /dev/null; then
    echo "✓ 成功"
else
    echo "✗ 失败"
fi

echo -n "  DNS 解析: "
if nslookup google.com 1.1.1.1 &> /dev/null; then
    echo "✓ 成功"
else
    echo "✗ 失败"
fi

echo -n "  访问 Google: "
if timeout 5 curl -s https://www.google.com &> /dev/null; then
    echo "✓ 成功"
else
    echo "✗ 失败"
fi

echo ""
echo "=== 测试完成 ==="
echo ""
echo "如果成功，说明配置文件正确，问题在于我们的连接方式"
echo "如果失败，说明配置文件本身有问题"
echo ""
echo "断开连接: sudo wg-quick down wg0"
echo "清理: rm -f $WG_CONFIG"
