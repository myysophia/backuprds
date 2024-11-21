package main

import (
	"backuprds/cmd"
	_ "backuprds/docs"
	"fmt"
	"os"
)

// @title        Nova RDS 跨云灾备系统 API
// @version      1.0
// @description  用于管理阿里云和AWS RDS备份的API系统
// @BasePath     /

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
