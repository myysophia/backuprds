// main.go
package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	// 加载配置
	loadConfig()

	// 初始化 Gin 路由
	r := gin.Default()

	// 配置静态文件路径
	r.Static("/static", "./static")

	// 定义 API 路由
	r.GET("/alirds/:env", backupHandler)
	r.GET("/awsrds/:env", awsBackupHandler)
	r.POST("/awsrds/export/:env", awsExportHandler) // 新增 AWS RDS 导出任务路由
	r.POST("/alirds/export/s3/:env", aliRDSExportToS3Handler) // 新增路由
	r.GET("/health", healthCheckHandler)

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
