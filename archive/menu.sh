#!/bin/bash
# Infinite Refill - Go 版本交互式菜单

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 加载配置
load_config() {
    if [ -f .env ]; then
        export $(cat .env | grep -v '^#' | xargs)
    else
        echo -e "${RED}错误: 找不到 .env 配置文件${NC}"
        exit 1
    fi
}

# 显示菜单
show_menu() {
    clear
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║         Infinite Refill - Go 版本 (macOS/Unix)             ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${GREEN}当前配置:${NC}"
    echo -e "  Server URL: ${YELLOW}$SERVER_URL${NC}"
    echo -e "  User Key: ${YELLOW}${USER_KEY:0:8}...${USER_KEY: -4}${NC}"
    echo -e "  Accounts Dir: ${YELLOW}$ACCOUNTS_DIR${NC}"
    echo -e "  Target Pool Size: ${YELLOW}$TARGET_POOL_SIZE${NC}"
    echo -e "  Total Hold Limit: ${YELLOW}$TOTAL_HOLD_LIMIT${NC}"
    echo ""
    echo -e "${GREEN}请选择操作:${NC}"
    echo ""
    echo -e "  ${BLUE}1${NC}. 立即单次续杯"
    echo -e "  ${BLUE}2${NC}. 设置配置 (编辑 .env 文件)"
    echo -e "  ${BLUE}3${NC}. 打开定时任务"
    echo -e "  ${BLUE}4${NC}. 停止定时任务"
    echo -e "  ${BLUE}5${NC}. 同步所有账号 (全量探测)"
    echo -e "  ${BLUE}6${NC}. 自动清理失效账号 (预览)"
    echo -e "  ${BLUE}7${NC}. 自动清理失效账号 (实际删除)"
    echo -e "  ${BLUE}8${NC}. 检查配置和环境"
    echo -e "  ${BLUE}9${NC}. 查看账号统计"
    echo -e "  ${BLUE}v${NC}. 详细日志模式 (verbose)"
    echo -e "  ${BLUE}0${NC}. 退出"
    echo ""
    echo -n -e "${GREEN}请输入选项 [0-9/v]: ${NC}"
}

# 等待用户按键
wait_key() {
    echo ""
    echo -n -e "${YELLOW}按任意键继续...${NC}"
    read -n 1 -s
}

# 1. 立即单次续杯
run_refill() {
    echo -e "${GREEN}=== 立即单次续杯 ===${NC}"
    echo ""
    refill run
    wait_key
}

# 2. 设置配置
edit_config() {
    echo -e "${GREEN}=== 设置配置 ===${NC}"
    echo ""
    echo "正在打开配置文件..."

    # 检测可用的编辑器
    if command -v nano &> /dev/null; then
        nano .env
    elif command -v vim &> /dev/null; then
        vim .env
    elif command -v vi &> /dev/null; then
        vi .env
    else
        echo -e "${RED}错误: 找不到文本编辑器 (nano/vim/vi)${NC}"
        echo "请手动编辑文件: .env"
        wait_key
        return
    fi

    echo ""
    echo -e "${GREEN}配置已更新，重新加载...${NC}"
    load_config
    wait_key
}

# 3. 打开定时任务
start_scheduler() {
    echo -e "${GREEN}=== 打开定时任务 ===${NC}"
    echo ""
    echo -e "${YELLOW}定时任务将在后台运行，间隔: ${SCHEDULER_INTERVAL_MINUTES} 分钟${NC}"
    echo ""
    echo "启动中..."

    # 检查是否已经在运行
    if [ -f "$ACCOUNTS_DIR/.refill.lock" ]; then
        echo -e "${YELLOW}警告: 检测到锁文件，可能已有实例在运行${NC}"
        echo -n "是否继续? [y/N]: "
        read -n 1 confirm
        echo ""
        if [[ ! $confirm =~ ^[Yy]$ ]]; then
            return
        fi
    fi

    # 后台运行
    nohup refill scheduler start > refill_scheduler.log 2>&1 &
    SCHEDULER_PID=$!
    echo $SCHEDULER_PID > .scheduler.pid

    echo -e "${GREEN}✓ 定时任务已启动 (PID: $SCHEDULER_PID)${NC}"
    echo "日志文件: refill_scheduler.log"
    echo ""
    echo "查看日志: tail -f refill_scheduler.log"
    echo "停止任务: 选择菜单选项 4"
    wait_key
}

