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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"net/http"
)

type CreateSSHKeyReq struct {
	Title string `json:"title"`
	Key   string `json:"key"`
}

func (glc *GitlabClient) AddSSHKey(title, key string) error {
	url := fmt.Sprintf(API_USER_SSH_KEY_PATH, glc.Host)
	createKeyReq, err := json.Marshal(&CreateSSHKeyReq{
		Title: title,
		Key:   key,
	})

	createKeyReqBuffer := bytes.NewBuffer(createKeyReq)
	req, err := http.NewRequest("POST", url, createKeyReqBuffer)

	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", glc.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := glc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to add SSH key, status: %d", resp.StatusCode)
	}

	glc.logger.Info("SSH key added successfully.")
	return nil
}

func (glc *GitlabClient) DeleteSSHKey(keyID int) error {
	url := fmt.Sprintf(API_USER_SSH_KEY_ID_PATH, glc.Host, keyID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", glc.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := glc.client.Do(req)

	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete SSH key, status: %d", resp.StatusCode)
	}

	glc.logger.Info("SSH key deleted successfully.")
	return nil
}

func (glc *GitlabClient) ListSSHKeys() ([]map[string]interface{}, error) {
	url := fmt.Sprintf(API_USER_SSH_KEY_PATH, glc.Host)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", glc.token))

	resp, err := glc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list SSH keys, status: %d", resp.StatusCode)
	}

	var keys []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	glc.logger.Info("Retrieved %d SSH keys.", len(keys))
	return keys, nil
}

func (glc *GitlabClient) GetExpiredSSH() ([]map[string]interface{}, error) {
	keys, err := glc.ListSSHKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to list SSH keys: %w", err)
	}

	var expiredKeys []map[string]interface{}
	now := time.Now()

	for _, key := range keys {
		if expiresAt, ok := key["expires_at"].(string); ok && expiresAt != "" {
			expirationDate, err := time.Parse(time.RFC3339, expiresAt)
			if err != nil {
				glc.logger.Warn("Failed to parse expiration date for key: %v", key)
				continue
			}
			if expirationDate.Before(now) {
				expiredKeys = append(expiredKeys, key)
			}
		}
	}

	glc.logger.Info("Found %d expired SSH keys.", len(expiredKeys))
	return expiredKeys, nil
}

func (glc *GitlabClient) DeleteSSHKeyByTitlePrefix(prefix string) {
	existingKeys, err := glc.ListSSHKeys()
	if err != nil {
		glc.logger.Fatal("failed to list existing SSH keys: %w", err)
	}

	var wg sync.WaitGroup
	for _, existingKey := range existingKeys {
		if title, ok := existingKey["title"].(string); ok && strings.HasPrefix(title, prefix) {
			wg.Add(1)
			go func(key map[string]interface{}) {
				defer wg.Done()
				keyID, ok := existingKey["id"].(float64)
				if !ok {
					glc.logger.Warn("Invalid key ID format")
				}

				err := glc.DeleteSSHKey(int(keyID))
				if err != nil {
					glc.logger.Warn("Failed to delete key ID %d: %v", int(keyID), err)
				}
				glc.logger.Info("Deleted existing SSH key with ID %d and title %s", int(keyID), title)
			}(existingKey)

		}
	}
	wg.Wait()

}
