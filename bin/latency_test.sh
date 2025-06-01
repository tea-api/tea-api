#!/bin/bash

# Tea API 首字时延测试脚本
# 用于测试和验证首字时延优化效果

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认参数
DOMAIN=""
API_KEY=""
COUNT=10
MODEL="gpt-3.5-turbo"
CONCURRENT=1
OUTPUT_FILE=""
VERBOSE=false

# 帮助信息
show_help() {
    echo "Tea API 首字时延测试脚本"
    echo ""
    echo "用法: $0 -d <domain> -k <api_key> [选项]"
    echo ""
    echo "必需参数:"
    echo "  -d, --domain     API域名 (例如: api.example.com)"
    echo "  -k, --key        API密钥"
    echo ""
    echo "可选参数:"
    echo "  -c, --count      测试次数 (默认: 10)"
    echo "  -m, --model      模型名称 (默认: gpt-3.5-turbo)"
    echo "  -p, --parallel   并发数 (默认: 1)"
    echo "  -o, --output     输出文件路径"
    echo "  -v, --verbose    详细输出"
    echo "  -h, --help       显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -d api.example.com -k sk-xxx -c 20 -m gpt-4"
    echo "  $0 -d api.example.com -k sk-xxx -p 5 -o results.json"
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--domain)
                DOMAIN="$2"
                shift 2
                ;;
            -k|--key)
                API_KEY="$2"
                shift 2
                ;;
            -c|--count)
                COUNT="$2"
                shift 2
                ;;
            -m|--model)
                MODEL="$2"
                shift 2
                ;;
            -p|--parallel)
                CONCURRENT="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                echo "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # 检查必需参数
    if [[ -z "$DOMAIN" || -z "$API_KEY" ]]; then
        echo -e "${RED}错误: 缺少必需参数${NC}"
        show_help
        exit 1
    fi
}

# 记录日志
log() {
    local level=$1
    shift
    local message="$@"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case $level in
        "INFO")
            echo -e "${GREEN}[INFO]${NC} ${timestamp} - $message"
            ;;
        "WARN")
            echo -e "${YELLOW}[WARN]${NC} ${timestamp} - $message"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} ${timestamp} - $message"
            ;;
        "DEBUG")
            if [[ "$VERBOSE" == "true" ]]; then
                echo -e "${BLUE}[DEBUG]${NC} ${timestamp} - $message"
            fi
            ;;
    esac
}

# 单次测试
single_test() {
    local test_id=$1
    local start_time=$(date +%s.%N)
    
    log "DEBUG" "开始测试 #$test_id"
    
    # 构建请求数据
    local request_data=$(cat <<EOF
{
    "model": "$MODEL",
    "messages": [
        {
            "role": "user",
            "content": "请说'你好'"
        }
    ],
    "stream": true,
    "max_tokens": 10
}
EOF
)
    
    # 执行请求并测量时间
    local result=$(curl -s -w "%{http_code}|%{time_total}|%{time_connect}|%{time_starttransfer}" \
                       -X POST \
                       -H "Content-Type: application/json" \
                       -H "Authorization: Bearer $API_KEY" \
                       -d "$request_data" \
                       "https://$DOMAIN/v1/chat/completions" \
                       -o /dev/null)
    
    local end_time=$(date +%s.%N)
    local total_time=$(echo "$end_time - $start_time" | bc -l)
    
    # 解析结果
    IFS='|' read -r http_code time_total time_connect time_starttransfer <<< "$result"
    
    # 计算首字时延（time_starttransfer 是到第一个字节的时间）
    local ttft=$(echo "$time_starttransfer * 1000" | bc -l)
    local total_ms=$(echo "$time_total * 1000" | bc -l)
    local connect_ms=$(echo "$time_connect * 1000" | bc -l)
    
    log "DEBUG" "测试 #$test_id 完成: HTTP=$http_code, TTFT=${ttft}ms, Total=${total_ms}ms"
    
    # 输出结果
    echo "$test_id,$http_code,$ttft,$total_ms,$connect_ms"
}

