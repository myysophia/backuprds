package cmd

import (
	"backuprds/internal/logger"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	port      string
	logLevel  string
	logConfig string
	rootCmd   = &cobra.Command{
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
	rootCmd.PersistentFlags().StringVar(&logLevel, "log.level", "", "日志级别")
	rootCmd.PersistentFlags().StringVar(&logConfig, "log.config", "config/logger.yaml", "日志配置文件路径")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		logger.LogError("Failed to read config file",
			logger.Error(err),
			logger.String("config_file", cfgFile))
	}
}
