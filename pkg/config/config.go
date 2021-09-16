package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

// S3Config holds the data read from the s3.yml
type S3Config struct {
	DigitalOceanS3StockDataBucketName string `yaml:"digital_ocean_s3_stock_data_bucket_name"`
	DigitalOceanS3Endpoint            string `yaml:"digital_ocean_s3_endpoint"`
	DigitalOceanS3AccessKeyID         string `yaml:"digital_ocean_s3_access_key_id"`
	DigitalOceanS3SecretAccessKey     string `yaml:"digital_ocean_s3_secret_access_key"`
	DigitalOceanS3CreateBucket        bool   `yaml:"digital_ocean_s3_create_bucket"`
}

// NewConfig returns a new config
func NewConfig(pathToConfig string) (*S3Config, error) {
	var fortunaConfig *S3Config = &S3Config{}
	jsonFile, err := os.Open(pathToConfig)
	if err != nil {
		return nil, fmt.Errorf("Opening config %s", err.Error())
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("Reading config %s", err.Error())
	}

	klog.V(10).Infof("Config File Content: %s", string(byteValue))
	klog.V(10).Infof("Config Object: %+v", fortunaConfig)

	err = yaml.Unmarshal(byteValue, fortunaConfig)
	if err != nil {
		return nil, fmt.Errorf("Unmarshalling config: %s", err.Error())
	}

	return fortunaConfig, nil
}
