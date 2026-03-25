#!/bin/bash

# 测试不同的 API 端点
TEAM="carmeltop"
TOKEN="eyJhbGciOiJSUzI1NiIsImtpZCI6IjFmNDBhOTEzMmMyMjFkNTM2MDBiNzNkZWY3MDU5MGNkODdmZDU2Y2IwY2U3MDk0MWVkMTNkMDczNmMyOGVhNGQifQ.eyJhdWQiOlsiNjI0Y2ZlMzg0ODA4YWE2Y2RlZTQyODAxYjkzYjRmMjYzMGZmYzBkYjBkYjE3NTEzZjYzZDM1YmIyOGI2MGQyOCJdLCJlbWFpbCI6ImNhcm1lbHRvcEBob3RtYWlsLmNvbSIsImV4cCI6MTc3NDM0MzUyNSwiaWF0IjoxNzc0MzQzNDY1LCJuYmYiOjE3NzQzNDM0NjUsImlzcyI6Imh0dHBzOi8vY2FybWVsdG9wLmNsb3VkZmxhcmVhY2Nlc3MuY29tIiwidHlwZSI6ImFwcCIsImlkZW50aXR5X25vbmNlIjoicGxQd1oxc2dkR2dIRk5xUCIsInN1YiI6IjA0MmJiYWExLWQ5ZTgtNTYwMi04MDhlLTI2ZWExNzhmZTJkOCIsIndhcnAiOnRydWUsImlwIjoiMjQwMDo4OTAyOjpmMDNjOjkyZmY6ZmUzZTo1MTc0IiwiYWNjb3VudF9pZCI6ImZmYjhiZThkZTNmMjQ0M2ExZDQ2YjRhN2UxMjY4NGM2IiwiaWRlbnRpdHkiOnsiZW1haWwiOiJjYXJtZWx0b3BAaG90bWFpbC5jb20ifSwiY291bnRyeSI6IkpQIn0.Ck_G8TbmAZeaIrCTH-eD7G_tOwKR8lRb2GDAWmwrhQitLeJxdQJD6JfsuuBzj7YIIXG7TFDyU91uXAh_d4jlTMwnAF7-eA6fGW7s3gF7oamtxrxX6sShx1SYrQft4ZVlEaMUwQqZjM6CPXvFDYbE4WRFDa12kYYEqn_7ZGnnCr-hOlnwaptp-yxnB6DUzm8g0FQJ54wIm1Ge82pXKkYIDR4gtdX0uwBASjyzVNTXzUaed14OOYibhKVKMX-YscgqcV10RhJfMw3HUEH6Wab-kynRGDVWITGQx5ynM_4eT_CaBBJivfGnDvV-2LUsun5gYz_HylraJ1n4Nw0tUWc_Cg"

echo "=== Testing different API endpoints ==="
echo ""

# 测试 1: /warp 端点
echo "Test 1: POST https://${TEAM}.cloudflareaccess.com/warp"
curl -v -X POST \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -H "User-Agent: okhttp/3.12.1" \
  -H "CF-Client-Version: a-6.3-1922" \
  -d '{"key":"test","tos":"2024-01-01T00:00:00Z","type":"Linux","model":"PC","locale":"en_US"}' \
  "https://${TEAM}.cloudflareaccess.com/warp" 2>&1 | head -50

echo ""
echo "==="
echo ""

# 测试 2: /warp/reg 端点
echo "Test 2: POST https://${TEAM}.cloudflareaccess.com/warp/reg"
curl -v -X POST \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -H "User-Agent: okhttp/3.12.1" \
  -H "CF-Client-Version: a-6.3-1922" \
  -d '{"key":"test","tos":"2024-01-01T00:00:00Z","type":"Linux","model":"PC","locale":"en_US"}' \
  "https://${TEAM}.cloudflareaccess.com/warp/reg" 2>&1 | head -50

echo ""
echo "==="
echo ""

# 测试 3: 使用 Cf-Access-Jwt-Assertion header
echo "Test 3: POST with Cf-Access-Jwt-Assertion header"
curl -v -X POST \
  -H "Cf-Access-Jwt-Assertion: ${TOKEN}" \
  -H "Content-Type: application/json" \
  -H "User-Agent: okhttp/3.12.1" \
  -H "CF-Client-Version: a-6.3-1922" \
  -d '{"key":"test","tos":"2024-01-01T00:00:00Z","type":"Linux","model":"PC","locale":"en_US"}' \
  "https://${TEAM}.cloudflareaccess.com/warp" 2>&1 | head -50
