package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// AgentConfig is the configuration for the InfraWatch agent.
type AgentConfig struct {
	Agent struct {
		NodeID          string        `yaml:"node_id"`
		CollectInterval time.Duration `yaml:"collect_interval"`
		LogLevel        string        `yaml:"log_level"`
	} `yaml:"agent"`
	Server struct {
		URL        string `yaml:"url"`
		CACert     string `yaml:"ca_cert"`
		ClientCert string `yaml:"client_cert"`
		ClientKey  string `yaml:"client_key"`
	} `yaml:"server"`
	Collectors struct {
		CPU     bool `yaml:"cpu"`
		Memory  bool `yaml:"memory"`
		Disk    bool `yaml:"disk"`
		Network bool `yaml:"network"`
		Docker  bool `yaml:"docker"`
		Systemd bool `yaml:"systemd"`
		Uptime  bool `yaml:"uptime"`
	} `yaml:"collectors"`
	Services []string `yaml:"services"`
}

// ServerConfig is the configuration for the InfraWatch server.
type ServerConfig struct {
	Server struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		LogLevel string `yaml:"log_level"`
		TLS      struct {
			Enabled    bool   `yaml:"enabled"`
			CACert     string `yaml:"ca_cert"`
			ServerCert string `yaml:"server_cert"`
			ServerKey  string `yaml:"server_key"`
		} `yaml:"tls"`
	} `yaml:"server"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
	Auth struct {
		JWTSecret   string        `yaml:"jwt_secret"`
		TokenExpiry time.Duration `yaml:"token_expiry"`
	} `yaml:"auth"`
	Alerts struct {
		EvaluationInterval time.Duration `yaml:"evaluation_interval"`
		Rules              []AlertRuleCfg `yaml:"rules"`
	} `yaml:"alerts"`
	Channels struct {
		Email struct {
			Enabled  bool     `yaml:"enabled"`
			SMTPHost string   `yaml:"smtp_host"`
			SMTPPort int      `yaml:"smtp_port"`
			Username string   `yaml:"username"`
			Password string   `yaml:"password"`
			From     string   `yaml:"from"`
			To       []string `yaml:"to"`
		} `yaml:"email"`
		Slack struct {
			Enabled    bool   `yaml:"enabled"`
			WebhookURL string `yaml:"webhook_url"`
			Channel    string `yaml:"channel"`
		} `yaml:"slack"`
	} `yaml:"channels"`
}

// AlertRuleCfg is an alert rule from the config file.
type AlertRuleCfg struct {
	Name      string  `yaml:"name"`
	Metric    string  `yaml:"metric"`
	Operator  string  `yaml:"operator"`
	Threshold float64 `yaml:"threshold"`
	Severity  string  `yaml:"severity"`
	Enabled   bool    `yaml:"enabled"`
}

// LoadAgentConfig reads and parses the agent YAML configuration file.
func LoadAgentConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AgentConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Agent.CollectInterval == 0 {
		cfg.Agent.CollectInterval = 5 * time.Second
	}
	if cfg.Agent.LogLevel == "" {
		cfg.Agent.LogLevel = "info"
	}
	return &cfg, nil
}

// LoadServerConfig reads and parses the server YAML configuration file.
func LoadServerConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ServerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8443
	}
	if cfg.Auth.TokenExpiry == 0 {
		cfg.Auth.TokenExpiry = 24 * time.Hour
	}
	if cfg.Alerts.EvaluationInterval == 0 {
		cfg.Alerts.EvaluationInterval = 10 * time.Second
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "./infrawatch.db"
	}
	return &cfg, nil
}