# 并发测试
concurrent_test() {
    local batch_size=$1
    local batch_count=$2
    
    log "INFO" "开始并发测试: 批次大小=$batch_size, 批次数量=$batch_count"
    
    for ((batch=1; batch<=batch_count; batch++)); do
        log "DEBUG" "执行批次 $batch/$batch_count"
        
        # 启动并发任务
        for ((i=1; i<=batch_size; i++)); do
            local test_id=$(((batch-1)*batch_size + i))
            single_test $test_id &
        done
        
        # 等待当前批次完成
        wait
        
        # 批次间短暂延迟
        sleep 0.1
    done
}

# 统计分析
analyze_results() {
    local results_file=$1
    
    log "INFO" "分析测试结果..."
    
    # 使用awk进行统计分析
    awk -F',' '
    BEGIN {
        count = 0
        sum_ttft = 0
        sum_total = 0
        sum_connect = 0
        min_ttft = 999999
        max_ttft = 0
        success_count = 0
    }
    NR > 1 {  # 跳过标题行
        count++
        if ($2 == 200) success_count++
        
        ttft = $3
        total = $4
        connect = $5
        
        sum_ttft += ttft
        sum_total += total
        sum_connect += connect
        
        if (ttft < min_ttft) min_ttft = ttft
        if (ttft > max_ttft) max_ttft = ttft
        
        # 存储所有TTFT值用于计算中位数
        ttft_values[count] = ttft
    }
    END {
        if (count > 0) {
            avg_ttft = sum_ttft / count
            avg_total = sum_total / count
            avg_connect = sum_connect / count
            success_rate = (success_count / count) * 100
            
            # 计算中位数
            n = asort(ttft_values)
            if (n % 2 == 1) {
                median_ttft = ttft_values[(n+1)/2]
            } else {
                median_ttft = (ttft_values[n/2] + ttft_values[n/2+1]) / 2
            }
            
            printf "\n=== 测试结果统计 ===\n"
            printf "总请求数: %d\n", count
            printf "成功请求数: %d\n", success_count
            printf "成功率: %.2f%%\n", success_rate
            printf "\n=== 首字时延 (TTFT) ===\n"
            printf "平均值: %.2f ms\n", avg_ttft
            printf "中位数: %.2f ms\n", median_ttft
            printf "最小值: %.2f ms\n", min_ttft
            printf "最大值: %.2f ms\n", max_ttft
            printf "\n=== 其他指标 ===\n"
            printf "平均连接时间: %.2f ms\n", avg_connect
            printf "平均总时间: %.2f ms\n", avg_total
        }
    }' "$results_file"
}

# 主函数
main() {
    parse_args "$@"
    
    log "INFO" "开始首字时延测试"
    log "INFO" "域名: $DOMAIN"
    log "INFO" "模型: $MODEL"
    log "INFO" "测试次数: $COUNT"
    log "INFO" "并发数: $CONCURRENT"
    
    # 创建临时结果文件
    local temp_results=$(mktemp)
    
    # 写入CSV标题
    echo "test_id,http_code,ttft_ms,total_ms,connect_ms" > "$temp_results"
    
    # 执行测试
    if [[ $CONCURRENT -eq 1 ]]; then
        # 串行测试
        for ((i=1; i<=COUNT; i++)); do
            single_test $i >> "$temp_results"
        done
    else
        # 并发测试
        local batch_count=$(((COUNT + CONCURRENT - 1) / CONCURRENT))
        concurrent_test $CONCURRENT $batch_count >> "$temp_results"
    fi
    
    # 分析结果
    analyze_results "$temp_results"
    
    # 保存结果文件
    if [[ -n "$OUTPUT_FILE" ]]; then
        cp "$temp_results" "$OUTPUT_FILE"
        log "INFO" "结果已保存到: $OUTPUT_FILE"
    fi
    
    # 清理临时文件
    rm -f "$temp_results"
    
    log "INFO" "测试完成"
}

# 检查依赖
check_dependencies() {
    local deps=("curl" "bc" "awk")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log "ERROR" "缺少依赖: $dep"
            exit 1
        fi
    done
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    check_dependencies
    main "$@"
fi
