package config

import (
	"backuprds/internal/logger"
	"github.com/spf13/viper"
)

type Config struct {
	RDS struct {
		Aliyun struct {
			Instances map[string]InstanceConfig `yaml:"instances"`
			S3Export  struct {
				Region     string `yaml:"region"`
				BucketName string `yaml:"bucketname"`
			} `yaml:"s3export"`
		} `yaml:"aliyun"`
		Aws struct {
			Instances  map[string]InstanceConfig `yaml:"instances"`
			ExportTask struct {
				S3Prefix                   string `yaml:"s3prefix"`
				IamRoleArn                 string `yaml:"iam_role_arn"`
				ExportTaskIdentifierPrefix string `yaml:"exportTaskIdentifierPrefix"`
			} `yaml:"exporttask"`
		} `yaml:"aws"`
	} `yaml:"rds"`
}

type InstanceConfig struct {
	ID           string `yaml:"id"`
	Region       string `yaml:"region"`
	KmsKeyId     string `yaml:"kmsKeyId"`
	S3BucketName string `yaml:"s3BucketName"`
}

var Cfg Config

func LoadConfig() {
	logger.LogInfo("Loading configuration")

	viper.BindEnv("rds.aliyun.s3export.region")
	viper.BindEnv("rds.aliyun.s3export.bucketname")

	if err := viper.ReadInConfig(); err != nil {
		logger.LogFatal("Failed to read config file",
			logger.Error(err))
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		logger.LogFatal("Failed to unmarshal config",
			logger.Error(err))
	}

	logger.LogInfo("Configuration loaded successfully",
		logger.String("aliyun_region", Cfg.RDS.Aliyun.S3Export.Region),
		logger.String("aliyun_bucket", Cfg.RDS.Aliyun.S3Export.BucketName))
}

func GetConfig() *Config {
	return &Cfg
}
