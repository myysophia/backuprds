// config.go
package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	RDS struct {
		Aliyun struct {
			AccessKey    string            `yaml:"access_key"`
			AccessSecret string            `yaml:"access_secret"`
			Region       string            `yaml:"region"`
			Instances    map[string]string `yaml:"instances"`
		} `yaml:"aliyun"`
		Aws struct {
			AccessKey  string            `yaml:"access_key"`
			SecretKey  string            `yaml:"secret_key"`
			Region     string            `yaml:"region"`
			Instances  map[string]string `yaml:"instances"`
			ExportTask struct {
				KmsKeyId                   string `yaml:"kms_key_id"`
				S3BucketName               string `yaml:"s3_bucket_name"`
				IamRoleArn                 string `yaml:"iam_role_arn"`
				ExportTaskIdentifierPrefix string `yaml:"export_task_identifier_prefix"`
			} `yaml:"export_task"`
		} `yaml:"aws"`
	} `yaml:"rds"`
}

var configs Config

func loadConfig() {
	data, err := ioutil.ReadFile("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	err = yaml.Unmarshal(data, &configs)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}
}
