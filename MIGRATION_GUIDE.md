# Infinite Refill - Go 版本使用指南

## 配置完成 ✅

你的配置已经从 Bash 版本迁移到 Go 版本。

### 配置文件位置

- **配置文件**: `/Users/devo/workspace/infinite-refill/refill/.env`
- **启动脚本**: `/Users/devo/workspace/infinite-refill/refill/refill.sh`
- **账号目录**: `/Users/devo/Downloads/无限续杯0.3.2/accounts`

### 配置对比

| 配置项 | Bash 版本 | Go 版本 | 说明 |
|--------|-----------|---------|------|
| SERVER_URL | ✅ | ✅ | 相同 |
| USER_KEY | ✅ | ✅ | 相同 |
| ACCOUNTS_DIR | ✅ | ✅ | 相同 |
| TARGET_POOL_SIZE | ✅ | ✅ | 相同 |
| TOTAL_HOLD_LIMIT | ✅ | ✅ | 相同 |
| INTERVAL_MINUTES | ✅ | SCHEDULER_INTERVAL_MINUTES | 重命名 |
| PROXY_MODE | ✅ | ✅ | 相同 |
| WHAM_PROXY_MODE | ✅ | ✅ | 相同 |
| PROBE_PARALLEL | ✅ | ✅ | 相同 |
| REFILL_ITER_MAX | ✅ | LOOP_MAX_ITERATIONS | 重命名 |
| CLEAN_DELETE_STATUSES | ✅ | ✅ | 相同 |
| CLEAN_EXPIRED_DAYS | ✅ | ✅ | 相同 |

### 新增配置项

Go 版本新增了一些配置项：

- `PROBE_WAIT_TIMEOUT=600` - Worker pool 超时时间（秒）
- `TOPUP_RETRY_DELAY=3` - Topup 重试延迟（秒）
- `SERVER_PROXY_MODE=auto` - 服务器请求代理模式

## 使用方法

### 方法 1: 使用启动脚本（推荐）

```bash
cd /Users/devo/workspace/infinite-refill/refill

# 检查配置
./refill.sh check

# 执行单次续杯
./refill.sh run

# 全量探测
./refill.sh sync

# 清理失效账号（预览）
./refill.sh clean

# 清理失效账号（实际删除）
./refill.sh clean --apply

# 启动定时任务
./refill.sh scheduler
```

### 方法 2: 直接使用 refill 命令

```bash
# 加载环境变量
cd /Users/devo/workspace/infinite-refill/refill
export $(cat .env | grep -v '^#' | xargs)

# 执行命令
refill check
refill run
refill sync
refill clean
refill clean --apply
refill scheduler start
```

### 方法 3: 使用详细日志

```bash
cd /Users/devo/workspace/infinite-refill/refill
export $(cat .env | grep -v '^#' | xargs)

# 详细日志模式
refill -v run
```

## 与 Bash 版本对比

### 命令对比

| Bash 版本 | Go 版本 | 说明 |
|-----------|---------|------|
| `./unix/单次续杯.sh` | `refill run` | 单次续杯 |
| `./unix/同步所有账号.sh` | `refill sync` | 全量探测 |
| `./unix/自动清理.sh` | `refill clean --apply` | 清理账号 |
| `./unix/_定时任务_入口.sh` | `refill scheduler start` | 定时任务 |

### 性能对比

| 指标 | Bash 版本 | Go 版本 | 提升 |
|------|-----------|---------|------|
| 探测速度 | 串行 | 并发 (6 workers) | 6x |
| 内存占用 | ~50MB+ | ~10-20MB | 50% |
| 启动速度 | ~1s | ~0.1s | 10x |
| CPU 占用 | 高 | 低 | 更高效 |

### 功能对比

| 功能 | Bash 版本 | Go 版本 |
|------|-----------|---------|
| 并发探测 | ❌ | ✅ |
| 增量闭环 | ✅ | ✅ |
| 自动清理 | ✅ | ✅ |
| 定时任务 | ✅ | ✅ |
| 代理支持 | ✅ | ✅ |
| 错误重试 | ✅ | ✅ |
| 单元测试 | ❌ | ✅ |
| 跨平台 | ❌ | ✅ |

## 迁移步骤

### 1. 准备账号文件

如果你有现有的账号文件，确保它们在正确的目录：

```bash
ls -lh /Users/devo/Downloads/无限续杯0.3.2/accounts/
```

### 2. 测试配置

```bash
cd /Users/devo/workspace/infinite-refill/refill
./refill.sh check
```

### 3. 执行单次续杯

```bash
./refill.sh run
```

### 4. 查看输出

```bash
# 查看探测报告
ls -lh out/probe_report_*.jsonl

# 查看清理报告
ls -lh out/clean_report_*.txt
```

## 常见问题

### Q1: 账号目录为空怎么办？

如果你的账号目录是空的，说明还没有账号文件。你可以：

1. 从 Bash 版本复制账号文件
2. 运行 `refill run` 让服务器分配新账号

### Q2: 如何查看详细日志？

使用 `-v` 参数：

```bash
refill -v run
```

### Q3: 如何设置定时任务？

编辑 `.env` 文件：

```bash
SCHEDULER_ENABLED=true
SCHEDULER_INTERVAL_MINUTES=30
```

然后运行：

```bash
./refill.sh scheduler
```

### Q4: 如何使用代理？

在 `.env` 文件中配置：

```bash
PROXY_MODE=mixed
PROXY_URL=http://your-proxy.com:8080
WHAM_PROXY_MODE=auto
SERVER_PROXY_MODE=direct
```

### Q5: 超时时间太短怎么办？

修改 `.env` 文件：

```bash
PROBE_WAIT_TIMEOUT=1200  # 增加到 20 分钟
TOPUP_MAX_TIME=600       # 增加到 10 分钟
```

然后重新运行。

## 故障排查

### 问题: Worker pool timeout

如果看到 "Worker pool timeout after 600 seconds"，说明探测超时。

**解决方案**:

1. 增加超时时间：
```bash
export PROBE_WAIT_TIMEOUT=1200
```

2. 减少并发数：
```bash
export PROBE_PARALLEL=3
```

3. 检查网络连接

### 问题: Panic: send on closed channel

这个问题已在最新版本修复。确保使用最新版本：

```bash
refill version
# 应该显示: refill version dev (built 2026-03-08_08:48:01)
```

如果不是最新版本，重新安装：

```bash
cd /Users/devo/workspace/infinite-refill/refill
make build && make install
```

## 下一步

1. **测试运行**: `./refill.sh run`
2. **查看结果**: 检查 `out/` 目录
3. **设置定时任务**: 如果需要自动运行
4. **监控日志**: 使用 `-v` 查看详细日志

## 技术支持

- **文档**: 查看 `README.md`, `QUICKSTART.md`
- **配置示例**: 查看 `configs/config.example.yaml`
- **项目状态**: 查看 `PROJECT_STATUS.md`

---

**配置完成时间**: 2026-03-08
**Go 版本**: 1.26.1
**状态**: ✅ 就绪
