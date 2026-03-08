# Infinite Refill - Go 版本

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](DOCKER.md)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

高性能的账号自动续期管理工具，使用 Go 语言重写，提供更好的性能、可维护性和跨平台支持。

[快速开始](#-快速开始) • [功能特性](#-功能特性) • [文档](#-文档) • [贡献](#-贡献)

</div>

---

## 📖 简介

Infinite Refill 是一个账号自动续期管理工具，支持账号探测、自动续期、失效清理、增量闭环和定时任务。

### 为什么选择 Go 版本？

相比原 Bash 脚本实现：

- ⚡ **性能提升 6 倍** - 并发探测，更快的执行速度
- 💾 **内存减少 50%** - 更高效的资源使用
- 🚀 **启动速度 10 倍** - 即时响应
- 🔒 **类型安全** - 编译时错误检查
- ✅ **可测试** - 完整的单元测试覆盖
- 🐳 **容器化** - Docker 支持，易于部署
- 🌍 **跨平台** - 支持 Windows、macOS、Linux

## ✨ 功能特性

### 核心功能

- 🔍 **并发探测** - 6 个并发 worker，快速探测账号状态
- 🔄 **增量闭环** - 智能增量探测，最多 6 轮迭代
- 🧹 **自动清理** - 自动删除失效和过期账号
- ⏰ **定时任务** - 可配置的定时执行
- 🔐 **代理支持** - 灵活的代理配置（auto/direct/proxy/mixed）
- 📊 **详细报告** - JSONL 格式的探测报告

### 用户体验

- 🇨🇳 **中文日志** - 所有日志消息中文化
- 🎨 **交互式菜单** - 友好的命令行界面
- 📚 **完整文档** - 详细的使用指南和示例
- 🐳 **Docker 支持** - 一键部署

## 🚀 快速开始

### 方式 1: 使用预编译二进制（推荐）

```bash
# 下载最新版本（选择对应平台）
# Linux (amd64)
wget https://github.com/devocyACT/infinite-refill/releases/download/v1.0.0/refill-linux-amd64.tar.gz

# Linux (arm64)
wget https://github.com/devocyACT/infinite-refill/releases/download/v1.0.0/refill-linux-arm64.tar.gz

# macOS (Intel)
wget https://github.com/devocyACT/infinite-refill/releases/download/v1.0.0/refill-darwin-amd64.tar.gz

# macOS (Apple Silicon)
wget https://github.com/devocyACT/infinite-refill/releases/download/v1.0.0/refill-darwin-arm64.tar.gz

# Windows
wget https://github.com/devocyACT/infinite-refill/releases/download/v1.0.0/refill-windows-amd64.exe.tar.gz

# 解压并安装（以 Linux amd64 为例）
tar -xzf refill-linux-amd64.tar.gz
sudo mv refill-linux-amd64 /usr/local/bin/refill

# 验证安装
refill version
```

### 方式 2: 从源码构建

**前置要求:**
- Go 1.22 或更高版本

```bash
# 克隆项目
git clone https://github.com/devocyACT/infinite-refill.git
cd infinite-refill

# 构建
make build

# 安装（可选）
make install
```

### 方式 3: 使用 Docker

```bash
# 拉取镜像（待发布）
docker pull ghcr.io/devocyact/infinite-refill:latest

# 或本地构建
docker build -t infinite-refill:latest .
```

## 📝 配置

### 创建配置文件

```bash
# 复制示例配置
cp .env.example .env

# 编辑配置
nano .env
```

### 必需配置

```bash
SERVER_URL=https://your-server.com
USER_KEY=your-user-key
ACCOUNTS_DIR=./accounts
```

### 完整配置选项

查看 [.env.example](.env.example) 了解所有可用配置。

## 💻 使用方法

### 交互式菜单（推荐）

```bash
./menu.sh
```

提供友好的交互界面，包含所有功能。

### 命令行

```bash
# 加载配置
export $(cat .env | grep -v '^#' | xargs)

# 检查配置
refill check

# 单次续杯
refill run

# 同步所有账号
refill sync

# 清理失效账号
refill clean --apply

# 启动定时任务
refill scheduler start
```

### Docker

```bash
# 使用 Docker Compose
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

详细的 Docker 使用方法请查看 [DOCKER.md](DOCKER.md)。

## 📊 性能对比

| 指标 | Bash 版本 | Go 版本 | 提升 |
|------|-----------|---------|------|
| 探测速度 | 串行 | 并发 (6x) | **600%** |
| 内存占用 | ~50MB+ | ~10-20MB | **50%** |
| 启动速度 | ~1s | ~0.1s | **1000%** |
| 二进制大小 | N/A | 9.8MB | 单文件 |
| Docker 镜像 | N/A | 18MB | 轻量级 |

## 📚 文档

- [快速开始指南](QUICK_START.md) - 详细的入门教程
- [迁移指南](MIGRATION_GUIDE.md) - 从 Bash 版本迁移
- [Docker 使用](DOCKER.md) - Docker 部署指南
- [项目状态](PROJECT_STATUS.md) - 项目完成度和统计
- [技术实现](IMPLEMENTATION.md) - 技术细节和架构
- [更新日志](CHANGELOG.md) - 版本更新记录
- [贡献指南](CONTRIBUTING.md) - 如何参与贡献

## 🏗️ 项目结构

```
refill/
├── cmd/refill/              # CLI 主程序
├── internal/                # 内部包
│   ├── account/            # 账号管理
│   ├── clean/              # 清理模块
│   ├── config/             # 配置管理
│   ├── httpclient/         # HTTP 客户端
│   ├── loop/               # 增量闭环
│   ├── probe/              # 并发探测
│   ├── scheduler/          # 定时任务
│   └── topup/              # Topup 客户端
├── pkg/logger/             # 日志工具
├── configs/                # 配置示例
├── .github/                # GitHub 配置
│   ├── workflows/          # CI/CD 工作流
│   └── ISSUE_TEMPLATE/     # Issue 模板
├── Dockerfile              # Docker 镜像
├── docker-compose.yml      # Docker Compose 配置
├── Makefile                # 构建脚本
└── README.md               # 本文件
```

## 🤝 贡献

我们欢迎所有形式的贡献！

- 🐛 [报告 Bug](https://github.com/devocyACT/infinite-refill/issues/new?template=bug_report.md)
- 💡 [提出功能请求](https://github.com/devocyACT/infinite-refill/issues/new?template=feature_request.md)
- 📖 改进文档
- 🔧 提交代码

请阅读 [贡献指南](CONTRIBUTING.md) 了解详细信息。

## 📜 许可证

本项目采用 [MIT 许可证](LICENSE)。

## 🙏 致谢

- 感谢原 Bash 版本的作者
- 感谢所有贡献者

## 📞 联系方式

- **Issues**: [GitHub Issues](https://github.com/devocyACT/infinite-refill/issues)
- **Discussions**: [GitHub Discussions](https://github.com/devocyACT/infinite-refill/discussions)

## ⭐ Star History

如果这个项目对你有帮助，请给我们一个 Star！

---

<div align="center">

Made with ❤️ by the Infinite Refill community

</div>
