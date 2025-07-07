#!/bin/bash
set -e

GO_VERSION_REQUIRED="1.24"
GO_TARBALL="go1.24.0.linux-amd64.tar.gz"
GO_URL="https://go.dev/dl/${GO_TARBALL}"

echo "==== 检查Go环境 ===="
if command -v go >/dev/null 2>&1; then
    GO_CUR_VER=$(go version | awk '{print $3}' | sed 's/go//')
    echo "当前Go版本: $GO_CUR_VER"
    if [[ $(printf '%s\n' "$GO_VERSION_REQUIRED" "$GO_CUR_VER" | sort -V | head -n1) != "$GO_VERSION_REQUIRED" ]]; then
        echo "Go版本过低，自动升级..."
        sudo rm -rf /usr/local/go
    else
        echo "Go版本满足要求"
    fi
else
    echo "未检测到Go，将自动安装"
fi

if ! command -v go >/dev/null 2>&1 || [[ $(go version | awk '{print $3}' | sed 's/go//') < "$GO_VERSION_REQUIRED" ]]; then
    echo "正在下载安装Go $GO_VERSION_REQUIRED ..."
    wget -q --show-progress $GO_URL
    sudo tar -C /usr/local -xzf $GO_TARBALL
    rm -f $GO_TARBALL
    if ! grep -q '/usr/local/go/bin' ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
    export PATH=$PATH:/usr/local/go/bin
fi

echo "Go版本为: $(go version)"

echo "==== 下载依赖 ===="
go mod tidy

echo "==== 编译项目 ===="
go build -o gemini_server main_web.go

echo "==== 启动服务 ===="
nohup ./gemini_server start > server.log 2>&1 &
echo "服务已启动，访问 http://你的ip:8080 或你在config.yaml设置的端口"
