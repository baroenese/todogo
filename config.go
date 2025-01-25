package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

/* Environment utility */

func loadEnvStr(key string, result *string) {
	s, ok := os.LookupEnv(key)
	if !ok {
		return
	}
	*result = s
}

func loadEnvUint(key string, result *uint) {
	s, ok := os.LookupEnv(key)
	if !ok {
		return
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return
	}
	*result = uint(n) // will clamp the negative value
}

/* Configuration */

type pgConfig struct {
	Host string `yaml:"host" json:"host"`
	Port uint   `yaml:"port" json:"port"`

	DBName  string `yaml:"db_name" json:"db_name"`
	SslMode string `yaml:"ssl_mode" json:"ssl_mode"`
}

func (p pgConfig) ConnStr() string {
	return fmt.Sprintf("postgres://postgres:example@%s:%d/%s", p.Host, p.Port, p.DBName)
}

func (p pgConfig) LoadFromEnv() {
	loadEnvStr("KAD_DB_HOST", &p.Host)
	loadEnvUint("KAD_DB_PORT", &p.Port)
	loadEnvStr("KAD_DB_NAME", &p.DBName)
	loadEnvStr("KAD_DB_SSL", &p.SslMode)
}

type listenConfig struct {
	Host string `yaml:"host" json:"host"`
	Port uint   `yaml:"port" json:"port"`
}

func (l listenConfig) Addr() string {
	return fmt.Sprintf("%s:%d", l.Host, l.Port)
}

func (l *listenConfig) loadFromEnv() {
	loadEnvStr("KAD_LISTEN_HOST", &l.Host)
	loadEnvUint("KAD_LISTEN_PORT", &l.Port)
}

func defaultListenConfig() listenConfig {
	return listenConfig{
		Host: "127.0.0.1",
		Port: 8080,
	}
}

type config struct {
	Listen   listenConfig `yaml:"listen" json:"listen"`
	DBConfig pgConfig     `yaml:"db" json:"db"`
}

func (c *config) loadFromEnv() {
	c.Listen.loadFromEnv()
}

func defaultConfig() config {
	return config{
		Listen: defaultListenConfig(),
	}
}

func loadConfigFromReader(r io.Reader, c *config) error {
	return yaml.NewDecoder(r).Decode(c)
}

func loadConfigFromFile(fn string, c *config) error {
	_, err := os.Stat(fn)
	if err != nil {
		return err
	}
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	return loadConfigFromReader(f, c)
}
