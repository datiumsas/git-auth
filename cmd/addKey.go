/*
Copyright © 2024 Montasser abed majid zehri <montasser.zehri@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/atnomoverflow/git-auth/pkg/gitlab"
	"github.com/atnomoverflow/git-auth/pkg/ssh"
	tokenstore "github.com/atnomoverflow/git-auth/pkg/token-store"
	"github.com/spf13/cobra"
)

// addKeyCmd represents the addKey command
var addKeyCmd = &cobra.Command{
	Use:   "add-key",
	Short: "Generate an SSH key pair and add it to GitLab.",
	Long: `The add-key command generates a new SSH key pair, stores it locally, 
	and adds the public key to your GitLab account. It automatically assigns 
	a title to the key based on the configured prefix and the current timestamp.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, glc, err := initializeConfigAndGitLabClient()
		if err != nil {
			logger.Fatal("Initialization failed: %v", err)
		}
		// fetch token from cache and check if we need new login
		ts, err := initializeTokenStore()
		if err != nil {
			logger.Fatal("Token store setup failed: %v", err)
		}
		_, err = validateOrRefreshToken(ts, cfg, glc)
		if err == tokenstore.TokenNotFound || err == gitlab.RefreshTokenFailedError {
			logger.Fatal("User not logged in!")
		}
		if err != nil {
			logger.Fatal("unexpected error: %v", err)
		}
		title := fmt.Sprintf("%s-%s", cfg.SSHPrefix, time.Now().Format("20060102_150405"))
		//genrate ssh key paire to be added
		sshManager := ssh.New(cfg.SSHHost,
			ssh.WithKeyName(cfg.Profile),
			ssh.WithPath(cfg.SSHPath),
			ssh.WithPort(cfg.SSHPort))

		privateKeyPath, publicKeyPath, err := sshManager.GenerateSSHKeyPair()
		if err != nil {
			logger.Fatal("Error generating SSH key pair: %v\n", err)
		}

		logger.Info("Generated keys:\nPrivate: %s\nPublic: %s\n", privateKeyPath, publicKeyPath)

		// Read the public key
		publicKey, err := os.ReadFile(publicKeyPath)
		if err != nil {
			logger.Fatal("Error reading public key: %v\n", err)
		}

		// Add the public key to GitLab
		if err := glc.AddSSHKey(title, string(publicKey)); err != nil {
			logger.Fatal("Error adding SSH key to GitLab: %v\n", err)
		}

	},
}

func init() {
	rootCmd.AddCommand(addKeyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addKeyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addKeyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
