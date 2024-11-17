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
	"time"

	"github.com/atnomoverflow/git-auth/pkg/gitlab"
	tokenstore "github.com/atnomoverflow/git-auth/pkg/token-store"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with GitLab using device flow",
	Long: `This command authenticates the user with the specified GitLab instance using the device flow. 
It checks for an existing token and refreshes it if necessary. If no token is found, it initiates the device flow to get new credentials. 
Once authenticated, the user's GitLab information is fetched and displayed.`,
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
		return

	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
