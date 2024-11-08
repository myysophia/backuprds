// aws_rds_client.go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// createAWSClient 创建 RDS 客户端
func createAWSClient() (*rds.Client, error) {
	// 从配置中加载 AWS 凭证和区域信息
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(config.RDS.Aws.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}
	return rds.NewFromConfig(cfg), nil
}

// startRDSSnapshotExport 启动 RDS 快照导出任务
func startRDSSnapshotExport(instanceID string, snapshotArn string) (string, error) {
	client, err := createAWSClient()
	if err != nil {
		return "", fmt.Errorf("failed to create AWS RDS client: %v", err)
	}

	// 生成唯一的导出任务标识符
	taskID := fmt.Sprintf("%s%s-%d", config.RDS.Aws.ExportTask.ExportTaskIdentifierPrefix, instanceID, time.Now().Unix())

	// 准备导出任务请求参数
	exportInput := &rds.StartExportTaskInput{
		ExportTaskIdentifier: aws.String(taskID),
		IamRoleArn:           aws.String(config.RDS.Aws.ExportTask.IamRoleArn),
		KmsKeyId:             aws.String(config.RDS.Aws.ExportTask.KmsKeyId),
		S3BucketName:         aws.String(config.RDS.Aws.ExportTask.S3BucketName),
		SourceArn:            aws.String(snapshotArn),
		S3Prefix:             aws.String(instanceID), // 可以使用实例 ID 作为 S3 前缀
	}

	// 启动导出任务
	resp, err := client.StartExportTask(context.TODO(), exportInput)
	if err != nil {
		return "", fmt.Errorf("failed to start export task: %v", err)
	}

	// 返回任务标识符以供查询
	return aws.ToString(resp.ExportTaskIdentifier), nil
}
