// Package aws provides AWS RDS and S3 related operations for backup management
package aws

import (
	"backuprds/internal/logger"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// createAWSClient 创建 RDS 客户端
func createAWSClient(region string) (*rds.Client, error) {
	// 从环境变量获取 AWS 凭证
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("missing required environment variables: AWS_ACCESS_KEY_ID or AWS_SECRET_ACCESS_KEY")
	}

	// 使用传入的 region 创建配置
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"", // token可以为空
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	return rds.NewFromConfig(cfg), nil
}

// StartRDSSnapshotExport 启动 RDS 快照导出任务
func StartRDSSnapshotExport(
	instanceID string,
	snapshotArn string,
	region string,
	iamRoleArn string,
	kmsKeyId string,
	s3BucketName string,
	s3Prefix string,
) (string, error) {
	client, err := createAWSClient(region)
	if err != nil {
		return "", fmt.Errorf("startRDSSnapshotExport funcation failed to create AWS RDS client: %v", err)
	}

	// 截取实例ID的关键部分
	shortInstanceID := instanceID
	if len(instanceID) > 20 {
		parts := strings.Split(instanceID, ":")
		shortInstanceID = parts[len(parts)-1]
		if len(shortInstanceID) > 20 {
			shortInstanceID = shortInstanceID[len(shortInstanceID)-20:]
		}
	}

	// 生成更短的导出任务标识符
	exportTaskIdentifier := fmt.Sprintf("exp-%s-%s",
		shortInstanceID,
		time.Now().Format("0102-1504"))

	// 构建完整的 S3 前缀路径
	fullS3Prefix := s3Prefix
	if s3Prefix != "" {
		fullS3Prefix = strings.TrimSuffix(s3Prefix, "/") + "/"
	}

	input := &rds.StartExportTaskInput{
		ExportTaskIdentifier: aws.String(exportTaskIdentifier),
		IamRoleArn:           aws.String(iamRoleArn),
		KmsKeyId:             aws.String(kmsKeyId),
		S3BucketName:         aws.String(s3BucketName),
		S3Prefix:             aws.String(fullS3Prefix),
		SourceArn:            aws.String(snapshotArn),
	}

	logger.LogInfo("Starting export task",
		logger.Any("params", input))

	result, err := client.StartExportTask(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("failed to start export task: %v", err)
	}

	return aws.ToString(result.ExportTaskIdentifier), nil
}

// GetLatestSnapshotInfo 获取最新的 AWS RDS 快照信息
func GetLatestSnapshotInfo(instanceID string, region string) (map[string]string, error) {
	logger.LogInfo("Fetching latest snapshot info",
		logger.String("instance_id", instanceID),
		logger.String("region", region))

	client, err := createAWSClient(region)
	if err != nil {
		logger.LogError("Failed to create AWS client",
			logger.Error(err),
			logger.String("region", region))
		return nil, fmt.Errorf("failed to create AWS RDS client: %v", err)
	}

	input := &rds.DescribeDBSnapshotsInput{
		DBInstanceIdentifier: aws.String(instanceID),
		SnapshotType:         aws.String("automated"),
		MaxRecords:           aws.Int32(20),
		IncludeShared:        aws.Bool(true),
		IncludePublic:        aws.Bool(true),
	}

	logger.LogDebug("Describing DB snapshots",
		logger.String("instance_id", instanceID),
		logger.Any("input", input))

	resp, err := client.DescribeDBSnapshots(context.TODO(), input)
	if err != nil {
		// 详细的错误信息处理
		return nil, fmt.Errorf("failed to describe DB snapshots: %v (instanceID: %s)", err, instanceID)
	}

	logger.LogDebug("Found snapshots",
		logger.Int("count", len(resp.DBSnapshots)),
		logger.String("instance_id", instanceID))

	// 获取最新快照
	var latestSnapshot *types.DBSnapshot
	for _, snapshot := range resp.DBSnapshots {
		if snapshot.Status != nil && *snapshot.Status == "available" {
			if latestSnapshot == nil || snapshot.SnapshotCreateTime.After(*latestSnapshot.SnapshotCreateTime) {
				latestSnapshot = &snapshot
			}
		}
	}

	if latestSnapshot == nil {
		logger.LogWarn("No available snapshots found",
			logger.String("instance_id", instanceID))
		return map[string]string{
			"SnapshotArn":        "",
			"SnapshotCreateTime": "",
			"SnapshotId":         "",
			"Status":             "",
		}, nil
	}

	return map[string]string{
		"SnapshotArn":        aws.ToString(latestSnapshot.DBSnapshotArn),
		"SnapshotCreateTime": latestSnapshot.SnapshotCreateTime.String(),
		"SnapshotId":         aws.ToString(latestSnapshot.DBSnapshotIdentifier),
		"Status":             aws.ToString(latestSnapshot.Status),
	}, nil
}
