#!/bin/bash

# 获取验证码
echo "1. 获取验证码..."
CAPTCHA_RESPONSE=$(curl -s http://localhost:8080/api/captcha)
CAPTCHA_ID=$(echo $CAPTCHA_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
CAPTCHA_CODE=$(echo $CAPTCHA_RESPONSE | grep -o '"code":"[^"]*"' | cut -d'"' -f4)

echo "验证码ID: $CAPTCHA_ID"
echo "验证码: $CAPTCHA_CODE"

# 登录获取 token
echo -e "\n2. 登录..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"123456ok!\",\"captcha_id\":\"$CAPTCHA_ID\",\"captcha\":\"$CAPTCHA_CODE\"}")

echo "登录响应: $LOGIN_RESPONSE"

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "登录失败！"
  exit 1
fi

echo "Token: $TOKEN"

# 获取当前 RSS 配置
echo -e "\n3. 获取当前 RSS 配置..."
curl -s -X GET http://localhost:8080/api/admin/rss/config \
  -H "Authorization: Bearer $TOKEN"

# 更新 RSS 配置（启用）
echo -e "\n\n4. 启用 RSS 功能..."
curl -s -X PUT http://localhost:8080/api/admin/rss/config \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "title": "",
    "description": "",
    "author_name": "",
    "author_email": "",
    "item_limit": 20,
    "cache_duration": 3600
  }'

echo -e "\n\n5. 验证 RSS 配置..."
curl -s -X GET http://localhost:8080/api/admin/rss/config \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\n完成！"
