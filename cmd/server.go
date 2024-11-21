package cmd

import (
	"backuprds/internal/config"
	"backuprds/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func runServer(cmd *cobra.Command, args []string) {
	config.LoadConfig()

	r := gin.Default()

	// 静态文件
	r.Static("/static", "./static")

	// Swagger
	r.GET("/doc/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API 路由
	r.GET("/alirds/:env", handlers.BackupHandler)
	r.POST("/alirds/export/s3/:env", handlers.AliRDSExportToS3Handler)
	r.GET("/alirds/s3config", handlers.GetS3ConfigHandler)
	r.GET("/awsrds/:env", handlers.AwsBackupHandler)
	r.POST("/awsrds/export/:env", handlers.AwsExportHandler)
	r.GET("/health", handlers.HealthCheckHandler)
	r.GET("/instances", handlers.GetInstancesHandler)

	// 前端路由
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	r.Run(":" + port)
}
