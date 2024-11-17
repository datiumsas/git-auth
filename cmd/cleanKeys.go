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
	"github.com/atnomoverflow/git-auth/pkg/gitlab"
	tokenstore "github.com/atnomoverflow/git-auth/pkg/token-store"
	"github.com/spf13/cobra"
)

var sshKeyPrefix string

// cleanKeysCmd represents the cleanKeys command
var cleanKeysCmd = &cobra.Command{
	Use:   "clean-keys",
	Short: "Clean up SSH keys by deleting keys with a specific prefix",
	Long: `This command deletes SSH keys from the specified GitLab instance that match the provided prefix.
It fetches the authentication token from the cache, verifies if the user is logged in, and checks if the token is valid. 
If not, it will attempt to refresh the token. After successful authentication, it removes SSH keys associated with the configured prefix.`,
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

		if sshKeyPrefix == "" {
			sshKeyPrefix = cfg.SSHPrefix
		}
		glc.DeleteSSHKeyByTitlePrefix(sshKeyPrefix)

	},
}

func init() {
	rootCmd.AddCommand(cleanKeysCmd)
	cleanKeysCmd.Flags().StringVarP(&sshKeyPrefix, "key-prefix", "x", "", "The prefix to match SSH keys for deletion")

}
