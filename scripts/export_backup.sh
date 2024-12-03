#!/bin/bash

API_HOST="http://13.236.126.165:18888"
S3_CONSOLE_URL="https://s3.console.aws.amazon.com/s3/buckets"

ALIYUN_ENVIRONMENTS=("vnnox-us-db" "vnnox-cn-db")
#ALIYUN_ENVIRONMENTS=("vnnox-sg-db" "vnnox-uat" "vnnox-cn-db" "care-cn-db" "care-eu-db" "vnnox-eu-db" "care-us-db" "vnnox-us-db")
ALIYUN_BUCKET="alirds-backup"

declare -A AWS_BUCKET_MAP
AWS_BUCKET_MAP=(
    ["au-mysql8-care"]="novacloud-devops"
    ["au-mysql8-vnnox"]="novacloud-devops"
    ["in-care-mysql"]="in-novacloud-backup"
    ["in-vnnox-mysql"]="in-novacloud-backup"
)
AWS_ENVIRONMENTS=("${!AWS_BUCKET_MAP[@]}")

#WEBHOOK_URL="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=77d13fe6-0047-48bc-803d-904b24590892"

# test
WEBHOOK_URL="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=682ca8af-3592-413f-9b58-a72b3d877cee"

CURRENT_DATE=$(date +%Y-%m-%d)
CURRENT_TIME=$(date +%H:%M:%S)

NOTIFICATION_CONTENT="## RDS 备份导出任务报告 \n\n**执行时间**: ${CURRENT_DATE} ${CURRENT_TIME}\n\n"
ALIYUN_RESULTS=""
AWS_RESULTS=""
SUCCESS_COUNT=0
FAILED_COUNT=0

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

check_response() {
    local response=$1
    local http_code=$2
    local api_name=$3
    
    if [ "$http_code" -eq 200 ]; then
#        log "✅ $api_name 导出成功"
        ((SUCCESS_COUNT++))
        return 0
    else
        log "❌ $api_name 导出失败: $response"
        ((FAILED_COUNT++))
        return 1
    fi
}

generate_aws_s3_link() {
    local env=$1
    local bucket=${AWS_BUCKET_MAP[$env]}
    local date_path=$(date +%Y/%m/%d)
    echo "${S3_CONSOLE_URL}/${bucket}?prefix=aws/${env}/${date_path}/"
}

send_wecom_notification() {
    local title="RDS 备份导出任务报告"
    local color=$([[ $FAILED_COUNT -eq 0 ]] && echo "info" || echo "warning")
    
    # 简化的通知内容
    local simple_content="## RDS 备份导出任务报告 \n\n"
    simple_content="${simple_content}**执行时间**: ${CURRENT_DATE} ${CURRENT_TIME}\n\n"
    simple_content="${simple_content}### 执行统计\n"
    simple_content="${simple_content}- 成功: ${SUCCESS_COUNT}\n"
    simple_content="${simple_content}- 失败: ${FAILED_COUNT}\n\n"
    
    # 只显示失败的任务
    if [ $FAILED_COUNT -gt 0 ]; then
        simple_content="${simple_content}### 失败任务\n"
        if [ ! -z "$ALIYUN_RESULTS" ]; then
            local failed_aliyun=$(echo -e "$ALIYUN_RESULTS" | grep "❌")
            if [ ! -z "$failed_aliyun" ]; then
                simple_content="${simple_content}#### 阿里云\n${failed_aliyun}\n"
            fi
        fi
        
        if [ ! -z "$AWS_RESULTS" ]; then
            local failed_aws=$(echo -e "$AWS_RESULTS" | grep "❌")
            if [ ! -z "$failed_aws" ]; then
                simple_content="${simple_content}#### AWS\n${failed_aws}\n"
            fi
        fi
    fi
    
    # 添加成功任务的汇总
    simple_content="${simple_content}### 成功任务\n"
    if [ ! -z "$ALIYUN_RESULTS" ]; then
        local success_aliyun=$(echo -e "$ALIYUN_RESULTS" | grep "✅" | wc -l)
        simple_content="${simple_content}- 阿里云: ${success_aliyun} 个环境\n"
    fi
    
    if [ ! -z "$AWS_RESULTS" ]; then
        local success_aws=$(echo -e "$AWS_RESULTS" | grep "✅" | wc -l)
        simple_content="${simple_content}- AWS: ${success_aws} 个环境\n"
    fi
    
    # 构造请求体
    local payload=$(cat <<EOF
{
    "msgtype": "markdown",
    "markdown": {
        "content": "${simple_content}"
    }
}
EOF
)
    
    curl -s -X POST -H "Content-Type: application/json" -d "$payload" "$WEBHOOK_URL"
}

