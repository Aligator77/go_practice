package config

import (
	"os"
	"strconv"
)

type Conf struct {
	Server struct {
		Port string
		Host string
	}
	AppVersion string
	SiteHost   string
	DB         struct {
		Host       string
		Port       string
		User       string
		Password   string
		Name       string
		MaxOpenCon int
		MaxIdleCon int
	}
}

func New() (Conf, error) {
	serverConf := Conf{}
	serverConf.Server.Port = os.Getenv("SERVER_PORT")
	serverConf.Server.Host = os.Getenv("SERVER_HOST")
	serverConf.AppVersion = os.Getenv("APP_VERSION")
	serverConf.SiteHost = os.Getenv("SITE_HOST")
	serverConf.DB.Host = os.Getenv("DB_HOST")
	serverConf.DB.Port = os.Getenv("DB_PORT")
	serverConf.DB.User = os.Getenv("DB_USER")
	serverConf.DB.Password = os.Getenv("DB_PASSWORD")
	serverConf.DB.Name = os.Getenv("DB_NAME")
	serverConf.DB.MaxIdleCon, _ = strconv.Atoi(os.Getenv("DB_MAX_OPEN_CON"))
	serverConf.DB.MaxOpenCon, _ = strconv.Atoi(os.Getenv("DB_MAX_IDLE_CON"))

	return serverConf, nil
}

func GetAppVersion() string {
	return os.Getenv("APP_VERSION")
}
