package main

import (
	"context"
	"fmt"
	"log"
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

func uploadBackupToS3(backupURL, bucketName, region, env, backupTime string) (*UploadResult, error) {
	// 从环境变量获取AWS凭证
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
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	// 创建S3客户端
	s3Client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(s3Client)

	// 下载备份文件
	resp, err := http.Get(backupURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download backup: %v", err)
	}
	defer resp.Body.Close()

	// 生成S3密钥
	timestamp := time.Now().Format("20060102-150405")
	s3Key := path.Join(env, fmt.Sprintf("backup-%s-%s.xb", env, timestamp))

	log.Printf(bucketName, s3Key, resp.Body)
	// 上传到S3
	result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &s3Key,
		Body:   resp.Body,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %v", err)
	}

	return &UploadResult{
		S3Key:    s3Key,
		Location: result.Location,
	}, nil
}
