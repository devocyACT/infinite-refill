# Infinite Refill - 快速启动指南 (macOS/Unix)

## 📋 对应 Windows 版本的使用流程

按照 Windows 版本的推荐流程，macOS/Unix 版本的操作步骤如下：

---

## 🚀 推荐操作流程

### 步骤 1: 启动交互式菜单

```bash
cd /Users/devo/workspace/infinite-refill/refill
./menu.sh
```

### 步骤 2: 设置配置 (对应 Windows "2设置配置")

在菜单中选择 **`2`** - 设置配置

这将打开 `.env` 文件编辑器，确认以下配置：

```bash
SERVER_URL=https://your-server.com
USER_KEY=your-user-key-here  # 你的密钥
ACCOUNTS_DIR=/Users/devo/Downloads/无限续杯0.3.2/accounts
TARGET_POOL_SIZE=10
TOTAL_HOLD_LIMIT=30
```

保存并退出（nano: Ctrl+X, Y, Enter）

### 步骤 3: 同步所有账号 (对应 Windows "5同步所有账号")

在菜单中选择 **`5`** - 同步所有账号

**原因**: 服务端提前绑定的账号可能已过期，需要先探测状态

这将：
- 探测所有现有账号的状态
- 生成探测报告到 `out/` 目录
- 识别失效账号

### 步骤 4: 立即单次续杯 (对应 Windows "1立即单次续杯")

在菜单中选择 **`1`** - 立即单次续杯

这将：
- 执行完整的续杯流程
- 删除失效账号
- 从服务器获取新账号
- 最多执行 6 轮增量闭环

### 步骤 5: 打开定时任务 (对应 Windows "3打开定时任务")

在菜单中选择 **`3`** - 打开定时任务

这将：
- 在后台启动定时任务
- 每 30 分钟自动执行一次续杯
- 日志保存到 `refill_scheduler.log`

---

## 📱 交互式菜单功能

```
╔══════════════════════════════════════════════════════════════╗
║         Infinite Refill - Go 版本 (macOS/Unix)             ║
╚══════════════════════════════════════════════════════════════╝

当前配置:
  Server URL: https://your-server.com
  User Key: k_cgKNwv...95_I
  Accounts Dir: /Users/devo/Downloads/无限续杯0.3.2/accounts
  Target Pool Size: 10
  Total Hold Limit: 30

请选择操作:

  1. 立即单次续杯
  2. 设置配置 (编辑 .env 文件)
  3. 打开定时任务
  4. 停止定时任务
  5. 同步所有账号 (全量探测)
  6. 自动清理失效账号 (预览)
  7. 自动清理失效账号 (实际删除)
  8. 检查配置和环境
  9. 查看账号统计
  v. 详细日志模式 (verbose)
  0. 退出
```

---

## 🔄 完整操作流程（首次使用）

### 方法 1: 使用交互式菜单（推荐）

```bash
cd /Users/devo/workspace/infinite-refill/refill
./menu.sh

# 然后按照以下顺序操作：
# 1. 选择 2 - 检查配置
# 2. 选择 5 - 同步所有账号
# 3. 选择 1 - 立即单次续杯
# 4. 选择 3 - 打开定时任务
```

### 方法 2: 使用命令行

```bash
cd /Users/devo/workspace/infinite-refill/refill

# 加载配置
export $(cat .env | grep -v '^#' | xargs)

# 1. 检查配置
refill check

# 2. 同步所有账号
refill sync

# 3. 立即单次续杯
refill run

# 4. 打开定时任务
nohup refill scheduler start > refill_scheduler.log 2>&1 &
```

---

## 📊 查看运行状态

### 查看账号统计

在菜单中选择 **`9`** - 查看账号统计

或使用命令：

```bash
ls -lh /Users/devo/Downloads/无限续杯0.3.2/accounts/
```

### 查看探测报告

```bash
ls -lh out/probe_report_*.jsonl
cat out/probe_report_*.jsonl | jq .
```

### 查看定时任务日志

```bash
tail -f refill_scheduler.log
```

### 查看清理报告

```bash
ls -lh out/clean_report_*.txt
cat out/clean_report_*.txt
```

---

## 🛠️ 常用操作

### 停止定时任务

在菜单中选择 **`4`** - 停止定时任务

或使用命令：

```bash
# 查找进程
ps aux | grep "refill scheduler"

# 停止进程
kill <PID>

# 清理锁文件
rm /Users/devo/Downloads/无限续杯0.3.2/accounts/.refill.lock
```

### 清理失效账号

在菜单中选择 **`6`** - 预览模式（不删除）

或选择 **`7`** - 实际删除

### 详细日志模式

在菜单中选择 **`v`** - 详细日志模式

然后选择要执行的操作，将显示详细的调试信息。

---

## 🔍 故障排查

### 问题 1: 账号目录为空

**现象**: 运行后没有账号文件

**解决**:
1. 检查配置中的 `USER_KEY` 是否正确
2. 运行 `refill sync` 同步账号
3. 运行 `refill run` 获取新账号

### 问题 2: 定时任务未运行

**现象**: 定时任务启动后没有执行

**解决**:
1. 查看日志: `tail -f refill_scheduler.log`
2. 检查锁文件: `ls -lh /Users/devo/Downloads/无限续杯0.3.2/accounts/.refill.lock`
3. 如果锁文件存在且过期，删除它

### 问题 3: Worker pool timeout

**现象**: 探测时出现超时

**解决**:
1. 编辑 `.env` 文件
2. 增加超时时间: `PROBE_WAIT_TIMEOUT=1200`
3. 或减少并发数: `PROBE_PARALLEL=3`

---

## 📝 配置说明

### 关键配置项

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `SERVER_URL` | 服务器地址 | https://your-server.com |
| `USER_KEY` | 用户密钥 | 必填 |
| `ACCOUNTS_DIR` | 账号目录 | 必填 |
| `TARGET_POOL_SIZE` | 目标账号数 | 10 |
| `TOTAL_HOLD_LIMIT` | 总持有上限 | 30 |
| `SCHEDULER_INTERVAL_MINUTES` | 定时间隔（分钟） | 30 |
| `PROBE_PARALLEL` | 并发探测数 | 6 |
| `PROBE_WAIT_TIMEOUT` | 探测超时（秒） | 600 |

### 代理配置

```bash
PROXY_MODE=mixed              # auto/direct/proxy/mixed
PROXY_URL=                    # 代理地址（可选）
WHAM_PROXY_MODE=auto          # WHAM 探测代理模式
SERVER_PROXY_MODE=auto        # 服务器请求代理模式
```

---

## 🎯 与 Windows 版本对比

| Windows 版本 | macOS/Unix 版本 | 说明 |
|--------------|-----------------|------|
| 无限续杯.bat | menu.sh | 交互式菜单 |
| 1立即单次续杯 | 菜单选项 1 | 单次续杯 |
| 2设置配置 | 菜单选项 2 | 编辑配置 |
| 3打开定时任务 | 菜单选项 3 | 启动定时任务 |
| 5同步所有账号 | 菜单选项 5 | 全量探测 |
| 无限续杯配置.env | .env | 配置文件 |

---

## 💡 提示

1. **首次使用**: 按照推荐流程操作（2 → 5 → 1 → 3）
2. **日常使用**: 定时任务会自动运行，无需手动操作
3. **查看日志**: 使用详细日志模式（菜单选项 v）排查问题
4. **备份账号**: 账号文件会在清理前自动备份到 `out/清理-*/backup/`

---

**创建时间**: 2026-03-08
**版本**: Go dev
**状态**: ✅ 就绪
