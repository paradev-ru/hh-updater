package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ClientID               string        `json:"client_id" yaml:"client_id"`
	ClientSecret           string        `json:"client_secret" yaml:"client_secret"`
	PublicURLRaw           string        `json:"public_url" yaml:"public_url"`
	RedirectURL            string        `json:"redirect_url" yaml:"redirect_url"`
	StateString            string        `json:"state_string" yaml:"state_string"`
	LoopSleep              time.Duration `json:"loop_sleep" yaml:"loop_sleep"`
	ListenAddress          string        `json:"listen_address" yaml:"listen_address"`
	LogLevel               string        `json:"log_level" yaml:"log_level"`
	DatabasePath           string        `json:"database_path" yaml:"database_path"`
	PublicURL              *url.URL      `json:"-" yaml:"-"`
	CookieName             string        `json:"cookie_name" yaml:"cookie_name"`
	CookieHostname         string        `json:"-" yaml:"-"`
	CookieEncryptionKey    string        `json:"cookie_encryption_key" yaml:"cookie_encryption_key"`
	CookieEncryptionCipher cipher.Block  `json:"-" yaml:"-"`
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
	url, err := url.Parse(c.PublicURLRaw)
	if err != nil {
		return nil, err
	}
	c.PublicURL = url
	c.CookieHostname = domainFromHost(c.PublicURL.Hostname())
	cipherBlock, err := aes.NewCipher([]byte(c.CookieEncryptionKey))
	if err != nil {
		return nil, err
	}
	c.CookieEncryptionCipher = cipherBlock
	return c, nil
}

func domainFromHost(host string) string {
	index := strings.Index(host, ":")
	if index > 0 {
		return host[:index]
	}
	return host
}

func (c *Config) String() string {
	data, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(data)
}
