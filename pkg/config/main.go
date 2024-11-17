package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atnomoverflow/git-auth/pkg/logger"
	"github.com/spf13/viper"
)

type Config struct {
	Profile   string
	URL       string
	ClientID  string
	Scope     []string
	SSHPath   string
	SSHPrefix string
	SSHPort   int
	SSHHost   string
	logger    logger.Logger
}

func (cfg *Config) init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Set the .git-auth directory path where the config file is located.
	configDir := filepath.Join(home, ".git-auth")

	// Check if the .git-auth directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// If the directory doesn't exist, log a warning and proceed with defaults
		cfg.logger.Warn("Configuration directory .git-auth does not exist. Using default values.")
	} else if err != nil {
		// Any other error when accessing the directory
		return fmt.Errorf("failed to access configuration directory: %w", err)
	}

	// Add the .git-auth directory to the config search path
	viper.AddConfigPath(configDir)
	viper.SetConfigType("toml")
	viper.SetConfigName("config") // Assuming the file is named 'config.toml'

	// Set default values if not present in the config file
	viper.SetDefault("ssh-path", filepath.Join(home, ".ssh"))
	viper.SetDefault("ssh-ttl", 7*24*time.Hour)
	viper.SetDefault("ssh-prefix", "gl_auth")
	viper.SetDefault("profile", "default")

	// Automatically read environment variables with a prefix (optional)
	viper.SetEnvPrefix("git_auth")
	viper.AutomaticEnv()

	// Read the configuration file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, log it and continue with default values
			cfg.logger.Warn("Configuration file not found. Using default values.")
		} else {
			// Any other error when reading the config file
			return fmt.Errorf("failed to read configuration file: %w", err)
		}
	}
	return nil
}

func LoadConfig(logger logger.Logger) (*Config, error) {
	cfg := &Config{
		logger: logger,
	}

	// Initialize configuration and check for errors
	if err := cfg.init(); err != nil {
		return nil, err
	}
	// Set values from viper to the config fields
	// First, load the common/default configuration values (profile is the profile key)
	profile := viper.GetString("profile")
	if profile == "" {
		return nil, fmt.Errorf("profile not specified in the configuration")
	}
	cfg.Profile = profile
	cfg.URL = viper.GetString(fmt.Sprintf("%s.url", cfg.Profile))
	if cfg.URL == "" {
		return nil, fmt.Errorf("No profile named %s", profile)
	}
	cfg.ClientID = viper.GetString(fmt.Sprintf("%s.client-id", cfg.Profile))
	cfg.Scope = viper.GetStringSlice(fmt.Sprintf("%s.scope", cfg.Profile))

	sshPath := viper.GetString(fmt.Sprintf("%s.ssh-path", cfg.Profile))
	if strings.HasPrefix(sshPath, "~/") {
		home, _ := os.UserHomeDir()
		sshPath = filepath.Join(home, sshPath[2:])
	}
	if sshPath == "" {
		sshPath = viper.GetString("ssh-path")
	}

	cfg.SSHPath = sshPath

	sshPrefix := viper.GetString(fmt.Sprintf("%s.ssh-prefix", cfg.Profile))
	if sshPrefix == "" {
		sshPrefix = viper.GetString("ssh-prefix")
	}
	cfg.SSHPrefix = sshPrefix

	sshPort := viper.GetInt(fmt.Sprintf("%s.ssh-port", cfg.Profile))

	if sshPort == 0 {
		sshPort = viper.GetInt("ssh-prefix")
	}
	cfg.SSHPort = sshPort
	sshHost := viper.GetString(fmt.Sprintf("%s.ssh-host", cfg.Profile))

	if sshHost == "" {
		sshHost = viper.GetString("ssh-host")
	}
	cfg.SSHHost = sshHost

	return cfg, nil
}
