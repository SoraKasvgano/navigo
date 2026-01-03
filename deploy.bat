@echo off
REM 网址导航系统 - Docker 快速部署脚本 (Windows)
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ========================================
echo   网址导航系统 - Docker 快速部署
echo ========================================
echo.

REM 检查Docker是否安装
echo 检查 Docker 环境...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] 未检测到 Docker，请先安装 Docker Desktop
    pause
    exit /b 1
)

docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] 未检测到 docker-compose，请先安装 Docker Desktop
    pause
    exit /b 1
)

echo [✓] Docker 环境检查通过
echo.

REM 检查是否已有容器在运行
echo 检查容器状态...
docker ps -a --format "{{.Names}}" | findstr /x "navigo" >nul 2>&1
if %errorlevel% equ 0 (
    echo [警告] 发现已存在的容器: navigo
    set /p "confirm=是否删除并重新部署? (Y/N): "
    if /i "!confirm!"=="Y" (
        echo 停止并删除旧容器...
        docker-compose down
        echo [✓] 旧容器已删除
    ) else (
        echo 取消部署
        pause
        exit /b 0
    )
)
echo.

REM 构建并启动容器
echo 开始构建 Docker 镜像...
echo 这可能需要几分钟时间，请耐心等待...
echo.

docker-compose up -d --build

if %errorlevel% equ 0 (
    echo.
    echo ========================================
    echo   部署成功！
    echo ========================================
    echo.
    echo 访问地址：
    echo   • 前台页面: http://localhost:8787/
    echo   • 管理后台: http://localhost:8787/admin
    echo   • 登录页面: http://localhost:8787/login
    echo.
    echo 默认账号: admin / admin
    echo [警告] 请立即登录并修改密码！
    echo.
    echo 常用命令：
    echo   • 查看日志: docker-compose logs -f navigo
    echo   • 重启容器: docker-compose restart
    echo   • 停止容器: docker-compose stop
    echo   • 启动容器: docker-compose start
    echo.

    REM 等待容器完全启动
    echo 等待服务启动...
    timeout /t 5 /nobreak >nul

    REM 查看容器状态
    echo 容器状态:
    docker-compose ps

    echo.
    echo [✓] 部署完成！
) else (
    echo.
    echo ========================================
    echo   部署失败
    echo ========================================
    echo.
    echo 请检查错误信息并尝试以下操作：
    echo   1. 查看详细日志: docker-compose logs
    echo   2. 检查 Docker Desktop 是否正常运行
    echo   3. 检查磁盘空间是否充足
    echo.
)

pause
