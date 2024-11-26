# Nova RDS 跨云灾备系统

## 项目简介
Nova RDS 跨云灾备系统是一个专为阿里云和AWS RDS设计的数据库备份解决方案，旨在提供一个简单、高效的方式来管理和自动化跨云数据库备份。支持通过REST API接口进行备份管理，并提供企业微信告警通知，确保备份过程的可靠性和及时性。

## 功能特性
### 核心功能
- **自动备份**：自动化备份阿里云和AWS RDS数据库
- **跨云管理**：支持在阿里云和AWS之间进行备份数据的迁移和管理
- **灵活的备份策略**：通过REST API接口自定义备份频率、备份时间等
- **监控与报警**：实时监控备份状态，并在备份失败时发送企微报警通知

### 特色功能
- **多环境支持**：支持多个环境，可独立配置
- **备份验证**：自动验证备份的完整性和可用性
- **失败重试**：针对SDK获取实例支持自动重试机制
- **API文档**：集成Swagger文档，便于接口调试和集成

## 快速开始

### 环境要求
- Go 1.22+
- 阿里云访问密钥
- AWS访问密钥
- 企业微信机器人 Token（用于告警）

### 安装步骤
1. 克隆仓库到本地
```bash
git clone xxx
cd backuprds
```

2. 安装依赖
```bash
go mod tidy
```

3. 配置环境变量
```bash
export AWS_ACCESS_KEY_ID=your_aws_access_key
export AWS_SECRET_ACCESS_KEY=your_aws_secret_key
export ALIYUN_ACCESS_KEY_ID=your_aliyun_access_key
export ALIYUN_ACCESS_KEY_SECRET=your_aliyun_access_key_secret
export WEWORK_BOT_KEY=your_wework_bot_key
```

4. 修改配置文件
```bash
# 编辑 config.yaml 添加必要的配置
cp config/config.example.yaml config/config.yaml
```

5. 编译
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-w" -o backuprdsv1.1.1
```

6. 运行

```bash
./backuprdsv1.1.6 --help
Nova RDS 跨云灾备系统支持阿里云和AWS RDS的备份管理

Usage:
  backuprds [flags]

Flags:
      --config string       配置文件路径 (default "config/config.yaml")
  -h, --help                help for backuprds
      --log.config string   日志配置文件路径 (default "config/logger.yaml")
      --log.level string    日志级别
  -p, --port string         Web服务端口号 (default "8080")

# 指定18888端口启动，日志为info级别
./backuprdsv1.1.6 -p 18888 --log.leve=info

```



## 配置说明

### 配置文件结构
```yaml
server:
  port: 8080
  mode: release

rds:
  aliyun:
    instances:
      prod:
        id: "rm-xxx"
        region: "cn-hangzhou"
    s3export:
      region: "ap-southeast-1"
      bucketname: "backup-bucket"
  
  aws:
    instances:
      prod:
        id: "db-xxx"
        region: "ap-southeast-1"

alert:
  wework:
    enabled: true
    botkey: "xxx"
    retry_times: 3
```

### 4.2 日志配置文件

```yaml
more config/logger.yaml
level: "info"              # 默认日志级别
format: "text"            # 日志格式 (json/text)
output:
  console: true           # 是否输出到控制台
  files:                  # 文件输出配置
    - level: "debug"
      path: "logs/debug.log"
      max_size: 100       # MB
      max_age: 7          # 天
      max_backups: 10     # 保留的旧文件个数
    - level: "info"
      path: "logs/info.log"
      max_size: 100
      max_age: 7
      max_backups: 10
    - level: "error"
      path: "logs/error.log"
      max_size: 100
      max_age: 7
      max_backups: 10
hooks:
  wecom:                  # 企业微信告警配置
    enabled: true
    levels: ["error", "fatal"]
    webhook_url: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=682ca8af-3592-413f-9b58-a72b3d877cee"
```



## API 接口说明

### 阿里云 RDS 接口
- `GET /alirds/{env}` - 获取指定环境的RDS备份列表
- `POST /alirds/export/s3/{env}` - 将RDS备份上传至S3
- `GET /alirds/s3config` - 获取S3配置信息

### AWS RDS 接口
- `GET /awsrds/{env}` - 获取指定环境的RDS快照列表
- `POST /awsrds/export/{env}` - 导出RDS快照

### 系统接口
- `GET /health` - 健康检查接口
- `GET /instances` - 获取所有实例配置

## 告警说明

### 告警级别
- **ERROR**: 备份失败、上传失败等严重错误
- **WARN**: 备份延迟、性能警告等
- **INFO**: 备份完成通知

### 告警模板
```json
RDS跨云异地备份任务报告
执行时间: 2024-11-25 05:00:01
执行统计
- 成功: 10
- 失败: 2
失败任务
阿里云
- ❌ XXXX: 导出失败
成功任务
- 阿里云: 6 个环境
- AWS: 4 个环境
```

## 常见问题

### 1. 备份上传失败
- 检查AWS凭证是否正确
- 确认S3存储桶权限配置

### 2. 备份获取失败
- 检查RDS实例状态
- 确认访问密钥权限
- 验证网络连接状态

## 性能优化建议
1. 合理配置备份时间窗口
2. 适当调整并发上传数
3. 根据实际需求配置重试策略
4. 定期清理过期备份

## 开发指南

### 本地开发
1. 安装开发依赖
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

2. 生成API文档
```bash
swag init
```

3. 运行测试
```bash
go test ./...
```

### 代码规范
- 遵循Go标准项目布局
- 使用gofmt格式化代码
- 添加必要的注释

## 贡献指南


## 致谢
