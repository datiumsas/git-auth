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
	"path/filepath"
	"time"

	"github.com/atnomoverflow/git-auth/pkg/config"
	"github.com/atnomoverflow/git-auth/pkg/gitlab"
	l "github.com/atnomoverflow/git-auth/pkg/logger"
	tokenstore "github.com/atnomoverflow/git-auth/pkg/token-store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var (
	logger  = l.New(l.DEBUG)
	rootCmd = &cobra.Command{
		Use:   "git-auth",
		Short: "A simple CLI to manage SSH keys for gitlab",
		Long: `git-auth is CLI library that helps manage SSH for gitlab.
It genrate an ssh and adds it to your gitlab account. 
It also delete any expired key that was created by the cli.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("profile", "p", "", "profile to be used")
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))

}

// initializeConfigAndGitLabClient loads the configuration and initializes the GitLab client
func initializeConfigAndGitLabClient() (*config.Config, *gitlab.GitlabClient, error) {
	cfg, err := config.LoadConfig(*logger)
	if err != nil {
		return nil, nil, fmt.Errorf("error loading config: %w", err)
	}

	glc := gitlab.New(
		cfg.URL,
		logger,
		gitlab.WithClientId(cfg.ClientID),
		gitlab.WithScope(cfg.Scope),
		gitlab.WithSshPrefix(cfg.SSHPrefix),
	)

	return cfg, glc, nil
}

// initializeTokenStore sets up the token store
func initializeTokenStore() (*tokenstore.TokenStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, ".git-auth")
	ts := tokenstore.New(configDir)
	return ts, nil
}

// validateOrRefreshToken validates or refreshes the token, returning the updated token
func validateOrRefreshToken(ts *tokenstore.TokenStore, cfg *config.Config, glc *gitlab.GitlabClient) (*tokenstore.Token, error) {
	token, err := ts.GetToken(cfg.Profile)
	if err != nil {
		return nil, tokenstore.TokenNotFound
	}
	if token == nil {
		return nil, tokenstore.TokenNotFound
	}

	isValid, err := glc.VerifyToken(token.Token)
	if err != nil {
		logger.Debug("verify token error %w", err)
		return nil, fmt.Errorf("error verifying token: %w", err)
	}

	if isValid {
		glc.SetToken(token.Token)
		return token, nil
	}

	newToken, err := glc.RefreshToken(token.RefreshToken)
	if err != nil {
		logger.Debug("token refresh failed; please login using the auth command: %w", err)
		return nil, gitlab.RefreshTokenFailedError
	}

	expireAt := time.Now().Unix() + newToken.ExpiresIn
	updatedToken := &tokenstore.Token{
		Profile:      cfg.Profile,
		Token:        newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		ExpireAt:     expireAt,
	}
	if err := ts.AddToken(updatedToken); err != nil {
		return nil, fmt.Errorf("error saving updated token: %w", err)
	}

	glc.SetToken(newToken.AccessToken)
	return updatedToken, nil
}
