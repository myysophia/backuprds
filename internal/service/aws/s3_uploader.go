package aws

import (
	"backuprds/internal/logger"
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type UploadResult struct {
	S3Key    string
	Location string
}

func UploadBackupToS3(backupURL, bucketName, region, env, backupTime string) (*UploadResult, error) {
	// 从环境变量获取AWS凭证
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if accessKey == "" || secretKey == "" {
		logger.LogError("Missing AWS credentials",
			logger.String("service", "s3_uploader"))
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
		logger.LogError("Failed to load AWS SDK config",
			logger.Error(err),
			logger.String("region", region))
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	// 创建S3客户端
	s3Client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(s3Client, func(u *manager.Uploader) {
		// 设置分片大小为 100MB
		u.PartSize = 200 * 1024 * 1024
		// 设置并发数
		u.Concurrency = 10
		// 启用分片上传
		u.LeavePartsOnError = false
	})

	// 下载备份文件
	logger.LogInfo("Starting backup download",
		logger.String("url", backupURL),
		logger.String("bucket", bucketName),
		logger.String("region", region),
		logger.String("env", env))
	resp, err := http.Get(backupURL)
	if err != nil {
		logger.LogError("Failed to download backup",
			logger.Error(err),
			logger.String("url", backupURL))
		return nil, fmt.Errorf("failed to download backup: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.LogError("Non-OK HTTP status while downloading",
			logger.Int("status_code", resp.StatusCode),
			logger.String("url", backupURL))
		return nil, fmt.Errorf("failed to download backup, status code: %d", resp.StatusCode)
	}

	// 生成S3密钥
	timestamp := time.Now().Format("20060102-150405")
	s3Key := path.Join(env, fmt.Sprintf("backup-%s-%s.xb", env, timestamp))

	// 上传到S3
	logger.LogInfo("Starting upload to S3",
		logger.String("bucket", bucketName),
		logger.String("key", s3Key),
		logger.String("region", region))
	result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &s3Key,
		Body:   resp.Body,
	})
	if err != nil {
		logger.LogError("Failed to upload to S3",
			logger.Error(err),
			logger.String("bucket", bucketName),
			logger.String("key", s3Key))
		return nil, fmt.Errorf("failed to upload to S3: %v", err)
	}

	logger.LogInfo("Upload completed successfully",
		logger.String("location", result.Location))
	return &UploadResult{
		S3Key:    s3Key,
		Location: result.Location,
	}, nil
}
