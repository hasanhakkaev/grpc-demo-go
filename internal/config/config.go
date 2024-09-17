package conf

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

//go:embed config.yaml
var defaultConfiguration []byte

func Read() (*Configuration, error) {
	// Environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Configuration file
	viper.SetConfigType("yaml")

	// Read configuration
	if err := viper.ReadConfig(bytes.NewBuffer(defaultConfiguration)); err != nil {
		return nil, err
	}

	// Unmarshal the configuration
	var config Configuration
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

type Metrics struct {
	Name        string `env:"NAME" envDefault:"metrics-collector" yaml:"name" `
	Enabled     bool   `env:"ENABLED" envDefault:"true" yaml:"enabled"`
	Host        string `env:"HOST" envDefault:"localhost" yaml:"host"`
	Port        uint16 `env:"PORT" envDefault:"4333" yaml:"port"`
	Environment string `env:"ENVIRONMENT" envDefault:"development" yaml:"environment"`
}

type Tracing struct {
	Name        string `env:"NAME" envDefault:"traces-collector" yaml:"name"`
	Enabled     bool   `env:"ENABLED" envDefault:"true" yaml:"enabled"`
	Host        string `env:"HOST" envDefault:"localhost" yaml:"host"`
	Port        uint16 `env:"PORT" envDefault:"4444" yaml:"port"`
	Environment string `env:"ENVIRONMENT" envDefault:"development" yaml:"environment"`
}

type Logger struct {
	Name        string `env:"NAME" envDefault:"development" yaml:"name"`
	Enabled     bool   `env:"ENABLED" envDefault:"true" yaml:"enabled"`
	Level       string `env:"LEVEL" envDefault:"info" yaml:"level"`
	Environment string `env:"ENVIRONMENT" envDefault:"development" yaml:"environment"`
	Output      string `env:"OUTPUT" envDefault:"stdout" yaml:"output"`
}
type Database struct {
	Host     string `env:"HOST" envDefault:"localhost" yaml:"host"`
	Port     uint16 `env:"PORT" envDefault:"5432" yaml:"port"`
	Username string `env:"USERNAME" envDefault:"postgres" yaml:"username"`
	Password string `env:"PASSWORD" envDefault:"postgres" yaml:"password"`
	Database string `env:"DATABASE" envDefault:"postgres" yaml:"database"`
}

type Consumer struct {
	Host        string `env:"HOST" envDefault:"localhost" yaml:"host"`
	Port        uint16 `env:"PORT" envDefault:"5432" yaml:"port"`
	Environment string `env:"ENVIRONMENT" envDefault:"development" yaml:"environment"`
	RateLimit   uint   `env:"RATE_LIMIT" envDefault:"1000" yaml:"rateLimit"`
}

type Producer struct {
	Host        string `env:"HOST" envDefault:"localhost" yaml:"host"`
	Port        uint16 `env:"PORT" envDefault:"5432" yaml:"port"`
	Environment string `env:"ENVIRONMENT" envDefault:"development" yaml:"environment"`
}

type Configuration struct {
	ConsumerServer Consumer `env:"CONSUMER" yaml:"consumer"`
	ProducerServer Producer `env:"PRODUCER" yaml:"producer"`
	DB             Database `envPrefix:"DATABASE_" yaml:"DB"`
	Metrics        Metrics  `envPrefix:"METRICS_" yaml:"metrics"`
	Tracing        Tracing  `envPrefix:"TRACING_" yaml:"tracing"`
	Logger         Logger   `envPrefix:"LOGGER_" yaml:"logger"`
}

func (p Producer) Address() string {
	return fmt.Sprintf("%s:%d", p.Host, p.Port)
}

func (c Consumer) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (t Tracing) Address() string {
	return fmt.Sprintf("%s:%d", t.Host, t.Port)
}

func (m Metrics) Address() string {
	return fmt.Sprintf("%s:%d", m.Host, m.Port)
}
