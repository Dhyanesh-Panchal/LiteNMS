package config

import (
	"fmt"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string
}

func NewConfig() *Config {
	return &Config{
		DBUser:     "nms_backend",
		DBPassword: "litenms",
		DBName:     "config_db",
		DBHost:     "localhost",
		DBPort:     "5432",
	}
}

func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
} 