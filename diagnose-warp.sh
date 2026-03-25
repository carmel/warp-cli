#!/bin/bash

echo "=== Cloudflare WARP 连接诊断 ==="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_ok() {
    echo -e "${GREEN}✓${NC} $1"
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# 1. 检查接口状态
echo "1. 检查 WireGuard 接口"
if ip link show wg0 &> /dev/null; then
    check_ok "接口 wg0 存在"
    
    if ip link show wg0 | grep -q "UP"; then
        check_ok "接口状态: UP"
    else
        check_fail "接口状态: DOWN"
    fi
    
    # 检查 IP 地址
    IP=$(ip addr show wg0 | grep "inet " | awk '{print $2}')
    if [ -n "$IP" ]; then
        check_ok "IP 地址: $IP"
    else
        check_fail "没有配置 IP 地址"
    fi
else
    check_fail "接口 wg0 不存在"
    exit 1
fi

echo ""

# 2. 检查 WireGuard 配置
echo "2. 检查 WireGuard 配置"
if command -v wg &> /dev/null; then
    WG_OUTPUT=$(sudo wg show wg0 2>&1)
    
    if echo "$WG_OUTPUT" | grep -q "endpoint:"; then
        ENDPOINT=$(echo "$WG_OUTPUT" | grep "endpoint:" | awk '{print $2}')
        check_ok "Endpoint: $ENDPOINT"
    else
        check_fail "没有配置 endpoint"
    fi
    
    if echo "$WG_OUTPUT" | grep -q "latest handshake:"; then
        HANDSHAKE=$(echo "$WG_OUTPUT" | grep "latest handshake:" | cut -d: -f2-)
        check_ok "最近握手:$HANDSHAKE"
    else
        check_warn "从未成功握手"
    fi
    
    # 检查数据传输
    TRANSFER=$(echo "$WG_OUTPUT" | grep "transfer:")
    RX=$(echo "$TRANSFER" | awk '{print $2, $3}')
    TX=$(echo "$TRANSFER" | awk '{print $5, $6}')
    
    echo "  接收: $RX"
    echo "  发送: $TX"
    
    if echo "$RX" | grep -q "0 B"; then
        check_fail "没有接收到任何数据"
    else
        check_ok "有数据接收"
    fi
else
    check_fail "wg 命令不可用"
fi

echo ""

# 3. 检查路由
echo "3. 检查路由配置"
if ip route | grep -q "wg0"; then
    check_ok "存在通过 wg0 的路由"
    ip route | grep wg0 | while read line; do
        echo "  $line"
    done
else
    check_fail "没有通过 wg0 的路由"
fi

echo ""

# 4. 检查 DNS
echo "4. 检查 DNS 配置"
if command -v resolvectl &> /dev/null; then
    DNS_STATUS=$(resolvectl status wg0 2>&1)
    if echo "$DNS_STATUS" | grep -q "DNS Servers:"; then
        DNS=$(echo "$DNS_STATUS" | grep "DNS Servers:" | cut -d: -f2)
        check_ok "DNS 服务器:$DNS"
    else
        check_warn "没有配置 DNS"
    fi
else
    if grep -q "1.1.1.1" /etc/resolv.conf; then
        check_ok "DNS 配置在 /etc/resolv.conf"
    else
        check_warn "DNS 可能未正确配置"
    fi
fi

echo ""

# 5. 检查防火墙
echo "5. 检查防火墙"
if command -v firewall-cmd &> /dev/null; then
    if sudo firewall-cmd --state 2>&1 | grep -q "running"; then
        check_warn "firewalld 正在运行"
        
        # 检查 wg0 是否在信任区域
        if sudo firewall-cmd --zone=trusted --list-interfaces | grep -q "wg0"; then
            check_ok "wg0 在 trusted 区域"
        else
            check_fail "wg0 不在 trusted 区域"
            echo "  运行: sudo firewall-cmd --zone=trusted --add-interface=wg0"
        fi
    else
        check_ok "firewalld 未运行"
    fi
fi

# 检查 iptables
if sudo iptables -L -n | grep -q "DROP.*wg0\|REJECT.*wg0"; then
    check_fail "iptables 可能阻止了 wg0 流量"
else
    check_ok "iptables 没有明显的阻止规则"
fi

echo ""

# 6. 网络连接测试
echo "6. 网络连接测试"

# 测试 endpoint 连通性
ENDPOINT_IP=$(sudo wg show wg0 | grep "endpoint:" | awk '{print $2}' | cut -d: -f1)
ENDPOINT_PORT=$(sudo wg show wg0 | grep "endpoint:" | awk '{print $2}' | cut -d: -f2)

if [ -n "$ENDPOINT_IP" ]; then
    echo -n "  测试 endpoint ($ENDPOINT_IP:$ENDPOINT_PORT)... "
    if timeout 3 bash -c "echo > /dev/udp/$ENDPOINT_IP/$ENDPOINT_PORT" 2>/dev/null; then
        echo "可达"
    else
        echo "不可达或超时"
    fi
fi

# 测试 DNS
echo -n "  测试 DNS 解析 (google.com)... "
if timeout 3 nslookup google.com 1.1.1.1 &> /dev/null; then
    check_ok "成功"
else
    check_fail "失败"
fi

# 测试 ping
echo -n "  测试 ping 1.1.1.1... "
if timeout 3 ping -c 1 1.1.1.1 &> /dev/null; then
    check_ok "成功"
else
    check_fail "失败"
fi

# 测试 HTTP
echo -n "  测试 HTTP (cloudflare.com)... "
if timeout 5 curl -s https://www.cloudflare.com/cdn-cgi/trace/ &> /dev/null; then
    check_ok "成功"
else
    check_fail "失败"
fi

echo ""

# 7. 建议
echo "=== 诊断建议 ==="
echo ""

if ! echo "$WG_OUTPUT" | grep -q "latest handshake:"; then
    echo "❌ 问题: WireGuard 握手从未成功"
    echo ""
    echo "可能原因:"
    echo "  1. Endpoint 无法访问"
    echo "  2. 防火墙阻止 UDP 流量"
    echo "  3. 密钥配置错误"
    echo "  4. Zero Trust 策略限制"
    echo ""
    echo "建议操作:"
    echo "  1. 检查防火墙: sudo firewall-cmd --zone=trusted --add-interface=wg0"
    echo "  2. 检查 UDP 出站: sudo tcpdump -i wlp3s0 udp port $ENDPOINT_PORT -n"
    echo "  3. 重新注册: ./warp-cli register-team --token \"...\""
    echo "  4. 检查 Zero Trust 设备策略"
fi

if echo "$RX" | grep -q "0 B" && ! echo "$TX" | grep -q "0 B"; then
    echo "❌ 问题: 只有发送没有接收"
    echo ""
    echo "这通常表示:"
    echo "  - 握手包发出但服务器没有响应"
    echo "  - 可能是防火墙阻止了入站 UDP"
    echo "  - 或者 NAT 配置问题"
    echo ""
    echo "建议操作:"
    echo "  1. 临时关闭防火墙测试: sudo systemctl stop firewalld"
    echo "  2. 检查路由器 NAT 设置"
    echo "  3. 尝试不同的网络环境"
fi

if ! ip route | grep -q "wg0"; then
    echo "❌ 问题: 没有配置路由"
    echo ""
    echo "建议操作:"
    echo "  sudo ip route add 0.0.0.0/1 dev wg0"
    echo "  sudo ip route add 128.0.0.0/1 dev wg0"
fi

echo ""
echo "=== 完整 WireGuard 状态 ==="
sudo wg show wg0

echo ""
echo "=== 配置文件内容 ==="
if [ -f "warp-cli-profile.conf" ]; then
    cat warp-cli-profile.conf
else
    echo "配置文件不存在"
fi
