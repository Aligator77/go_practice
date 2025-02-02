package config

import "os"

type Conf struct {
	Server struct {
		Port string
		Host string
	}
	AppVersion string
}

func New() (Conf, error) {
	serverConf := Conf{}
	serverConf.Server.Port = os.Getenv("SERVER_PORT")
	serverConf.Server.Host = os.Getenv("SERVER_HOST")
	serverConf.AppVersion = os.Getenv("APP_VERSION")
	return serverConf, nil
}

func GetAppVersion() string {
	return os.Getenv("APP_VERSION")
}
