#!/bin/bash
# Infinite Refill - Go 版本启动脚本

# 加载配置
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# 显示配置
echo "==================================="
echo "Infinite Refill - Go 版本"
echo "==================================="
echo "Server URL: $SERVER_URL"
echo "Accounts Dir: $ACCOUNTS_DIR"
echo "Target Pool Size: $TARGET_POOL_SIZE"
echo "Total Hold Limit: $TOTAL_HOLD_LIMIT"
echo "==================================="
echo ""

# 执行命令
case "$1" in
    run)
        echo "执行单次续杯..."
        refill run
        ;;
    sync)
        echo "全量探测账号..."
        refill sync
        ;;
    clean)
        if [ "$2" = "--apply" ]; then
            echo "清理失效账号（实际删除）..."
            refill clean --apply
        else
            echo "清理失效账号（预览模式）..."
            refill clean
        fi
        ;;
    scheduler)
        echo "启动定时任务..."
        refill scheduler start
        ;;
    check)
        echo "检查配置..."
        refill check
        ;;
    *)
        echo "用法: $0 {run|sync|clean|scheduler|check}"
        echo ""
        echo "命令说明:"
        echo "  run       - 执行单次续杯"
        echo "  sync      - 全量探测所有账号"
        echo "  clean     - 清理失效账号（预览）"
        echo "  scheduler - 启动定时任务"
        echo "  check     - 检查配置"
        echo ""
        echo "示例:"
        echo "  $0 run"
        echo "  $0 clean --apply"
        exit 1
        ;;
esac
