#!/bin/bash
set -e
if [ -z "$1" ]; then
  echo "用法: ./restore.sh 备份文件名.tar.gz"
  exit 1
fi

tar xzvf "$1"

# 检查并还原 nginx 配置
if [ -f ./gemini ]; then
  sudo mv -f ./gemini /etc/nginx/sites-available/gemini
  sudo ln -sf /etc/nginx/sites-available/gemini /etc/nginx/sites-enabled/gemini
  echo "已还原 nginx 配置到 /etc/nginx/sites-available/gemini"
  sudo nginx -t && sudo systemctl reload nginx
  echo "Nginx 配置已重载"
fi

echo "还原完成"
