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
			Instances    map[string]InstanceConfig `yaml:"instances"`
			S3Export     struct {
				Region     string `yaml:"region"`
				BucketName string `yaml:"bucket_name"`
			} `yaml:"s3_export"`
		} `yaml:"aliyun"`
		Aws struct {
			Instances   map[string]InstanceConfig `yaml:"instances"`
			ExportTask  struct {
				S3Prefix      string `yaml:"s3_prefix"`
				IamRoleArn    string `yaml:"iam_role_arn"`
			} `yaml:"export_task"`
		} `yaml:"aws"`
	} `yaml:"rds"`
}

type InstanceConfig struct {
	ID            string `yaml:"id"`
	Region        string `yaml:"region"`
	KmsKeyId      string `yaml:"kms_key_id"`
	S3BucketName  string `yaml:"s3_bucket_name"`
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
