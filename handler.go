// handler.go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxRetries = 1
	retryDelay = 2 * time.Second
)

// backupHandler godoc
// @Summary      获取阿里云RDS备份下载链接
// @Description  获取指定环境的阿里云RDS最新备份下载链接
// @Tags         阿里云RDS
// @Accept       json
// @Produce      json
// @Param        env  path      string  true  "环境名称"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /alirds/{env} [get]
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

		// 找到备份返回结果
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

// awsExportHandler godoc
// @Summary      启动AWS RDS快照导出任务
// @Description  为指定环境的AWS RDS实例启动快照导出任务
// @Tags         AWS RDS
// @Accept       json
// @Produce      json
// @Param        env  path      string  true  "环境名称"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /awsrds/export/{env} [post]
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
		instanceConfig.KmsKeyId,
		instanceConfig.S3BucketName,
		configs.RDS.Aws.ExportTask.S3Prefix,
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
		"kms_key_id":     instanceConfig.KmsKeyId,
		"s3_bucket_name": instanceConfig.S3BucketName,
	})
}

// healthCheckHandler godoc
// @Summary      健康检查
// @Description  API服务健康状态检查
// @Tags         系统
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// aliRDSExportToS3Handler godoc
// @Summary      将阿里云RDS备份上传到S3
// @Description  获取指定环境的阿里云RDS最新备份并上传到AWS S3
// @Tags         阿里云RDS
// @Accept       json
// @Produce      json
// @Param        env  path      string  true  "环境名称"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /alirds/export/s3/{env} [post]
func aliRDSExportToS3Handler(c *gin.Context) {
	env := c.Param("env")

	// 获取实例配置
	instanceConfig, ok := configs.RDS.Aliyun.Instances[env]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid environment"})
		return
	}

	// 获取S3配置信息
	s3Config := configs.RDS.Aliyun.S3Export
	if s3Config.Region == "" || s3Config.BucketName == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "S3 configuration is missing"})
		return
	}

	// 获取备份下载链接
	backupURLs, err := getLastBackupURLs(instanceConfig.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get backup URLs",
			"details": err.Error(),
		})
		return
	}

	if backupURLs["BackupDownloadURL"] == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no backup found"})
		return
	}

	// 执行上传任务并等待完成
	result, err := uploadBackupToS3(
		backupURLs["BackupDownloadURL"],
		configs.RDS.Aliyun.S3Export.BucketName,
		configs.RDS.Aliyun.S3Export.Region,
		env,
		backupURLs["BackupStartTime"],
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to upload to S3",
			"details": err.Error(),
		})
		return
	}

	// 返回成功结果
	c.JSON(http.StatusOK, gin.H{
		"message":   "Backup upload completed",
		"s3_bucket": configs.RDS.Aliyun.S3Export.BucketName,
		"s3_key":    result.S3Key,
		"location":  result.Location,
		"region":    configs.RDS.Aliyun.S3Export.Region,
	})
}

// getS3ConfigHandler godoc
// @Summary      获取S3配置信息
// @Description  获取用于上传的AWS S3配置信息
// @Tags         配置
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /alirds/s3config [get]
func getS3ConfigHandler(c *gin.Context) {
	s3Config := configs.RDS.Aliyun.S3Export
	if s3Config.Region == "" || s3Config.BucketName == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "S3 configuration is missing",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"region":      s3Config.Region,
		"bucket_name": s3Config.BucketName,
	})
}

// getInstancesHandler godoc
// @Summary      获取所有实例配置
// @Description  获取阿里云和AWS的所有实例配置信息
// @Tags         配置
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /instances [get]
func getInstancesHandler(c *gin.Context) {
	instances := gin.H{
		"aliyun": make([]string, 0),
		"aws":    make([]string, 0),
	}

	// 获取阿里云实例列表
	for env := range configs.RDS.Aliyun.Instances {
		instances["aliyun"] = append(instances["aliyun"].([]string), env)
	}

	// 获取AWS实例列表
	for env := range configs.RDS.Aws.Instances {
		instances["aws"] = append(instances["aws"].([]string), env)
	}

	c.JSON(http.StatusOK, instances)
}
