package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type GitlabUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

func (glc *GitlabClient) VerifyToken(token string) (bool, error) {
	url := fmt.Sprintf(API_TOKEN_INFO, glc.Host)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := glc.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return false, nil
	}
	return true, nil
}

func (glc *GitlabClient) GetUser(token string) (*GitlabUser, error) {
	url := fmt.Sprintf(API_GET_USER, glc.Host)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := glc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var user GitlabUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &user, nil
}

// RefreshToken refreshes the OAuth token.
func (glc *GitlabClient) RefreshToken(refreshToken string) (*TokenResponse, error) {
	// Endpoint for refreshing the token.
	endpoint := fmt.Sprintf(API_TOKEN_PATH, glc.Host)

	// Parameters for the refresh token request.
	data := url.Values{
		"client_id":     {glc.ClientId},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
		"redirect_uri":  {glc.redirectURI},
	}

	// Create a POST request.
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request.
	resp, err := glc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for a successful status code.
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}

	// Decode the response.
	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &tokenResponse, nil
}

func (glc *GitlabClient) SetToken(token string) {
	glc.token = token
}
