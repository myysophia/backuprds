# 查看RDS实例备份集列表完整工程示例

该项目为DescribeBackups的完整工程示例。

该示例**无法在线调试**，如需调试可下载到本地后替换 [AK](https://usercenter.console.aliyun.com/#/manage/ak) 以及参数后进行调试。

## 运行条件

- 下载并解压需要语言的代码;


- 在阿里云帐户中获取您的 [凭证](https://usercenter.console.aliyun.com/#/manage/ak) 并通过它替换下载后代码中的 ACCESS_KEY_ID 以及 ACCESS_KEY_SECRET;

- 执行对应语言的构建及运行语句

## 执行步骤

下载的代码包，在根据自己需要更改代码中的参数和 AK 以后，可以在**解压代码所在目录下**按如下的步骤执行：

- *Go 环境版本必须不低于 1.10.x*
- *安装 SDK 核心库 OpenAPI*
```sh
go get github.com/alibabacloud-go/darabonba-openapi/v2/client
```
- *执行命令*
```sh
GOPROXY=https://goproxy.cn,direct go run ./main
```
## 使用的 API

-  DescribeBackups：该接口用于查看RDS实例的备份集列表。 更多信息可参考：[文档](https://next.api.aliyun.com/document/Rds/2014-08-15/DescribeBackups)

## API 返回示例

*实际输出结构可能稍有不同，属于正常返回；下列输出值仅作为参考，以实际调用为准*


- JSON 格式 
```js
{
  "Items": {
    "Backup": [
      {
        "BackupDownloadLinkByDB": {
          "BackupDownloadLinkByDB": [
            {
              "DataBase": "dbs",
              "DownloadLink": "https://cn-hangzhou.bak.rds.aliyuncs.com/custins53664665/hins18676859_2021072909473127987849.zip?Expires=****&dbList=tb1",
              "IntranetDownloadLink": "https://cn-hangzhou-internal.bak.rds.aliyuncs.com/custins53664665/hins18676859_2021072909473127987849.zip?Expires=****&dbList=tb1"
            }
          ]
        },
        "BackupDownloadURL": "http://rdsbak-hz-v3.oss-cn-hangzhou.aliyuncs.com/****",
        "BackupEndTime": "2019-02-13T12:20:00Z",
        "BackupId": "321020562",
        "BackupInitiator": "System",
        "BackupIntranetDownloadURL": "http://rdsbak-hz-v3.oss-cn-hangzhou-internal.aliyuncs.com/****",
        "BackupMethod": "Physical",
        "BackupMode": "Automated",
        "BackupSize": 2167808,
        "BackupStartTime": "2019-02-03T12:20:00Z",
        "BackupStatus": "Success",
        "BackupType": "FullBackup",
        "Checksum": "1835830439****",
        "ConsistentTime": 1576506856,
        "CopyOnlyBackup": "0",
        "DBInstanceId": "rm-uf6wjk5****",
        "Encryption": "{}",
        "Engine": "MySQL",
        "EngineVersion": "8.0",
        "HostInstanceID": "5882781",
        "IsAvail": 1,
        "MetaStatus": "OK",
        "StorageClass": "0",
        "StoreStatus": "Disabled"
      }
    ]
  },
  "PageNumber": "1",
  "PageRecordCount": "30",
  "RequestId": "1A6D328C-84B8-40DC-BF49-6C73984D7494",
  "TotalEcsSnapshotSize": 0,
  "TotalRecordCount": "100"
}
```

