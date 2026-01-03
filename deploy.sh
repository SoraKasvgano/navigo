#!/bin/bash
# 网址导航系统 - Docker 快速部署脚本

set -e

echo "========================================"
echo "  网址导航系统 - Docker 快速部署"
echo "========================================"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置变量
DATA_DIR="/home/docker/navigo"
CONTAINER_NAME="navigo"
HOST_PORT="8787"

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    echo -e "${YELLOW}警告: 建议使用 sudo 运行此脚本${NC}"
    echo "某些操作可能需要root权限"
    echo ""
fi

# 检查Docker是否安装
echo "检查 Docker 环境..."
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: 未检测到 Docker，请先安装 Docker${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}错误: 未检测到 docker-compose，请先安装 docker-compose${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Docker 环境检查通过${NC}"
echo ""

# 创建数据目录
echo "创建数据目录..."
mkdir -p ${DATA_DIR}/data
mkdir -p ${DATA_DIR}/uploads
mkdir -p ${DATA_DIR}/static

# 设置目录权限（容器内使用UID 1000运行）
chown -R 1000:1000 ${DATA_DIR}
echo -e "${GREEN}✓ 数据目录创建成功: ${DATA_DIR}${NC}"
echo ""

# 检查端口是否被占用
echo "检查端口 ${HOST_PORT}..."
if netstat -tuln 2>/dev/null | grep -q ":${HOST_PORT} "; then
    echo -e "${YELLOW}警告: 端口 ${HOST_PORT} 已被占用${NC}"
    echo "请修改 docker-compose.yml 中的端口映射或停止占用该端口的服务"
    read -p "是否继续? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo -e "${GREEN}✓ 端口 ${HOST_PORT} 可用${NC}"
fi
echo ""

# 检查是否已有容器在运行
echo "检查容器状态..."
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${YELLOW}发现已存在的容器: ${CONTAINER_NAME}${NC}"
    read -p "是否删除并重新部署? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "停止并删除旧容器..."
        docker-compose down
        echo -e "${GREEN}✓ 旧容器已删除${NC}"
    else
        echo "取消部署"
        exit 0
    fi
fi
echo ""

# 构建并启动容器
echo "开始构建 Docker 镜像..."
echo "这可能需要几分钟时间，请耐心等待..."
echo ""

if docker-compose up -d --build; then
    echo ""
    echo -e "${GREEN}========================================"
    echo "  部署成功！"
    echo "========================================${NC}"
    echo ""
    echo "访问地址："
    echo "  • 前台页面: http://localhost:${HOST_PORT}/"
    echo "  • 管理后台: http://localhost:${HOST_PORT}/admin"
    echo "  • 登录页面: http://localhost:${HOST_PORT}/login"
    echo ""
    echo "默认账号: admin / admin"
    echo -e "${YELLOW}请立即登录并修改密码！${NC}"
    echo ""
    echo "数据目录: ${DATA_DIR}"
    echo ""
    echo "常用命令："
    echo "  • 查看日志: docker-compose logs -f navigo"
    echo "  • 重启容器: docker-compose restart"
    echo "  • 停止容器: docker-compose stop"
    echo "  • 启动容器: docker-compose start"
    echo ""

    # 等待容器完全启动
    echo "等待服务启动..."
    sleep 5

    # 查看容器状态
    echo "容器状态:"
    docker-compose ps

    echo ""
    echo -e "${GREEN}部署完成！${NC}"
else
    echo ""
    echo -e "${RED}========================================"
    echo "  部署失败"
    echo "========================================${NC}"
    echo ""
    echo "请检查错误信息并尝试以下操作："
    echo "  1. 查看详细日志: docker-compose logs"
    echo "  2. 检查 Docker 是否正常运行"
    echo "  3. 检查磁盘空间是否充足"
    echo ""
    exit 1
fi
