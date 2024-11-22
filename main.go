package main

import (
	"backuprds/cmd"
	_ "backuprds/docs"
	"backuprds/internal/logger"
	"fmt"
	"os"
)

// @title        Nova RDS 跨云灾备系统 API
// @version      1.0
// @description  用于管理阿里云和AWS RDS备份的API系统
// @BasePath     /

func main() {
	// 初始化日志
	if err := logger.InitFromFile("config/logger.yaml"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// 测试日志
	logger.LogInfo("Application starting",
		logger.String("app", "backuprds"),
		logger.String("version", "1.0"))

	if err := cmd.Execute(); err != nil {
		logger.LogError("Failed to execute command", logger.Error(err))
		os.Exit(1)
	}
}
