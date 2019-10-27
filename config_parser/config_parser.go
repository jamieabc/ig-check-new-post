package config_parser

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Accounts []string

type Config struct {
	Accounts []string `yaml:"accounts"`
}

func Parse(fileName string) (Config, error) {
	data, err := ioutil.ReadFile(fileName)
	if nil != err {
		fmt.Printf("open file %s with error: %s\n", fileName, err)
		return Config{}, err
	}

	c := Config{}
	err = yaml.Unmarshal(data, &c)
	if nil != err {
		fmt.Printf("unmarshal yaml config with error: %s\n", err)
	}
	return c, nil
}
