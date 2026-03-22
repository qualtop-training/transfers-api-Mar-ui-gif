package config

import (
	"encoding/json"
	"time"
	"transfers-api/internal/logging"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Business       BusinessConfig `json:"business"`
	MongoDBConfig  MongoDB        `json:"mongodb"`
	MySQLConfig    MySQL          `json:"mysql"`
	CCacheConfig   CCache         `json:"ccache"`
	RabbitMQConfig RabbitMQ       `json:"rabbitmq"`
}

type BusinessConfig struct {
	TransferMinAmount int `env:"TRANSFER_MIN_AMOUNT" envDefault:"1" json:"transfer_min_amount"`
}

type MongoDB struct {
	ConnectTimeout time.Duration `env:"MONGODB_CONNECT_TIMEOUT" envDefault:"10s" json:"connect_timeout"`
	Hostname       string        `env:"MONGODB_HOSTNAME" envDefault:"mongodb" json:"hostname"`
	Port           int           `env:"MONGODB_PORT" envDefault:"27017" json:"port"`
	Username       string        `env:"MONGODB_USERNAME" envDefault:"root" json:"username"`
	Password       string        `env:"MONGODB_PASSWORD" envDefault:"root" json:"password"`
	Database       string        `env:"MONGODB_DATABASE" envDefault:"transfers-db" json:"database"`
	Collection     string        `env:"MONGODB_COLLECTION" envDefault:"transfers" json:"collection"`
}

type MySQL struct {
	ConnectTimeout time.Duration `env:"MYSQL_CONNECT_TIMEOUT" envDefault:"10s" json:"connect_timeout"`
	Hostname       string        `env:"MYSQL_HOSTNAME" envDefault:"localhost" json:"hostname"`
	Port           int           `env:"MYSQL_PORT" envDefault:"3306" json:"port"`
	Username       string        `env:"MYSQL_USER" envDefault:"root" json:"username"`
	Password       string        `env:"MYSQL_PASSWORD" envDefault:"root" json:"password"`
	Database       string        `env:"MYSQL_DATABASE" envDefault:"transfers" json:"database"`

	MaxOpenConns    int           `env:"MYSQL_MAX_OPEN_CONNS" envDefault:"25" json:"max_open_conns"`
	MaxIdleConns    int           `env:"MYSQL_MAX_IDLE_CONNS" envDefault:"25" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `env:"MYSQL_CONN_MAX_LIFETIME" envDefault:"5m" json:"conn_max_lifetime"`
}

type CCache struct {
	TTLSeconds     int   `env:"CCACHE_TTL_SECONDS" envDefault:"30" json:"ttl_seconds"`
	MaxSize        int64 `env:"CCACHE_MAX_SIZE" envDefault:"5000" json:"max_size"`
	GetsPerPromote int32 `env:"CCACHE_GETS_PER_PROMOTE" envDefault:"3" json:"gets_per_promote"`
	PercentToPrune uint8 `env:"CCACHE_PERCENT_TO_PRUNE" envDefault:"10" json:"percent_to_prune"`
}

type RabbitMQ struct {
	Hostname  string `env:"RABBITMQ_HOSTNAME" envDefault:"rabbitmq" json:"hostname"`
	Port      int    `env:"RABBITMQ_PORT" envDefault:"5672" json:"port"`
	Username  string `env:"RABBITMQ_USERNAME" envDefault:"guest" json:"username"`
	Password  string `env:"RABBITMQ_PASSWORD" envDefault:"guest" json:"password"`
	QueueName string `env:"RABBITMQ_QUEUE_NAME" envDefault:"transfers-events" json:"queue_name"`
}

func ParseFromEnv() *Config {
	var cfg Config
	for _, nested := range []interface{}{
		&cfg.Business,
		&cfg.MongoDBConfig,
		&cfg.MySQLConfig,
		&cfg.CCacheConfig,
		&cfg.RabbitMQConfig,
	} {
		if err := env.Parse(nested); err != nil {
			logging.Logger.Fatalf("error parsing config: %v", err)
		}
	}
	return &cfg
}

func ParseFromJSON(input []byte) *Config {
	var cfg Config
	if err := json.Unmarshal(input, &cfg); err != nil {
		logging.Logger.Fatalf("error parsing config: %v", err)
	}
	return &cfg
}

func (c *Config) String() string {
	bytes, err := json.Marshal(c)
	if err != nil {
		logging.Logger.Fatalf("error marshaling config: %v", err)
	}
	return string(bytes)
}
