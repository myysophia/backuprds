package models

// S3Config 定义S3配置信息
type S3Config struct {
	Region     string `json:"region" example:"ap-southeast-1"`
	BucketName string `json:"bucket_name" example:"my-backup-bucket"`
}

// BackupInfo 定义备份信息
type BackupInfo struct {
	BackupStartTime           string `json:"backup_start_time" example:"2024-03-20 10:00:00"`
	BackupDownloadURL         string `json:"backup_download_url"`
	BackupIntranetDownloadURL string `json:"backup_intranet_download_url"`
	Retries                   int    `json:"retries" example:"1"`
}

// ErrorResponse 定义错误响应
type ErrorResponse struct {
	Error   string `json:"error" example:"operation failed"`
	Details string `json:"details,omitempty" example:"detailed error message"`
}
