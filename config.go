package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Database      Database `yaml:"database"`
	Authorization string   `yaml:"authorization"`
	Log           Log      `yaml:"log"`
	Ceph          Ceph     `yaml:"ceph"`
}

type Database struct {
	Dialect  string `yaml:"dialect"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Address  string `yaml:"address"`
	Name     string `yaml:"name"`
}

type Log struct {
	Filename string `yaml:"filename"`
}

type Ceph struct {
	User     string `yaml:"user"`
	Monitors string `yaml:"monitors"`
	Keyring  string `yaml:"keyring"`
}

func readConfig(filename string) (config *Config, err error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	config = &Config{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return
	}
	return
}
