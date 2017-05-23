package main

import (
	"io/ioutil"
	"time"

	"github.com/Sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ClientID      string        `json:"client_id" yaml:"client_id"`
	ClientSecret  string        `json:"client_secret" yaml:"client_secret"`
	RedirectURL   string        `json:"redirect_url" yaml:"redirect_url"`
	StateString   string        `json:"state_string" yaml:"state_string"`
	RequestSleep  time.Duration `json:"request_sleep" yaml:"request_sleep"`
	LoopSleep     time.Duration `json:"loop_sleep" yaml:"loop_sleep"`
	ListenAddress string        `json:"listen_address" yaml:"listen_address"`
	LogLevel      string        `json:"log_level" yaml:"log_level"`
	DatabasePath  string        `json:"database_path" yaml:"database_path"`
}

func ConfigFromFile(file string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var c *Config
	if err := yaml.Unmarshal([]byte(configBytes), &c); err != nil {
		return nil, err
	}
	if len(c.LogLevel) != 0 {
		level, err := logrus.ParseLevel(c.LogLevel)
		if err != nil {
			return nil, err
		}
		logrus.SetLevel(level)
	}
	logrus.Debugf("%#v", c)
	return c, nil
}
