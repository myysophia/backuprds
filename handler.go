// handler.go
package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

const (
	maxRetries = 1
	retryDelay = 2 * time.Second
)

// backupHandler 处理获取备份下载链接的请求
func backupHandler(c *gin.Context) {
	env := c.Param("env")

	// 根据环境参数获取实例 ID
	instanceID, ok := configs.RDS.Aliyun.Instances[env]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid environment"})
		return
	}

	// 添加重试逻辑
	var backupURLs map[string]string
	var err error
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		backupURLs, err = getLastBackupURLs(instanceID.ID)
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

func awsBackupHandler(c *gin.Context) {
	env := c.Param("env")

	// 获取实例配置
	instanceConfig, ok := configs.RDS.Aws.Instances[env]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid environment"})
		return
	}

	log.Printf("Fetching snapshots for instance: %s in region: %s", instanceConfig.ID, instanceConfig.Region)

	// 获取最新快照信息
	snapshotInfo, err := getLatestSnapshotInfo(instanceConfig.ID, instanceConfig.Region)
	if err != nil {
		log.Printf("Error getting snapshot info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "failed to get snapshot info",
			"details":    err.Error(),
			"instanceId": instanceConfig.ID,
			"region":     instanceConfig.Region,
		})
		return
	}

	// 检查是否找到快照
	if snapshotInfo["SnapshotArn"] == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"message":    "no snapshots found",
			"instanceId": instanceConfig.ID,
			"region":     instanceConfig.Region,
		})
		return
	}

	// 返回快照信息
	c.JSON(http.StatusOK, gin.H{
		"snapshot_create_time": snapshotInfo["SnapshotCreateTime"],
		"snapshot_arn":         snapshotInfo["SnapshotArn"],
		"snapshot_id":          snapshotInfo["SnapshotId"],
		"status":               snapshotInfo["Status"],
		"instance_id":          instanceConfig.ID,
		"region":               instanceConfig.Region,
	})
}

// awsExportHandler 启动 AWS RDS 快照的导出任务
func awsExportHandler(c *gin.Context) {
	env := c.Param("env")

	// 获取实例配置
	instanceConfig, ok := configs.RDS.Aws.Instances[env]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid environment"})
		return
	}

	log.Printf("Starting export task for instance: %s in region: %s", instanceConfig.ID, instanceConfig.Region)

	// 先获取最新的快照信息
	snapshotInfo, err := getLatestSnapshotInfo(instanceConfig.ID, instanceConfig.Region)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "failed to get snapshot info",
			"details":    err.Error(),
			"instanceId": instanceConfig.ID,
			"region":     instanceConfig.Region,
		})
		return
	}

	// 检查是否找到快照
	if snapshotInfo["SnapshotArn"] == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"message":    "no snapshots found",
			"instanceId": instanceConfig.ID,
			"region":     instanceConfig.Region,
		})
		return
	}

	// 启动快照导出任务
	exportTaskID, err := startRDSSnapshotExport(
		instanceConfig.ID,
		snapshotInfo["SnapshotArn"],
		instanceConfig.Region,
		configs.RDS.Aws.ExportTask.IamRoleArn,
		configs.RDS.Aws.ExportTask.KmsKeyId,
		configs.RDS.Aws.ExportTask.S3BucketName,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "failed to start export task",
			"details":    err.Error(),
			"instanceId": instanceConfig.ID,
			"region":     instanceConfig.Region,
		})
		return
	}

	// 返回导出任务 ID
	c.JSON(http.StatusOK, gin.H{
		"export_task_id": exportTaskID,
		"snapshot_arn":   snapshotInfo["SnapshotArn"],
		"instance_id":    instanceConfig.ID,
		"region":         instanceConfig.Region,
	})
}

// healthCheckHandler 处理健康检查请求
func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
