



```mermaid
graph TD
    subgraph Client[客户端层]
        Browser[浏览器]
        APIClient[API客户端]
    end

    subgraph WebServer[Web服务器层]
        GinEngine[Gin引擎]
        Swagger[Swagger文档]
    end

    subgraph Handlers[controller层]
        direction LR
        subgraph AliHandlers[aliyun rds handler]
            AliBackupHandler[备份查询]
            AliExportHandler[备份上传到S3]
        end
        subgraph AwsHandlers[AWS rds handler]
            AwsBackupHandler[快照查询]
            AwsExportHandler[快照导出]
        end
        subgraph SystemHandlers[系统处理器]
            HealthHandler[健康检查接口]
            ConfigHandler[配置查询接口]
        end
    end

    subgraph Services[服务层]
        direction LR
        subgraph AliyunService[阿里云RDS SDK]
            GetBackupURLs[获取backupURL]
            UploadToS3[分片上传至S3]
        end
        subgraph AWSService[AWS RDS SDK]
            GetSnapshot[获取快照]
            ExportSnapshot[导出快照]
        end
    end

    subgraph Storage[存储层]
        S3[AWS S3]
        subgraph S3Paths[S3目录]
            MySQLPath[mysql]
        end
        OSS[阿里云OSS]
        subgraph OSSPaths[OSS目录]
            IOTDBPath[iotdb-backup/mysql]
        end
    end

    subgraph Sync[实时同步]
        Lambda[AWS Lambda]
        S3Trigger[S3事件触发器]
    end

    subgraph AlertSystem[告警系统]
        AlertHandler[告警处理器]
        AlertRules[告警规则]
        WeworkBot[企业微信机器人]
    end

    subgraph Config[配置管理]
        ConfigLoader[viper配置解析模块]
        Logger[Zap日志模块]
    end

    Browser -->|HTTP请求| GinEngine
    APIClient -->|HTTP请求| GinEngine
    GinEngine -->|API文档| Swagger
    GinEngine -->|路由分发| Handlers

    AliBackupHandler -->|调用| GetBackupURLs
    AliExportHandler -->|调用| UploadToS3
    AwsBackupHandler -->|调用| GetSnapshot
    AwsExportHandler -->|调用| ExportSnapshot

    GetBackupURLs -->|验证| ValidateBackup
    GetSnapshot -->|验证| ValidateSnapshot
    
    ValidateBackup -->|失败| AlertHandler
    ValidateSnapshot -->|失败| AlertHandler
    UploadToS3 -->|失败| AlertHandler
    ExportSnapshot -->|失败| AlertHandler
    
    AlertHandler -->|检查| AlertRules
    AlertHandler -->|发送| WeworkBot
    
    Services -->|读取配置| ConfigLoader
    Services -->|记录日志| Logger
    Handlers -->|记录日志| Logger

    ExportSnapshot -->|导出| S3
    S3 -->|监听/mysql| S3Trigger
    S3Trigger -->|触发| Lambda
    Lambda -->|同步| OSS

    classDef client fill:#f9f,stroke:#333,stroke-width:2px
    classDef server fill:#ccf,stroke:#333,stroke-width:2px
    classDef handler fill:#cfc,stroke:#333,stroke-width:2px
    classDef service fill:#fcf,stroke:#333,stroke-width:2px
    classDef storage fill:#fdf,stroke:#333,stroke-width:2px
    classDef sync fill:#dff,stroke:#333,stroke-width:2px
    classDef alert fill:#ffc,stroke:#333,stroke-width:2px
    classDef config fill:#cff,stroke:#333,stroke-width:2px
    
    class Browser,APIClient client
    class GinEngine,Swagger server
    class AliBackupHandler,AliExportHandler,AwsBackupHandler,AwsExportHandler,HealthHandler,ConfigHandler handler
    class GetBackupURLs,UploadToS3,GetSnapshot,ExportSnapshot,ValidateBackup,ValidateSnapshot service
    class S3,OSS,MySQLPath,IOTDBPath storage
    class Lambda,S3Trigger sync
    class AlertHandler,AlertRules,WeworkBot alert
    class ConfigLoader,Logger config

```





```mermaid
sequenceDiagram
    participant C as Client
    participant A as API Server
    participant AL as Aliyun Service
    participant AW as AWS Service
    participant S3 as AWS S3
    
    C->>A: Request Backup
    A->>AL: Get RDS Backup
    AL-->>A: Backup URL
    A->>AW: Upload to S3
    AW->>S3: Store Backup
    S3-->>AW: Upload Complete
    AW-->>A: Success
    A-->>C: Backup Complete
```

