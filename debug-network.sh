#!/bin/bash

echo "=== WireGuard 网络调试 ==="
echo ""

# 获取 endpoint 信息
ENDPOINT=$(sudo wg show wg0 | grep "endpoint:" | awk '{print $2}')
ENDPOINT_IP=$(echo $ENDPOINT | cut -d: -f1)
ENDPOINT_PORT=$(echo $ENDPOINT | cut -d: -f2)

echo "Endpoint: $ENDPOINT"
echo "IP: $ENDPOINT_IP"
echo "Port: $ENDPOINT_PORT"
echo ""

# 1. 测试 UDP 连通性
echo "=== 测试 1: UDP 连通性 ==="
echo "发送 UDP 测试包..."
timeout 3 bash -c "echo 'test' > /dev/udp/$ENDPOINT_IP/$ENDPOINT_PORT" 2>&1
if [ $? -eq 0 ]; then
    echo "✓ UDP 端口可达"
else
    echo "✗ UDP 端口不可达或超时"
fi
echo ""

# 2. 检查出站接口
echo "=== 测试 2: 检查出站接口 ==="
DEFAULT_IFACE=$(ip route | grep default | awk '{print $5}' | head -1)
echo "默认网络接口: $DEFAULT_IFACE"
echo ""

# 3. 抓包分析
echo "=== 测试 3: 抓包分析 ==="
echo "开始抓包 10 秒..."
echo "同时会发送 ping 到 1.1.1.1"
echo ""

# 启动抓包
PCAP_FILE="/tmp/wg-debug-$(date +%s).pcap"
timeout 10 sudo tcpdump -i $DEFAULT_IFACE -w $PCAP_FILE udp port $ENDPOINT_PORT &
TCPDUMP_PID=$!

sleep 2

# 发送一些流量
ping -c 3 1.1.1.1 &> /dev/null &

# 等待抓包完成
wait $TCPDUMP_PID 2>/dev/null

echo "抓包完成，分析结果..."
echo ""

# 分析抓包文件
if [ -f "$PCAP_FILE" ]; then
    PACKET_COUNT=$(sudo tcpdump -r $PCAP_FILE 2>/dev/null | wc -l)
    echo "捕获的数据包总数: $PACKET_COUNT"
    
    if [ $PACKET_COUNT -gt 0 ]; then
        echo ""
        echo "数据包详情:"
        sudo tcpdump -r $PCAP_FILE -n -vv 2>/dev/null | head -50
        
        # 统计方向
        OUTBOUND=$(sudo tcpdump -r $PCAP_FILE -n 2>/dev/null | grep "> $ENDPOINT_IP" | wc -l)
        INBOUND=$(sudo tcpdump -r $PCAP_FILE -n 2>/dev/null | grep "$ENDPOINT_IP >" | wc -l)
        
        echo ""
        echo "出站数据包 (到 $ENDPOINT_IP): $OUTBOUND"
        echo "入站数据包 (从 $ENDPOINT_IP): $INBOUND"
        
        if [ $OUTBOUND -gt 0 ] && [ $INBOUND -eq 0 ]; then
            echo ""
            echo "❌ 问题确认: 只有出站流量，没有入站流量"
            echo "   这表明服务器没有响应"
        fi
    else
        echo "❌ 没有捕获到任何数据包"
        echo "   这可能表明 WireGuard 没有发送数据"
    fi
    
    rm -f $PCAP_FILE
else
    echo "❌ 抓包文件不存在"
fi

echo ""

# 4. 检查 NAT 和路由
echo "=== 测试 4: NAT 和路由检查 ==="
echo "路由表:"
ip route show
echo ""

echo "NAT 规则:"
sudo iptables -t nat -L -n -v | grep -v "^Chain\|^target" | head -20
echo ""

# 5. 检查本地防火墙规则
echo "=== 测试 5: 本地防火墙规则 ==="
echo "INPUT 链:"
sudo iptables -L INPUT -n -v | head -20
echo ""

echo "OUTPUT 链:"
sudo iptables -L OUTPUT -n -v | head -20
echo ""

# 6. 测试直接 ping endpoint
echo "=== 测试 6: Ping Endpoint ==="
ping -c 3 $ENDPOINT_IP
echo ""

# 7. 检查 WireGuard 内核模块
echo "=== 测试 7: WireGuard 内核模块 ==="
if lsmod | grep -q wireguard; then
    echo "✓ WireGuard 内核模块已加载"
    lsmod | grep wireguard
else
    echo "⚠ WireGuard 内核模块未加载（使用 wireguard-go）"
fi
echo ""

# 8. 检查系统日志
echo "=== 测试 8: 系统日志 ==="
echo "最近的 WireGuard 相关日志:"
sudo journalctl -n 50 | grep -i wireguard || echo "没有找到相关日志"
echo ""

# 9. 建议
echo "=== 诊断建议 ==="
echo ""

if [ $INBOUND -eq 0 ] && [ $OUTBOUND -gt 0 ]; then
    echo "根据抓包分析，问题是服务器没有响应。"
    echo ""
    echo "可能的原因:"
    echo "  1. ISP 阻止了 UDP 流量"
    echo "  2. 路由器/网关阻止了 UDP"
    echo "  3. Cloudflare 服务器拒绝连接"
    echo "  4. 密钥不匹配"
    echo ""
    echo "建议尝试:"
    echo "  1. 使用手机热点测试（排除 ISP/路由器问题）"
    echo "  2. 检查路由器是否有 UDP 限制"
    echo "  3. 尝试不同的 endpoint（如果可能）"
    echo "  4. 联系 ISP 确认是否限制 VPN"
fi

echo ""
echo "=== 完整 WireGuard 状态 ==="
sudo wg show wg0
