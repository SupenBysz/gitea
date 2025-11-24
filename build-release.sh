#!/bin/bash

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 打印标题
print_banner() {
    echo -e "${GREEN}"
    echo "╔════════════════════════════════════════════════════════╗"
    echo "║          Gitea 一键编译发布脚本                        ║"
    echo "║          Build and Release Script                      ║"
    echo "╚════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 未安装，请先安装 $1"
        return 1
    fi
    log_success "$1 已安装"
    return 0
}

# 检查依赖
check_dependencies() {
    log_info "检查系统依赖..."

    local all_ok=true

    if ! check_command "go"; then
        all_ok=false
    else
        log_info "Go 版本: $(go version)"
    fi

    if ! check_command "node"; then
        all_ok=false
    else
        log_info "Node 版本: $(node --version)"
    fi

    if ! check_command "pnpm"; then
        log_warning "pnpm 未安装，尝试使用 npm 安装..."
        npm install -g pnpm
    else
        log_info "pnpm 版本: $(pnpm --version)"
    fi

    if ! check_command "make"; then
        all_ok=false
    fi

    if [ "$all_ok" = false ]; then
        log_error "依赖检查失败，请安装缺失的依赖"
        exit 1
    fi

    log_success "所有依赖检查通过"
}

# 读取版本信息
get_version() {
    if [ -f "VERSION" ]; then
        VERSION=$(cat VERSION | tr -d '\n' | tr -d ' ')
        log_info "当前版本: ${VERSION}"
    else
        VERSION="dev"
        log_warning "未找到 VERSION 文件，使用默认版本: ${VERSION}"
    fi
}

# 清理旧的构建产物
clean_build() {
    log_info "清理旧的构建产物..."

    # 清理后端产物
    if [ -f "gitea" ]; then
        rm -f gitea
        log_info "删除旧的 gitea 二进制文件"
    fi

    # 清理前端产物
    if [ -d "public/assets/js" ] || [ -d "public/assets/css" ]; then
        rm -rf public/assets/js public/assets/css public/assets/fonts
        log_info "删除旧的前端构建文件"
    fi

    # 清理发布目录
    if [ -d "dist" ]; then
        rm -rf dist
        log_info "删除旧的发布目录"
    fi

    log_success "清理完成"
}

# 安装依赖
install_dependencies() {
    log_info "安装项目依赖..."

    # 安装前端依赖
    log_info "安装前端依赖..."
    pnpm install --frozen-lockfile

    # 下载 Go 模块
    log_info "下载 Go 模块..."
    go mod download

    log_success "依赖安装完成"
}

# 编译前端
build_frontend() {
    log_info "编译前端资源..."

    export BROWSERSLIST_IGNORE_OLD_DATA=true
    pnpm exec webpack --disable-interpret

    log_success "前端编译完成"
}

# 编译后端
build_backend() {
    log_info "编译后端程序..."

    # 设置构建参数
    export CGO_ENABLED=0
    export GOEXPERIMENT=jsonv2

    # 获取构建标签
    TAGS="bindata sqlite sqlite_unlock_notify"

    # 构建
    go build -v -tags "${TAGS}" -ldflags "-s -w -X main.Version=${VERSION}" -o gitea

    log_success "后端编译完成"
}

# 创建发布包
create_release_package() {
    log_info "创建发布包..."

    # 创建发布目录
    RELEASE_DIR="dist/gitea-${VERSION}-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)"
    mkdir -p "${RELEASE_DIR}"

    # 复制二进制文件
    cp gitea "${RELEASE_DIR}/"

    # 复制必要的文件
    if [ -d "custom" ]; then
        cp -r custom "${RELEASE_DIR}/"
    fi

    if [ -d "options" ]; then
        cp -r options "${RELEASE_DIR}/"
    fi

    if [ -d "public" ]; then
        cp -r public "${RELEASE_DIR}/"
    fi

    if [ -d "templates" ]; then
        cp -r templates "${RELEASE_DIR}/"
    fi

    if [ -f "LICENSE" ]; then
        cp LICENSE "${RELEASE_DIR}/"
    fi

    if [ -f "README.md" ]; then
        cp README.md "${RELEASE_DIR}/"
    fi

    # 创建启动脚本
    cat > "${RELEASE_DIR}/start.sh" << 'EOF'
#!/bin/bash
# Gitea 启动脚本

GITEA_WORK_DIR="$(cd $(dirname $0) && pwd)"
export GITEA_WORK_DIR

cd "${GITEA_WORK_DIR}"

# 检查是否首次运行
if [ ! -f "custom/conf/app.ini" ]; then
    echo "检测到首次运行，请先配置 custom/conf/app.ini"
    echo "或者访问 http://localhost:3000 进行初始化配置"
fi

# 启动 Gitea
./gitea web
EOF

    chmod +x "${RELEASE_DIR}/start.sh"
    chmod +x "${RELEASE_DIR}/gitea"

    # 打包
    log_info "压缩发布包..."
    cd dist
    tar -czf "gitea-${VERSION}-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m).tar.gz" \
        "$(basename ${RELEASE_DIR})"
    cd ..

    # 生成校验和
    cd dist
    shasum -a 256 "gitea-${VERSION}-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m).tar.gz" > \
        "gitea-${VERSION}-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m).tar.gz.sha256"
    cd ..

    log_success "发布包创建完成: ${RELEASE_DIR}"
    log_success "压缩包: dist/gitea-${VERSION}-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m).tar.gz"
}

# 显示构建信息
show_build_info() {
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                 构建完成                               ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}版本信息:${NC} ${VERSION}"
    echo -e "${BLUE}二进制文件:${NC} $(pwd)/gitea"
    if [ -f "gitea" ]; then
        echo -e "${BLUE}文件大小:${NC} $(du -h gitea | cut -f1)"
    fi
    echo -e "${BLUE}发布包目录:${NC} $(pwd)/dist/"
    echo ""
    echo -e "${YELLOW}快速启动命令:${NC}"
    echo -e "  ./gitea web"
    echo ""
    echo -e "${YELLOW}或使用发布包:${NC}"
    echo -e "  cd ${RELEASE_DIR}"
    echo -e "  ./start.sh"
    echo ""
}

# 主函数
main() {
    local start_time=$(date +%s)

    print_banner

    # 检查是否在项目根目录
    if [ ! -f "go.mod" ] || [ ! -f "Makefile" ]; then
        log_error "请在 Gitea 项目根目录下运行此脚本"
        exit 1
    fi

    # 执行构建流程
    check_dependencies
    get_version

    read -p "是否清理旧的构建产物? (y/n) [默认: y]: " clean_choice
    clean_choice=${clean_choice:-y}
    if [ "$clean_choice" = "y" ] || [ "$clean_choice" = "Y" ]; then
        clean_build
    fi

    install_dependencies
    build_frontend
    build_backend
    create_release_package

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    show_build_info

    log_success "总耗时: ${duration} 秒"
}

# 运行主函数
main "$@"
