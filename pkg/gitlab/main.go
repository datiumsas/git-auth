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
package gitlab

import (
	"net/http"
	"time"

	l "github.com/atnomoverflow/git-auth/pkg/logger"
)

const (
	API_AUTHORIZE_DEVICE_PATH = "%s/oauth/authorize_device"
	API_TOKEN_PATH            = "%s/oauth/token"
	API_USER_SSH_KEY_PATH     = "%s/api/v4/user/keys"
	API_USER_SSH_KEY_ID_PATH  = "%s/api/v4/user/keys/%d"
	API_TOKEN_INFO            = "%s/oauth/token/info"
	API_GET_USER              = "%s/api/v4/user"
)

type GitlabClient struct {
	Host        string
	scope       string
	ClientId    string
	client      *http.Client
	SshPrefix   string
	logger      *l.Logger
	token       string
	redirectURI string
}

func New(host string, logger *l.Logger, ops ...Options) *GitlabClient {
	glc := &GitlabClient{
		Host: host,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:      logger,
		redirectURI: "urn:ietf:wg:oauth:2.0:oob:auto",
	}
	for _, op := range ops {
		op(glc)
	}
	return glc
}
