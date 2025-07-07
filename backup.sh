#!/bin/bash
set -e

# 备份目标
BACKUP_NAME="gemini_backup_$(date +%Y%m%d_%H%M%S).tar.gz"

tar czvf "$BACKUP_NAME" \
    config.yaml \
    data/ \
    font/ \
    *.pem \
    /etc/nginx/sites-available/gemini \
    /etc/nginx/sites-enabled/gemini

echo "备份完成: $BACKUP_NAME"
