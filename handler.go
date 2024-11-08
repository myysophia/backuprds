// handler.go
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	maxRetries = 5
	retryDelay = 2 * time.Second
)

// backupHandler 处理获取备份下载链接的请求
func backupHandler(c *gin.Context) {
	env := c.Param("env")

	// 根据环境参数获取实例 ID
	instanceID, ok := config.RDS.Instances[env]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid environment"})
		return
	}

	// 添加重试逻辑
	var backupURLs map[string]string
	var err error
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		backupURLs, err = getLastBackupURLs(instanceID)
		if err != nil {
			lastErr = err
			time.Sleep(retryDelay)
			continue
		}

		// 检查是否找到备份
		if backupURLs["BackupDownloadURL"] == "" && backupURLs["BackupIntranetDownloadURL"] == "" {
			if i < maxRetries-1 {
				time.Sleep(retryDelay)
				continue
			}
			c.JSON(http.StatusNotFound, gin.H{
				"message": "no backups found",
				"retries": i + 1,
			})
			return
		}

		// 找到备份，返回结果
		c.JSON(http.StatusOK, gin.H{
			"backup_start_time":            backupURLs["BackupStartTime"],
			"backup_download_url":          backupURLs["BackupDownloadURL"],
			"backup_intranet_download_url": backupURLs["BackupIntranetDownloadURL"],
			"retries":                      i + 1,
		})
		return
	}

	// 所有重试都失败
	if lastErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get backup URLs after retries",
			"details": lastErr.Error(),
			"retries": maxRetries,
		})
		return
	}
}

// awsExportHandler 启动 AWS RDS 快照的导出任务
func awsExportHandler(c *gin.Context) {
	env := c.Param("env")

	// 获取实例 ID
	instanceID, ok := config.RDS.Aws.Instances[env]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid environment"})
		return
	}

	// 示例中，我们假设快照 ARN 是根据实例 ID 构造的，实际情况可能需要查询快照来获取 ARN
	// 注意：此处的 snapshotArn 应该替换为您实际的快照 ARN
	snapshotArn := fmt.Sprintf("arn:aws:rds:%s:account-id:snapshot:%s", config.RDS.Aws.Region, instanceID)

	// 启动快照导出任务
	exportTaskID, err := startRDSSnapshotExport(instanceID, snapshotArn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start export task", "details": err.Error()})
		return
	}

	// 返回导出任务 ID
	c.JSON(http.StatusOK, gin.H{"export_task_id": exportTaskID})
}

// healthCheckHandler 处理健康检查请求
func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
