# Gitea 编译构建指南

本项目提供了两个便捷的构建脚本，帮助您快速编译和发布 Gitea。

## 脚本说明

### 1. `build-quick.sh` - 快速构建脚本

用于日常开发，快速编译前端和后端，生成可运行的 `gitea` 二进制文件。

**使用方法：**

```bash
./build-quick.sh
```

**特点：**
- 🚀 快速构建，适合开发调试
- 自动安装前端依赖
- 编译前端资源（webpack）
- 编译后端二进制文件
- 构建时间短

**输出：**
- `./gitea` - 可执行的二进制文件

### 2. `build-release.sh` - 完整发布脚本

用于生产环境发布，包含完整的依赖检查、清理、编译和打包流程。

**使用方法：**

```bash
./build-release.sh
```

**特点：**
- ✅ 完整的依赖检查（Go、Node.js、pnpm、make）
- 🧹 可选清理旧构建产物
- 📦 自动打包发布包
- 🔐 生成 SHA256 校验和
- 📝 包含启动脚本和必要文件
- 📊 显示详细构建信息和耗时

**输出：**
- `./gitea` - 可执行的二进制文件
- `dist/gitea-{version}-{platform}-{arch}/` - 发布包目录
- `dist/gitea-{version}-{platform}-{arch}.tar.gz` - 压缩的发布包
- `dist/gitea-{version}-{platform}-{arch}.tar.gz.sha256` - 校验和文件

## 系统要求

在运行构建脚本前，请确保系统已安装以下依赖：

- **Go** >= 1.25.x
- **Node.js** >= 18.x
- **pnpm** (会自动安装如果缺失)
- **make** (用于原生 Makefile 构建)
- **git** (用于版本管理)

## 快速开始

### 开发构建

```bash
# 1. 克隆或进入项目目录
cd /path/to/gitea

# 2. 快速构建
./build-quick.sh

# 3. 运行 Gitea
./gitea web
```

### 生产发布

```bash
# 1. 完整构建和打包
./build-release.sh

# 2. 解压发布包
cd dist
tar -xzf gitea-1.25.2-kysion-darwin-arm64.tar.gz

# 3. 进入目录并启动
cd gitea-1.25.2-kysion-darwin-arm64
./start.sh
```

## 构建配置

### 版本号

版本号从 `VERSION` 文件读取，当前版本：`1.25.2-kysion`

### 构建标签

默认使用以下构建标签：
- `bindata` - 内嵌静态资源
- `sqlite` - 支持 SQLite 数据库
- `sqlite_unlock_notify` - SQLite 解锁通知

### 环境变量

可以通过环境变量自定义构建：

```bash
# 启用 CGO（默认禁用）
CGO_ENABLED=1 ./build-quick.sh

# 添加额外的构建标签
TAGS="bindata sqlite pam" ./build-quick.sh

# 自定义版本号
echo "1.25.3-custom" > VERSION
./build-quick.sh
```

## 使用原生 Makefile

如果需要更多控制，也可以直接使用 Makefile：

```bash
# 编译所有（前端 + 后端）
make build

# 只编译前端
make frontend

# 只编译后端
make backend

# 完整发布构建（多平台）
make release

# 清理构建产物
make clean

# 清理所有（包括 node_modules）
make clean-all
```

## 故障排除

### 1. pnpm 未安装

```bash
npm install -g pnpm
```

### 2. Go 版本过低

请升级到 Go 1.25.x 或更高版本：
```bash
# macOS
brew upgrade go

# Linux
# 从官网下载: https://go.dev/dl/
```

### 3. 前端构建失败

```bash
# 删除 node_modules 重新安装
rm -rf node_modules pnpm-lock.yaml
pnpm install
```

### 4. 后端编译失败

```bash
# 清理 Go 模块缓存
go clean -modcache
go mod download
```

## 目录结构

```
gitea/
├── build-quick.sh          # 快速构建脚本
├── build-release.sh        # 完整发布脚本
├── BUILD_GUIDE.md          # 本文档
├── VERSION                 # 版本文件
├── Makefile                # 原生构建文件
├── go.mod                  # Go 模块定义
├── package.json            # 前端依赖
├── webpack.config.ts       # Webpack 配置
├── gitea                   # 编译后的二进制文件
└── dist/                   # 发布包目录
    └── gitea-{version}-{platform}-{arch}/
        ├── gitea           # 二进制文件
        ├── start.sh        # 启动脚本
        ├── custom/         # 自定义配置
        ├── options/        # 选项文件
        ├── public/         # 静态资源
        ├── templates/      # 模板文件
        ├── LICENSE         # 许可证
        └── README.md       # 说明文档
```

## 性能建议

### 加速构建

1. **使用快速构建脚本**：开发时使用 `build-quick.sh`
2. **跳过清理**：在 `build-release.sh` 中选择不清理
3. **并行编译**：
   ```bash
   GOMAXPROCS=$(nproc) ./build-quick.sh
   ```

### 减小文件大小

构建脚本已默认使用 `-ldflags "-s -w"` 来减小二进制文件大小：
- `-s`：省略符号表
- `-w`：省略 DWARF 调试信息

## 部署建议

### 使用发布包部署

1. 上传 `dist/gitea-*.tar.gz` 到服务器
2. 解压到目标目录
3. 配置 `custom/conf/app.ini`
4. 运行 `./start.sh`

### 使用 systemd 服务

创建 `/etc/systemd/system/gitea.service`：

```ini
[Unit]
Description=Gitea
After=network.target

[Service]
Type=simple
User=git
Group=git
WorkingDirectory=/opt/gitea
ExecStart=/opt/gitea/gitea web
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
systemctl daemon-reload
systemctl enable gitea
systemctl start gitea
```

## 许可证

本项目遵循 Gitea 的原始许可证。详见 `LICENSE` 文件。

## 支持

如有问题，请查看：
- Gitea 官方文档: https://docs.gitea.io/
- 项目仓库: https://github.com/go-gitea/gitea
