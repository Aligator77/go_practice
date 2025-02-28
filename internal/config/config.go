package config

import (
	"github.com/caarlos0/env/v11"
)

type Conf struct {
	Server struct {
		Port    string `env:"SERVER_PORT" envDefault:"8080"`
		Host    string `env:"SERVER_HOST" envDefault:"localhost"`
		Address string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	}
	AppVersion     string `env:"APP_VERSION" envDefault:"0.0.1"`
	BaseURL        string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	DisableDBStore string `env:"DISABLE_DB_STORE" envDefault:"1"`
	LocalStore     string `env:"FILE_STORAGE_PATH" envDefault:"/tmp/short-url-db.json"`
	DB             struct {
		Host       string `env:"DB_HOST" envDefault:"localhost"`
		Port       string `env:"DB_PORT" envDefault:"5432"`
		User       string `env:"DB_USER" envDefault:"yapr"`
		Password   string `env:"DB_PASSWORD" envDefault:"yapr"`
		Name       string `env:"DB_NAME" envDefault:"yapr"`
		MaxOpenCon int    `env:"DB_MAX_OPEN_CON" envDefault:"10"`
		MaxIdleCon int    `env:"DB_MAX_IDLE_CON" envDefault:"10"`
	}
}

func New() (Conf, error) {
	serverConf := Conf{}
	err := env.Parse(&serverConf)
	if err != nil {
		return serverConf, err
	}
	return serverConf, nil
}
