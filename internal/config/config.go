package conf

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/hasanhakkaev/yqapp-demo/assets"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"strings"
)

func Read() (*Configuration, error) {
	// Environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Configuration file
	viper.SetConfigType("yaml")

	// Read configuration
	configuration, err := assets.EmbeddedFiles.ReadFile("configuration.yaml")
	if err != nil {
		return nil, err
	}

	if err := viper.ReadConfig(bytes.NewBuffer(configuration)); err != nil {
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
	Enabled     bool   `env:"ENABLED" envDefault:"true" yaml:"enabled"`
	Host        string `env:"HOST" envDefault:"localhost" yaml:"host"`
	Port        uint16 `env:"PORT" envDefault:"4333" yaml:"port"`
	Endpoint    string `env:"ENDPOINT" envDefault:"/metrics" yaml:"endpoint"`
	Environment string `env:"ENVIRONMENT" envDefault:"development" yaml:"environment"`
}

type Logger struct {
	Enabled     bool   `env:"ENABLED" envDefault:"true" yaml:"enabled"`
	Level       string `env:"LOG_LEVEL" envDefault:"info" yaml:"level"`
	Environment string `env:"ENVIRONMENT" envDefault:"development" yaml:"environment"`
	Encoding    string `env:"OUTPUT" envDefault:"console" yaml:"output"`
}
type Database struct {
	Host     string `env:"HOST" envDefault:"localhost" yaml:"host"`
	Port     uint16 `env:"PORT" envDefault:"5432" yaml:"port"`
	Engine   string `env:"ENGINE" envDefault:"postgres" yaml:"engine"`
	Username string `env:"USERNAME" envDefault:"postgres" yaml:"username"`
	Password string `env:"PASSWORD" envDefault:"postgres" yaml:"password"`
	Database string `env:"DATABASE" envDefault:"postgres" yaml:"database"`
}

type Consumer struct {
	MessageConsumptionRate uint `env:"MESSAGE_CONSUMPTION_RATE" envDefault:"1000" yaml:"messageConsumptionRate"`
}

type Producer struct {
	MessageProductionRate uint `env:"MESSAGE_PRODUCTION_RATE" envDefault:"1000/s" yaml:"messageProductionRate"`
	MaxBacklog            uint `env:"MAX_BACKLOG" envDefault:"10" yaml:"maxBacklog"`
}

type Server struct {
	Name        string `env:"NAME" envDefault:"app" yaml:"name"`
	Environment string `env:"ENVIRONMENT" envDefault:"development" yaml:"environment"`
	Host        string `env:"HOST" envDefault:"localhost" yaml:"host"`
	Port        uint16 `env:"PORT" envDefault:"8080" yaml:"port"`
}
type Configuration struct {
	Server          Server   `yaml:"server" yaml:"server"`
	ConsumerService Consumer `env:"CONSUMER" yaml:"consumer"`
	ProducerService Producer `env:"PRODUCER" yaml:"producer"`
	Database        Database `envPrefix:"DATABASE_" yaml:"database"`
	Metrics         Metrics  `envPrefix:"METRICS_" yaml:"metrics"`
	Logger          Logger   `envPrefix:"LOGGER_" yaml:"logger"`
}

func (m Metrics) Address() string {
	return fmt.Sprintf("%s:%d", m.Host, m.Port)
}

// Address returns the configuration needed to initialize a net.Listener instance.
func (s Server) Address() (network string, address string) {
	return "tcp", fmt.Sprintf(":%d", s.Port)
}

func (s Server) URI() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
