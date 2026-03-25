#!/bin/bash

echo "=== 测试 WireGuard 连接（带重试）==="
echo ""

MAX_RETRIES=10
RETRY_INTERVAL=5

# 确保 wg0 已启动
if ! ip link show wg0 &>/dev/null; then
    echo "启动 wg0..."
    sudo wg-quick up wg0
    if [ $? -ne 0 ]; then
        echo "✗ 无法启动 wg0"
        exit 1
    fi
fi

echo "等待 WireGuard 握手..."
echo "最多重试 $MAX_RETRIES 次，每次间隔 $RETRY_INTERVAL 秒"
echo ""

for i in $(seq 1 $MAX_RETRIES); do
    echo "尝试 $i/$MAX_RETRIES..."
    
    # 检查握手状态
    WG_STATUS=$(sudo wg show wg0)
    
    # 检查是否有握手
    if echo "$WG_STATUS" | grep -q "latest handshake:"; then
        echo "✓ 握手成功！"
        echo ""
        sudo wg show wg0
        echo ""
        
        # 测试连接
        echo "测试连接..."
        if ping -c 2 -W 3 1.1.1.1 &>/dev/null; then
            echo "✓ Ping 成功！"
            
            if curl -s --max-time 5 https://www.google.com &>/dev/null; then
                echo "✓ 可以访问 Google！"
                echo ""
                echo "=== 连接成功 ==="
                exit 0
            fi
        fi
    fi
    
    # 检查传输数据
    RECEIVED=$(echo "$WG_STATUS" | grep "transfer:" | awk '{print $2, $3}')
    SENT=$(echo "$WG_STATUS" | grep "transfer:" | awk '{print $5, $6}')
    
    echo "  接收: $RECEIVED"
    echo "  发送: $SENT"
    
    if [ "$i" -lt "$MAX_RETRIES" ]; then
        echo "  等待 $RETRY_INTERVAL 秒后重试..."
        sleep $RETRY_INTERVAL
    fi
    echo ""
done

echo "✗ 达到最大重试次数，仍未成功握手"
echo ""
echo "最终状态:"
sudo wg show wg0
