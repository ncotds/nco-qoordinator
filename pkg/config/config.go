package config

import (
	"bytes"
	"encoding"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogFile    string           `yaml:"log_file" env:"NCOQ_LOG_FILE" env-default:""`
	LogLevel   string           `yaml:"log_level" env:"NCOQ_LOG_LEVEL" env-default:"ERROR"`
	HTTPServer HTTPServerConfig `yaml:"http_server"`
	OMNIbus    OMNIbus          `yaml:"omnibus"`
}

type HTTPServerConfig struct {
	Listen      string        `yaml:"listen" env:"NCOQ_HTTP_LISTEN" env-default:":5000"`
	Timeout     time.Duration `yaml:"timeout" env:"NCOQ_HTTP_TIMEOUT" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"NCOQ_HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

type OMNIbus struct {
	Clusters        map[string]SeedList `yaml:"clusters" env:"NCOQ_OMNI_CLUSTERS" env-default:"AGG1:localhost:4100|localhost:4101"`
	ConnectionLabel string              `yaml:"connection_label" env:"NCOQ_OMNI_CONN_LABEL" env-default:"nco-qoordinator"`
	MaxConnections  int                 `yaml:"max_connections" env:"NCOQ_OMNI_MAX_CONN" env-default:"10"`
	RandomFailOver  bool                `yaml:"random_fail_over" env:"NCOQ_OMNI_RAND_FAILOVER" env-default:"false"`
	FailBack        bool                `yaml:"fail_back" env:"NCOQ_OMNI_FAILBACK" env-default:"true"`
	FailBackDelay   time.Duration       `yaml:"fail_back_delay" env:"NCOQ_OMNI_FAILBACK_DELAY" env-default:"300s"`
}

var _ encoding.TextUnmarshaler = (*SeedList)(nil)

type SeedList []string

func (s *SeedList) UnmarshalText(text []byte) error {
	for _, seed := range bytes.Split(text, []byte("|")) {
		*s = append(*s, string(seed))
	}
	return nil
}

func LoadConfig(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); err != nil {
		return nil, err
	}
	var conf Config
	if err := cleanenv.ReadConfig(configPath, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
