package config

import (
	"log"

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
	viper.BindEnv("rds.aliyun.s3export.region")
	viper.BindEnv("rds.aliyun.s3export.bucketname")

	viper.SetDefault("rds.aliyun.s3export.region", "ap-southeast-2")
	viper.SetDefault("rds.aliyun.s3export.bucketname", "alirds-backup")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}

	if Cfg.RDS.Aliyun.S3Export.Region == "" || Cfg.RDS.Aliyun.S3Export.BucketName == "" {
		log.Printf("Warning: S3 export configuration is incomplete - Region: %q, BucketName: %q",
			Cfg.RDS.Aliyun.S3Export.Region,
			Cfg.RDS.Aliyun.S3Export.BucketName)
	}
}

func GetConfig() *Config {
	return &Cfg
}