# 4. 停止定时任务
stop_scheduler() {
    echo -e "${GREEN}=== 停止定时任务 ===${NC}"
    echo ""

    if [ -f .scheduler.pid ]; then
        SCHEDULER_PID=$(cat .scheduler.pid)
        if ps -p $SCHEDULER_PID > /dev/null 2>&1; then
            kill $SCHEDULER_PID
            echo -e "${GREEN}✓ 定时任务已停止 (PID: $SCHEDULER_PID)${NC}"
            rm .scheduler.pid
        else
            echo -e "${YELLOW}定时任务未运行 (PID: $SCHEDULER_PID 不存在)${NC}"
            rm .scheduler.pid
        fi
    else
        echo -e "${YELLOW}未找到运行中的定时任务${NC}"
    fi

    # 清理锁文件
    if [ -f "$ACCOUNTS_DIR/.refill.lock" ]; then
        rm "$ACCOUNTS_DIR/.refill.lock"
        echo -e "${GREEN}✓ 已清理锁文件${NC}"
    fi

    wait_key
}

# 5. 同步所有账号
sync_accounts() {
    echo -e "${GREEN}=== 同步所有账号 (全量探测) ===${NC}"
    echo ""
    echo -e "${YELLOW}这将探测所有账号的状态，可能需要几分钟...${NC}"
    echo ""
    refill sync
    wait_key
}

# 6. 自动清理 (预览)
clean_preview() {
    echo -e "${GREEN}=== 自动清理失效账号 (预览模式) ===${NC}"
    echo ""
    echo -e "${YELLOW}这将显示哪些账号会被删除，但不会实际删除${NC}"
    echo ""
    refill clean
    wait_key
}

# 7. 自动清理 (实际删除)
clean_apply() {
    echo -e "${GREEN}=== 自动清理失效账号 (实际删除) ===${NC}"
    echo ""
    echo -e "${RED}警告: 这将实际删除失效账号！${NC}"
    echo -n "确认继续? [y/N]: "
    read -n 1 confirm
    echo ""

    if [[ $confirm =~ ^[Yy]$ ]]; then
        refill clean --apply
    else
        echo "已取消"
    fi
    wait_key
}

# 8. 检查配置
check_config() {
    echo -e "${GREEN}=== 检查配置和环境 ===${NC}"
    echo ""
    refill check
    wait_key
}

# 9. 查看账号统计
show_stats() {
    echo -e "${GREEN}=== 账号统计 ===${NC}"
    echo ""

    if [ -d "$ACCOUNTS_DIR" ]; then
        ACCOUNT_COUNT=$(ls -1 "$ACCOUNTS_DIR"/*.json 2>/dev/null | wc -l | tr -d ' ')
        echo -e "账号总数: ${YELLOW}$ACCOUNT_COUNT${NC}"
        echo ""

        if [ $ACCOUNT_COUNT -gt 0 ]; then
            echo "最近的账号文件:"
            ls -lht "$ACCOUNTS_DIR"/*.json 2>/dev/null | head -5
        else
            echo -e "${YELLOW}账号目录为空${NC}"
            echo ""
            echo "建议操作:"
            echo "  1. 先运行 '5. 同步所有账号'"
            echo "  2. 然后运行 '1. 立即单次续杯'"
        fi
    else
        echo -e "${RED}错误: 账号目录不存在: $ACCOUNTS_DIR${NC}"
    fi

    echo ""

    # 检查输出目录
    if [ -d "out" ]; then
        echo "最近的报告:"
        ls -lht out/*.jsonl out/*.txt 2>/dev/null | head -5
    fi

    wait_key
}

# 详细日志模式
verbose_mode() {
    echo -e "${GREEN}=== 详细日志模式 ===${NC}"
    echo ""
    echo "请选择操作:"
    echo "  1. 单次续杯 (详细日志)"
    echo "  2. 同步账号 (详细日志)"
    echo "  3. 清理账号 (详细日志)"
    echo "  0. 返回"
    echo ""
    echo -n "请选择 [0-3]: "
    read -n 1 choice
    echo ""
    echo ""

    case $choice in
        1)
            refill -v run
            ;;
        2)
            refill -v sync
            ;;
        3)
            refill -v clean
            ;;
        0)
            return
            ;;
        *)
            echo -e "${RED}无效选项${NC}"
            ;;
    esac
    wait_key
}

# 主循环
main() {
    # 切换到脚本所在目录
    cd "$(dirname "$0")"

    # 加载配置
    load_config

    while true; do
        show_menu
        read -n 1 choice
        echo ""
        echo ""

        case $choice in
            1)
                run_refill
                ;;
            2)
                edit_config
                ;;
            3)
                start_scheduler
                ;;
            4)
                stop_scheduler
                ;;
            5)
                sync_accounts
                ;;
            6)
                clean_preview
                ;;
            7)
                clean_apply
                ;;
            8)
                check_config
                ;;
            9)
                show_stats
                ;;
            v|V)
                verbose_mode
                ;;
            0)
                echo -e "${GREEN}再见！${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}无效选项，请重新选择${NC}"
                sleep 1
                ;;
        esac
    done
}

# 运行主程序
main
