package config

import (
	"github.com/spf13/viper"
	"zadanie-6105/logger"
)

type Config struct {
	ServerAddress    string `mapstructure:"SERVER_ADDRESS"`
	PostgresConn     string `mapstructure:"POSTGRES_CONN"`
	PostgresJdbcUrl  string `mapstructure:"POSTGRES_JDBC_URL"`
	PostgresUsername string `mapstructure:"POSTGRES_USERNAME"`
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresHost     string `mapstructure:"POSTGRES_HOST"`
	PostgresPort     string `mapstructure:"POSTGRES_PORT"`
	PostgresDatabase string `mapstructure:"POSTGRES_DATABASE"`
}

func InitializeConfig() (*Config, error) {
	logger.InitializeLogger()

	viper.AutomaticEnv()
	viper.SetDefault("SERVER_ADDRESS", "0.0.0.0:8080")
	viper.SetDefault("POSTGRES_CONN", "postgres://postgres:postgres@localhost:5432/postgres")
	viper.SetDefault("POSTGRES_JDBC_URL", "jdbc:postgresql://localhost:5432/postgres")
	viper.SetDefault("POSTGRES_USERNAME", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "postgres")
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("POSTGRES_DATABASE", "postgres")

	var config Config
	err := viper.Unmarshal(&config)

	return &config, err
}
