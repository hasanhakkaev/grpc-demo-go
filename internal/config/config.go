package conf

import (
	_ "embed"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"strings"
)

func Read() (*Configuration, error) {
	// Environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Use environment variable `CONFIG_ENV` to determine which environment-specific config to load
	env := viper.GetString("CONFIG_ENV")
	if env == "" {
		env = "dev" // Default to 'dev' if not specified
	}

	// Configuration file setup
	viper.SetConfigType("yaml")
	viper.SetConfigName("configuration") // Base configuration file
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configuration")
	viper.AddConfigPath("/etc/yqapp-demo/") // Optionally look in a system-wide path

	// Read the base configuration
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading base configuration: %w", err)
	}

	// Load environment-specific configuration (e.g., configuration.dev.yaml)
	viper.SetConfigName(fmt.Sprintf("configuration.%s", env))
	if err := viper.MergeInConfig(); err != nil {
		fmt.Printf("No environment-specific configuration found for '%s', continuing with base config.\n", env)
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
	Endpoint    string `env:"ENDPOINT" envDefault:"/metrics" yaml:"endpoint"`
	Environment string `env:"ENVIRONMENT" envDefault:"development" yaml:"environment"`
}

type Logger struct {
	Enabled     bool   `env:"ENABLED" envDefault:"true" yaml:"enabled"`
	Environment string `env:"ENVIRONMENT" envDefault:"development" yaml:"environment"`
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
	MessageConsumptionRate uint   `env:"MESSAGE_CONSUMPTION_RATE" envDefault:"1000" yaml:"messageConsumptionRate"`
	LogLevel               string `env:"LOG_LEVEL" envDefault:"info" yaml:"logLevel"`
	LogEncoding            string `env:"LOG_ENCODING" yaml:"logEncoding"`
	MetricsPort            uint16 `env:"METRICS_PORT" envDefault:"5000" yaml:"metricsPort"`
	ProfilingPort          uint16 `env:"PROFILING_PORT" envDefault:"8080" yaml:"profilingPort"`
}

type Producer struct {
	MessageProductionRate uint   `env:"MESSAGE_PRODUCTION_RATE" envDefault:"1000/s" yaml:"messageProductionRate"`
	MaxBacklog            uint   `env:"MAX_BACKLOG" envDefault:"10" yaml:"maxBacklog"`
	LogLevel              string `env:"LOG_LEVEL" envDefault:"info" yaml:"logLevel"`
	LogEncoding           string `env:"LOG_ENCODING" yaml:"logEncoding"`
	MetricsPort           uint16 `env:"METRICS_PORT" envDefault:"5000" yaml:"metricsPort"`
	ProfilingPort         uint16 `env:"PROFILING_PORT" envDefault:"8080" yaml:"profilingPort"`
}

type Server struct {
	Name        string `env:"NAME" envDefault:"yqapp-demo-server" yaml:"name"`
	Environment string `env:"ENVIRONMENT" envDefault:"development" yaml:"environment"`
	Host        string `env:"HOST" envDefault:"localhost" yaml:"host"`
	Port        uint16 `env:"PORT" envDefault:"8080" yaml:"port"`
}

type Client struct {
	Name        string `env:"NAME" envDefault:"yqapp-demo-client" yaml:"name"`
	Environment string `env:"ENVIRONMENT" envDefault:"development" yaml:"environment"`
}
type Configuration struct {
	Server          Server   `env:"SERVER" yaml:"server"`
	ConsumerService Consumer `env:"CONSUMER" yaml:"consumer"`
	ProducerService Producer `env:"PRODUCER" yaml:"producer"`
	Database        Database `envPrefix:"DATABASE_" yaml:"database"`
	Metrics         Metrics  `envPrefix:"METRICS_" yaml:"metrics"`
	Logger          Logger   `envPrefix:"LOGGER_" yaml:"logger"`
	Client          Client   `envPrefix:"CLIENT" yaml:"client"`
}

func (c Configuration) GetProducerMetricsPort() string {
	return fmt.Sprintf("%d", c.ProducerService.MetricsPort)
}

func (c Configuration) GetProducerProfilingPort() string {
	return fmt.Sprintf("%d", c.ProducerService.ProfilingPort)
}

func (c Configuration) GetProducerLogLevel() string {
	return c.ProducerService.LogLevel
}

func (c Configuration) GetProducerLogEncoding() string {
	return c.ProducerService.LogEncoding
}

func (c Configuration) GetConsumerMetricsPort() string {
	return fmt.Sprintf("%d", c.ConsumerService.MetricsPort)
}

func (c Configuration) GetConsumerProfilingPort() string {
	return fmt.Sprintf("%d", c.ConsumerService.ProfilingPort)
}

func (c Configuration) GetConsumerLogLevel() string {
	return c.ConsumerService.LogLevel
}

func (c Configuration) GetConsumerLogEncoding() string {
	return c.ConsumerService.LogEncoding
}

// Address returns the configuration needed to initialize a net.Listener instance.
func (s Server) Address() (network string, address string) {
	return "tcp", fmt.Sprintf(":%d", s.Port)
}

func (s Server) URI() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
