package cmd

import (
	"github.com/atnomoverflow/git-auth/pkg/ssh"
	"github.com/spf13/cobra"
)

type SSHConfig struct {
	Host         string
	HostName     string
	Port         int
	IdentityFile string
}

var generateSshConfigCmd = &cobra.Command{
	Use:   "generate-ssh-config",
	Short: "Generates and appends SSH configuration for a given host",
	Long: `This command generates an SSH configuration file for the provided host,
if one does not already exist. It adds the host, hostname, port, and identity file
details into the user's SSH config file located at ~/.ssh/config. If the configuration
already exists, it will update the configuration for the host rather than appending.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize config and client (assuming initializeConfigAndGitLabClient exists)
		cfg, _, err := initializeConfigAndGitLabClient()
		if err != nil {
			logger.Fatal("Initialization failed: %v", err)
		}
		sshManager := ssh.New(cfg.SSHHost,
			ssh.WithKeyName(cfg.Profile),
			ssh.WithPath(cfg.SSHPath),
			ssh.WithPort(cfg.SSHPort))
		err = sshManager.AddSSHConfig()
		if err != nil {
			logger.Fatal("SSH config genration failed: %v", err)
		}
		logger.Info("SSH configuration successfully generated and appended for host: %s", cfg.SSHHost)

	},
}

func init() {
	rootCmd.AddCommand(generateSshConfigCmd)
}
