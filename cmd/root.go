package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	port    string
	rootCmd = &cobra.Command{
		Use:   "backuprds",
		Short: "Nova RDS 跨云灾备系统",
		Long:  `Nova RDS 跨云灾备系统支持阿里云和AWS RDS的备份管理`,
		Run:   runServer,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config/config.yaml", "配置文件路径")
	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "Web服务端口号")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("无法读取配置文件: %s\n", err)
	}
}
