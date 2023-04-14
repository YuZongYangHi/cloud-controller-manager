package config

import "github.com/YuZongYangHi/cloud-controller-manager/pkg/util/parsers"

type LoadBalanceConfig struct {
	LoadBalanceSet LoadBalanceSetConfig `yaml:"loadbalance"`
	Region         string               `yaml:"region"`
}

type CloudProviderConfig struct {
	HTTP CloudProviderHTTPConfig `yaml:"http"`
	DB   CloudProviderDBConfig   `yaml:"db"`
}

type CloudProviderHTTPConfig struct {
	Host string `yaml:"host"`
	Port int64  `yaml:"port"`
}

type CloudProviderDBConfig struct {
	User     string `yaml:"user"`
	Host     string `yaml:"host"`
	Port     int64  `yaml:"port"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type LoadBalanceSetConfig struct {
	Bind     string `yaml:"bind"`
	Released string `yaml:"released"`
	List     string `yaml:"list"`
}

func NewLoadBalanceConfig(in string) (*LoadBalanceConfig, error) {
	var config *LoadBalanceConfig
	err := parsers.ParserConfigurationByFile(parsers.YAML, in, &config)
	return config, err
}

func NewCloudProviderHTTPConfig(in string) (*CloudProviderConfig, error) {
	var config *CloudProviderConfig
	err := parsers.ParserConfigurationByFile(parsers.YAML, in, &config)
	return config, err
}
