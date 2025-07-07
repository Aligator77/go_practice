// Package config parse and create configurations for all project
package config

import (
	"flag"

	"github.com/caarlos0/env/v11"

	"github.com/Aligator77/go_practice/internal/helpers"
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
		MaxOpenCon int    `env:"DB_MAX_OPEN_CON" envDefault:"30"`
		MaxIdleCon int    `env:"DB_MAX_IDLE_CON" envDefault:"30"`
		DSN        string `env:"DATABASE_DSN"`
	}
}

func New() (Conf, error) {
	serverConf := Conf{}
	err := env.Parse(&serverConf)
	if err != nil {
		return serverConf, err
	}

	serverAddrFlag := flag.String("a", "", "input server address")
	baseURLFlag := flag.String("b", "", "input server address")
	localStoreFile := flag.String("f", "", "input server address")
	dbDsn := flag.String("d", "", "input db dsn address")
	flag.Parse()

	if len(*serverAddrFlag) > 0 && helpers.CheckFlag(serverAddrFlag) {
		serverConf.Server.Address = *serverAddrFlag
	}

	if len(*baseURLFlag) > 0 && helpers.CheckFlagHTTP(baseURLFlag) {
		serverConf.BaseURL = *baseURLFlag
	}

	if len(*localStoreFile) > 0 {
		serverConf.LocalStore = *localStoreFile
	}
	if len(*dbDsn) > 0 {
		serverConf.DB.DSN = *dbDsn
	}
	if len(serverConf.DB.DSN) > 0 {
		serverConf.DisableDBStore = "0"
	}

	return serverConf, nil
}
