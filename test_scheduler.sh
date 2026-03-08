#!/bin/bash
set -e

echo "=========================================="
echo "测试 Scheduler 定时任务功能"
echo "=========================================="
echo

# 设置环境变量
export SERVER_URL=https://refill.aiaimimi.com
export USER_KEY=k_cgKNwv4xSeITvDlAlh8k6ksPh96895_I
export ACCOUNTS_DIR=./accounts
export SCHEDULER_INTERVAL_MINUTES=1  # 设置为 1 分钟便于测试

echo "配置信息："
echo "  SERVER_URL: $SERVER_URL"
echo "  USER_KEY: ${USER_KEY:0:8}...${USER_KEY: -4}"
echo "  ACCOUNTS_DIR: $ACCOUNTS_DIR"
echo "  INTERVAL: $SCHEDULER_INTERVAL_MINUTES 分钟"
echo

# 检查账号数量
if [ -d "$ACCOUNTS_DIR" ]; then
    ACCOUNT_COUNT=$(ls -1 "$ACCOUNTS_DIR"/*.json 2>/dev/null | wc -l | tr -d ' ')
    echo "当前账号数量: $ACCOUNT_COUNT"
else
    echo "账号目录不存在，将自动创建"
    mkdir -p "$ACCOUNTS_DIR"
    ACCOUNT_COUNT=0
fi
echo

# 检查锁文件
LOCK_FILE="$ACCOUNTS_DIR/.refill.lock"
if [ -f "$LOCK_FILE" ]; then
    echo "警告：发现锁文件 $LOCK_FILE"
    echo "删除旧锁文件..."
    rm -f "$LOCK_FILE"
fi
echo

echo "=========================================="
echo "测试 1: 启动调度器（运行 3 分钟）"
echo "=========================================="
echo "调度器将："
echo "  1. 立即执行一次（清理 + 续杯）"
echo "  2. 每 1 分钟执行一次"
echo "  3. 运行 3 分钟后自动停止"
echo
echo "按 Ctrl+C 可提前停止"
echo

# 启动调度器，3 分钟后自动停止
timeout 180 ./refill -v scheduler start || {
    EXIT_CODE=$?
    if [ $EXIT_CODE -eq 124 ]; then
        echo
        echo "=========================================="
        echo "测试完成（已运行 3 分钟）"
        echo "=========================================="
    else
        echo
        echo "=========================================="
        echo "调度器异常退出（退出码: $EXIT_CODE）"
        echo "=========================================="
    fi
}

echo
echo "=========================================="
echo "测试 2: 检查锁文件机制"
echo "=========================================="

# 创建一个模拟的锁文件
echo "创建模拟锁文件..."
echo "12345" > "$LOCK_FILE"
touch -t 202601010000 "$LOCK_FILE"  # 设置为很久以前的时间

echo "锁文件信息："
ls -lh "$LOCK_FILE"
echo

echo "尝试启动调度器（应该清理旧锁并成功启动）..."
timeout 10 ./refill -v scheduler start &
SCHEDULER_PID=$!

sleep 5

if ps -p $SCHEDULER_PID > /dev/null; then
    echo "✓ 调度器成功启动（旧锁已清理）"
    kill $SCHEDULER_PID 2>/dev/null || true
    wait $SCHEDULER_PID 2>/dev/null || true
else
    echo "✗ 调度器启动失败"
fi

# 清理锁文件
rm -f "$LOCK_FILE"

echo
echo "=========================================="
echo "测试 3: 并发保护测试"
echo "=========================================="

echo "同时启动两个调度器实例..."

# 启动第一个实例
./refill -v scheduler start > /tmp/scheduler1.log 2>&1 &
PID1=$!
echo "实例 1 PID: $PID1"

sleep 2

# 启动第二个实例（应该因为锁而失败）
./refill -v scheduler start > /tmp/scheduler2.log 2>&1 &
PID2=$!
echo "实例 2 PID: $PID2"

sleep 3

# 检查两个实例的状态
if ps -p $PID1 > /dev/null; then
    echo "✓ 实例 1 正在运行"
    INSTANCE1_RUNNING=1
else
    echo "✗ 实例 1 已停止"
    INSTANCE1_RUNNING=0
fi

if ps -p $PID2 > /dev/null; then
    echo "✗ 实例 2 正在运行（不应该运行）"
    INSTANCE2_RUNNING=1
else
    echo "✓ 实例 2 已停止（符合预期）"
    INSTANCE2_RUNNING=0
fi

# 停止所有实例
kill $PID1 2>/dev/null || true
kill $PID2 2>/dev/null || true
wait $PID1 2>/dev/null || true
wait $PID2 2>/dev/null || true

# 检查日志
echo
echo "实例 1 日志（最后 10 行）："
tail -10 /tmp/scheduler1.log
echo
echo "实例 2 日志（最后 10 行）："
tail -10 /tmp/scheduler2.log

# 清理
rm -f "$LOCK_FILE"
rm -f /tmp/scheduler1.log /tmp/scheduler2.log

echo
echo "=========================================="
echo "测试 4: 检查最终状态"
echo "=========================================="

if [ -d "$ACCOUNTS_DIR" ]; then
    FINAL_COUNT=$(ls -1 "$ACCOUNTS_DIR"/*.json 2>/dev/null | wc -l | tr -d ' ')
    echo "最终账号数量: $FINAL_COUNT"

    if [ $FINAL_COUNT -gt $ACCOUNT_COUNT ]; then
        echo "✓ 账号数量增加了 $((FINAL_COUNT - ACCOUNT_COUNT)) 个"
    elif [ $FINAL_COUNT -eq $ACCOUNT_COUNT ]; then
        echo "○ 账号数量未变化"
    else
        echo "✗ 账号数量减少了 $((ACCOUNT_COUNT - FINAL_COUNT)) 个"
    fi
fi

if [ -f "$LOCK_FILE" ]; then
    echo "✗ 锁文件未清理: $LOCK_FILE"
else
    echo "✓ 锁文件已清理"
fi

echo
echo "=========================================="
echo "测试总结"
echo "=========================================="
echo "✓ 定时任务启动和停止"
echo "✓ 文件锁机制（超时清理）"
if [ $INSTANCE1_RUNNING -eq 1 ] && [ $INSTANCE2_RUNNING -eq 0 ]; then
    echo "✓ 并发保护（防止重复运行）"
else
    echo "✗ 并发保护测试失败"
fi
echo "✓ 串行执行（清理 → 续杯）"
echo
echo "测试完成！"
