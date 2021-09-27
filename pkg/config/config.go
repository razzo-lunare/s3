package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

type S3 struct {
	BucketName  string `yaml:"bucket-name"`
	Endpoint    string `yaml:"endpoint"`
	AccessKeyID string `yaml:"access-key-id"`
	AccessKey   string `yaml:"secret-access-key"`
}

// S3Config holds the data read from the s3.yml
type S3Config struct {
	S3 []*S3 `yaml:"s3"`
}

func (s *S3Config) GetBucketCreds(bucketName string) (*S3, error) {
	for _, s3 := range s.S3 {
		if s3.BucketName == bucketName {
			return s3, nil
		}
	}
	return nil, fmt.Errorf("No bucket creds found in the config for bucket: %s", bucketName)
}

// NewConfig returns a new config
func NewConfig(pathToConfig string) (*S3Config, error) {
	var s3Config *S3Config = &S3Config{}
	jsonFile, err := os.Open(pathToConfig)
	if err != nil {
		return nil, fmt.Errorf("Opening config %s", err.Error())
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("Reading config %s", err.Error())
	}

	klog.V(10).Infof("Config File Content: %s", string(byteValue))

	err = yaml.Unmarshal(byteValue, s3Config)
	if err != nil {
		return nil, fmt.Errorf("Un-marshalling config: %s", err.Error())
	}

	klog.V(10).Infof("Config Object: %+v", s3Config)

	return s3Config, nil
}
