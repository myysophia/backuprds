// main.go
package main

import (
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

	// 静态文件
	r.Static("/static", "./static")

	// Swagger
	r.GET("/doc/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API 路由
	r.GET("/alirds/:env", backupHandler)
	r.POST("/alirds/export/s3/:env", aliRDSExportToS3Handler)
	r.GET("/alirds/s3config", getS3ConfigHandler)
	r.GET("/awsrds/:env", awsBackupHandler)
	r.POST("/awsrds/export/:env", awsExportHandler)
	r.GET("/health", healthCheckHandler)
	r.GET("/instances", getInstancesHandler)

	// 前端路由
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	r.Run(":8080")
}
