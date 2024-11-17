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

// magicAuthCmd represents the magicAuth command
var magicAuthCmd = &cobra.Command{
	Use:   "magic-auth",
	Short: "A simple command to authenticate with GitLab, manage tokens, and set up SSH keys automatically.",
	Long: `AThe 'magic-auth' command streamlines the authentication process with GitLab and simplifies SSH key management. This command:
	
- **Manages Tokens:**  
  Fetches and validates stored tokens, refreshes them if expired, or initiates a new login using GitLab's device flow authentication.

- **User Authentication:**  
  Retrieves user information securely, displaying a welcome message for the authenticated user.

- **SSH Key Management:**  
  Automatically deletes previously stored SSH keys with a specific prefix and generates new SSH key pairs. These keys are then added to the authenticated GitLab account, ensuring seamless and secure interactions.

This command is a valuable tool for developers and teams managing GitLab interactions, combining authentication and SSH key setup in a single step.

**Example Usage:**  
magic-auth --profile <profile-name> --ssh-prefix <prefix>
	
Ensure your configuration file is correctly set up before using this command to enjoy effortless GitLab integration.`,
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
		token, err := validateOrRefreshToken(ts, cfg, glc)
		if err == tokenstore.TokenNotFound || err == gitlab.RefreshTokenFailedError {
			newToken := glc.InitDeviceFlow()
			expireAt := time.Now().Unix() + newToken.ExpiresIn
			updatedToken := &tokenstore.Token{
				Profile:      cfg.Profile,
				Token:        newToken.AccessToken,
				RefreshToken: newToken.RefreshToken,
				ExpireAt:     expireAt,
			}
			if err := ts.AddToken(updatedToken); err != nil {
				logger.Warn("error saving updated token: %w", err)
			}
			token = updatedToken
			// set the error to nil
			err = nil

		}
		if err != nil {
			logger.Fatal("unexpected error: %v", err)
		}
		user, err := glc.GetUser(token.Token)
		if err != nil {
			logger.Fatal("unexpected error: %v", err)
		}
		logger.Info("welcome %s", user.Name)
		glc.SetToken(token.Token)
		glc.DeleteSSHKeyByTitlePrefix(cfg.SSHPrefix)
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
	rootCmd.AddCommand(magicAuthCmd)

}
