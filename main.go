// main.go
package main

import (
	"log"

	_ "backuprds/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title        Nova RDS 跨云灾备系统 API
// @version      1.0
// @description  用于管理阿里云和AWS RDS备份的API系统
// @BasePath     /

func main() {
	loadConfig()

	r := gin.Default()

	// 添加swagger路由
	r.GET("/doc/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 配置静态文件路径
	r.Static("/static", "./static")

	// 定义 API 路由
	r.GET("/alirds/:env", backupHandler)
	r.GET("/awsrds/:env", awsBackupHandler)
	r.POST("/awsrds/export/:env", awsExportHandler)           // 新增 AWS RDS 导出任务路由
	r.POST("/alirds/export/s3/:env", aliRDSExportToS3Handler) // 新增路由
	r.GET("/health", healthCheckHandler)
	r.GET("/alirds/s3config", getS3ConfigHandler)

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
