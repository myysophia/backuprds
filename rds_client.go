// rds_client.go
package main

import (
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	rds20140815 "github.com/alibabacloud-go/rds-20140815/v8/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"os"
)

// CreateClient 创建 RDS 客户端
func CreateClient() (*rds20140815.Client, error) {
	accessKey := os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
	accessSecret := os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")

	if accessKey == "" || accessSecret == "" {
		return nil, fmt.Errorf("missing required environment variables: ALIBABA_CLOUD_ACCESS_KEY_ID or ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	}

	config := &openapi.Config{
		AccessKeyId:     tea.String(accessKey),
		AccessKeySecret: tea.String(accessSecret),
	}
	config.Endpoint = tea.String("rds.aliyuncs.com")
	return rds20140815.NewClient(config)
}

// getLastBackupURLs 获取最新备份文件的下载链接，包括内网和公网
func getLastBackupURLs(instanceID string) (map[string]string, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create RDS client: %v", err)
	}

	// 配置 DescribeBackups 请求
	describeBackupsRequest := &rds20140815.DescribeBackupsRequest{
		DBInstanceId: tea.String(instanceID),
	}
	runtime := &util.RuntimeOptions{}

	// 调用 DescribeBackupsWithOptions 获取备份信息
	resp, err := client.DescribeBackupsWithOptions(describeBackupsRequest, runtime)
	if err != nil {
		// 错误处理，输出详细的错误信息
		var sdkErr = &tea.SDKError{}
		if _t, ok := err.(*tea.SDKError); ok {
			sdkErr = _t
		} else {
			sdkErr.Message = tea.String(err.Error())
		}
		return nil, fmt.Errorf("API request error: %s", tea.StringValue(sdkErr.Message))
	}

	// 检查是否有备份
	if len(resp.Body.Items.Backup) == 0 {
		return map[string]string{
			"BackupStartTime":           "",
			"BackupDownloadURL":         "",
			"BackupIntranetDownloadURL": "",
		}, nil
	}

	// 获取最新备份的公网和内网下载 URL
	return map[string]string{
		"BackupStartTime":           tea.StringValue(resp.Body.Items.Backup[0].BackupStartTime),
		"BackupDownloadURL":         tea.StringValue(resp.Body.Items.Backup[0].BackupDownloadURL),
		"BackupIntranetDownloadURL": tea.StringValue(resp.Body.Items.Backup[0].BackupIntranetDownloadURL),
	}, nil
}
