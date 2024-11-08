// config.go
package main

import (
	"log"

	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	RDS struct {
		AccessKey    string            `yaml:"access_key"`
		AccessSecret string            `yaml:"access_secret"`
		Region       string            `yaml:"region"`
		Instances    map[string]string `yaml:"instances"`
	} `yaml:"rds"`
}

var config Config

func loadConfig() {
	data, err := ioutil.ReadFile("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}
}
