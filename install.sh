#!/usr/bin/env bash
# 自动安装 Docker 和 nc（netcat）
# 支持 macOS、Ubuntu/Debian、CentOS/RHEL
# 幂等设计，已安装则跳过

set -e

# 检查是否为 root 或 sudo
if [ "$EUID" -ne 0 ]; then
  if command -v sudo >/dev/null 2>&1; then
    SUDO="sudo"
  else
    echo "请以 root 权限或安装 sudo 后重试。"
    exit 1
  fi
else
  SUDO=""
fi

# 检测操作系统
OS="$(uname -s)"

install_docker_mac() {
  if ! command -v docker >/dev/null 2>&1; then
    if command -v brew >/dev/null 2>&1; then
      echo "[macOS] 安装 Docker Desktop..."
      brew install --cask docker
      echo "请手动启动 Docker Desktop 应用。"
    else
      echo "未检测到 Homebrew，请先安装 Homebrew。"
      exit 1
    fi
  else
    echo "[macOS] Docker 已安装。"
  fi
}

install_docker_ubuntu() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "[Ubuntu/Debian] 安装 Docker..."
    $SUDO apt-get update
    $SUDO apt-get install -y \
      ca-certificates \
      curl \
      gnupg \
      lsb-release
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | $SUDO gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    echo \ 
      "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
      $(lsb_release -cs) stable" | $SUDO tee /etc/apt/sources.list.d/docker.list > /dev/null
    $SUDO apt-get update
    $SUDO apt-get install -y docker-ce docker-ce-cli containerd.io
    $SUDO systemctl enable docker
    $SUDO systemctl start docker
  else
    echo "[Ubuntu/Debian] Docker 已安装。"
  fi
}

install_docker_centos() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "[CentOS/RHEL] 安装 Docker..."
    $SUDO yum install -y yum-utils
    $SUDO yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
    $SUDO yum install -y docker-ce docker-ce-cli containerd.io
    $SUDO systemctl enable docker
    $SUDO systemctl start docker
  else
    echo "[CentOS/RHEL] Docker 已安装。"
  fi
}

install_nc_mac() {
  if ! command -v nc >/dev/null 2>&1; then
    if command -v brew >/dev/null 2>&1; then
      echo "[macOS] 安装 netcat..."
      brew install netcat
    else
      echo "未检测到 Homebrew，请先安装 Homebrew。"
      exit 1
    fi
  else
    echo "[macOS] nc 已安装。"
  fi
}

install_nc_ubuntu() {
  if ! command -v nc >/dev/null 2>&1; then
    echo "[Ubuntu/Debian] 安装 netcat..."
    $SUDO apt-get update
    $SUDO apt-get install -y netcat-openbsd
  else
    echo "[Ubuntu/Debian] nc 已安装。"
  fi
}

install_nc_centos() {
  if ! command -v nc >/dev/null 2>&1; then
    echo "[CentOS/RHEL] 安装 netcat..."
    $SUDO yum install -y nmap-ncat || $SUDO yum install -y nc
  else
    echo "[CentOS/RHEL] nc 已安装。"
  fi
}

case "$OS" in
  Darwin)
    install_docker_mac
    install_nc_mac
    ;;
  Linux)
    if [ -f /etc/os-release ]; then
      . /etc/os-release
      case "$ID" in
        ubuntu|debian)
          install_docker_ubuntu
          install_nc_ubuntu
          ;;
        centos|rhel|rocky|almalinux)
          install_docker_centos
          install_nc_centos
          ;;
        *)
          echo "暂不支持的 Linux 发行版: $ID"
          exit 1
          ;;
      esac
    else
      echo "无法识别的 Linux 系统。"
      exit 1
    fi
    ;;
  *)
    echo "暂不支持的操作系统: $OS"
    exit 1
    ;;
esac

echo "\n[完成] Docker 和 nc 安装流程已执行。" 

docker-compose up -d --build