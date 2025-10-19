package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string        `yaml:"env" env-required:"true"`
	AppSecret  string        `yaml:"AppSecret" env-required:"true" env:"APP_SECRET"`
	HTTPServer `yaml:"http_server"`
	Clients    ClientsConfig `yaml:"clients"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Client struct {
	Addr         string        `yaml:"addr"`
	Timeout      time.Duration `yaml:"timeout"`
	RetriesCount int           `yaml:"retriesCount"`
}

type ClientsConfig struct {
	SSO      Client `yaml:"sso"`
	Payments Client `yaml:"payments"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(configPath string) *Config {

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}
	return res
}
