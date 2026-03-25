#!/bin/bash

echo "=== 快速连接脚本（使用 wg-quick）==="
echo ""

# 1. 确保配置文件存在
if [ ! -f "warp-cli-profile.conf" ]; then
    echo "✗ warp-cli-profile.conf 不存在"
    echo "请先运行: ./warp-cli register 或 ./warp-cli register-team"
    exit 1
fi

# 2. 复制为 wg0.conf
cp warp-cli-profile.conf wg0.conf
echo "✓ 配置文件已准备: wg0.conf"

# 3. 断开现有连接
echo "断开现有连接..."
sudo wg-quick down wg0 2>/dev/null || true

# 4. 使用绝对路径启动
echo "启动 WireGuard..."
sudo wg-quick up $(pwd)/wg0.conf

if [ $? -ne 0 ]; then
    echo "✗ 启动失败"
    exit 1
fi

echo ""
echo "✓ WireGuard 已启动"
echo "⏳ 等待握手（最多 30 秒）..."
echo ""

# 5. 等待握手
for i in {1..15}; do
    sleep 2
    
    if sudo wg show wg0 | grep -q "latest handshake:"; then
        echo "✓ 握手成功！"
        echo ""
        sudo wg show wg0
        echo ""
        
        # 测试连接
        echo "测试连接..."
        if ping -c 2 -W 3 1.1.1.1 &>/dev/null; then
            echo "✓ Ping 1.1.1.1 成功"
            
            if curl -s --max-time 5 https://www.google.com &>/dev/null; then
                echo "✓ 可以访问 Google"
                echo ""
                echo "=== 连接成功！==="
                exit 0
            fi
        fi
        
        echo "⚠️  握手成功但无法访问外网"
        echo "   可能需要等待更长时间"
        break
    fi
    
    if [ $((i % 3)) -eq 0 ]; then
        echo "  等待中... ($((i*2))/30 秒)"
    fi
done

echo ""
echo "当前状态:"
sudo wg show wg0
echo ""
echo "如果仍无法连接，请运行: sudo ./test-with-retry.sh"
