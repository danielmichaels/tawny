package config

import (
	"log"
	"time"

	"github.com/joeshaw/envdecode"
)

type Conf struct {
	Db     dbConf
	Server serverConf
}

type dbConf struct {
	Host     string `env:"POSTGRES_HOST,default=localhost"`
	Db       string `env:"POSTGRES_DB,default=db"`
	User     string `env:"POSTGRES_USER,default=dbuser"`
	Password string `env:"POSTGRES_PASSWORD,default=dbuser"`
	// PG SSL MODES: allow, disable
	SSLMode  string `env:"POSTGRES_SSL_MODE,default=disable"`
	Port     int    `env:"POSTGRES_PORT,default=5433"`
	MaxConns int    `env:"POSTGRES_MAX_CONNS,default=16"`
	//DatabaseConnectionContext time.Duration `env:"DATABASE_CONNECTION_CONTEXT,default=15s"`
}

type serverConf struct {
	APIPort      int           `env:"API_SERVER_PORT,default=9090"`
	WebPort      int           `env:"WEB_SERVER_PORT,default=9091"`
	TimeoutRead  time.Duration `env:"SERVER_TIMEOUT_READ,default=5s"`
	TimeoutIdle  time.Duration `env:"SERVER_TIMEOUT_IDLE,default=5s"`
	TimeoutWrite time.Duration `env:"SERVER_TIMEOUT_WRITE,default=5s"`
}

// AppConfig Setup and install the applications' configuration environment variables
func AppConfig() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}
	return &c
}