export_aliyun_backup() {
    local env=$1
    local failed=0
    local retry_count=0
    local max_retries=3
    
    while [ $retry_count -lt $max_retries ]; do
        log "开始导出阿里云 RDS 备份到 S3 (环境: $env, 尝试次数: $((retry_count + 1)))"
        local http_code
        response=$(curl -s -w "%{http_code}" -X POST "${API_HOST}/alirds/export/s3/${env}")
        http_code=${response: -3}
        response=${response:0:-3}
        
        if check_response "$response" "$http_code" "阿里云 RDS -> S3"; then
            local download_url=$(echo $response | jq -r '.data.backup_download_url')
            ALIYUN_RESULTS="${ALIYUN_RESULTS}\n- ✅ ${env}: [查看备份文件](${download_url})"
            log "✅ ${env} 环境备份导出成功"
            return 0
        else
            ((retry_count++))
            if [ $retry_count -lt $max_retries ]; then
                log "⚠️ ${env} 环境备份导出失败，${retry_count}/${max_retries} 次尝试，等待 60 秒后重试..."
                sleep 60
            else
                ALIYUN_RESULTS="${ALIYUN_RESULTS}\n- ❌ ${env}: 导出失败 (已重试 ${retry_count} 次)"
                log "❌ ${env} 环境备份导出失败 (已重试 ${retry_count} 次)"
                failed=1
            fi
        fi
    done
    
    return $failed
}

export_aws_backup() {
    local env=$1
    local bucket=${AWS_BUCKET_MAP[$env]}
    local failed=0
    local retry_count=0
    local max_retries=3
    
    while [ $retry_count -lt $max_retries ]; do
        log "开始导出 AWS RDS 备份 (环境: $env, 尝试次数: $((retry_count + 1)))"
        local http_code
        response=$(curl -s -w "%{http_code}" -X POST "${API_HOST}/awsrds/export/${env}")
        http_code=${response: -3}
        response=${response:0:-3}
        
        if check_response "$response" "$http_code" "AWS RDS"; then
            local s3_link=$(generate_aws_s3_link "$env")
            AWS_RESULTS="${AWS_RESULTS}\n- ✅ ${env} (${bucket}): [查看备份文件](${s3_link})"
            log "✅ ${env} 环境备份导出成功"
            return 0
        else
            ((retry_count++))
            if [ $retry_count -lt $max_retries ]; then
                log "⚠️ ${env} 环境备份导出失败，${retry_count}/${max_retries} 次尝试，等待 60 秒后重试..."
                sleep 60
            else
                AWS_RESULTS="${AWS_RESULTS}\n- ❌ ${env} (${bucket}): 导出失败 (已重试 ${retry_count} 次)"
                log "❌ ${env} 环境备份导出失败 (已重试 ${retry_count} 次)"
                failed=1
            fi
        fi
    done
    
    return $failed
}

main() {
    local overall_status=0
    
    log "开始备份导出任务 (日期: $CURRENT_DATE)"
    
    for env in "${ALIYUN_ENVIRONMENTS[@]}"; do
        if ! export_aliyun_backup "$env"; then
            log "警告: $env 环境的备份导出失败，继续处理其他环境"
            overall_status=1
        fi
    done
    
    log "开始处理 AWS RDS 备份"
    for env in "${AWS_ENVIRONMENTS[@]}"; do
        if ! export_aws_backup "$env"; then
            log "警告: $env 环境的备份导出失败，继续处理其他环境"
            overall_status=1
        fi
    done
    
    log "所有备份导出任务完成"
    
    # 发送企业微信通知
    send_wecom_notification
    
    return $overall_status
}

# 错误处理
trap 'log "❌ 脚本执行出现错误"; send_wecom_notification; exit 1' ERR

main
exit $? 
