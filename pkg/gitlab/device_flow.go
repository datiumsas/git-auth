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
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type requestDeviceAuthorization struct {
	ClientId string `json:"client_id"`
	Scope    string `json:"scope"`
}

type DeviceFlowResp struct {
	DeviceCode              string `json:"device_code"`
	VerificationUriComplete string `json:"verification_uri_complete"`
	ExpireIn                int64  `json:"expire_in"`
	Interval                int64  `json:"interval"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (glc *GitlabClient) RequestDeviceAuthorization() (*DeviceFlowResp, error) {
	data, err := json.Marshal(requestDeviceAuthorization{
		ClientId: glc.ClientId,
		Scope:    glc.scope,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshalling device authorization request: %w", err)
	}

	responseBody := bytes.NewBuffer(data)
	resp, err := glc.client.Post(fmt.Sprintf(API_AUTHORIZE_DEVICE_PATH, glc.Host), "application/json", responseBody)
	if err != nil {
		return nil, fmt.Errorf("error making device authorization request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device authorization failed: %s", body)
	}

	var metadata DeviceFlowResp
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling device authorization response: %w", err)
	}

	return &metadata, nil
}

func (glc *GitlabClient) PullToken(deviceCode string, interval time.Duration) (*TokenResponse, error) {
	url := fmt.Sprintf(API_TOKEN_PATH, glc.Host)
	formData := fmt.Sprintf("client_id=%s&device_code=%s&grant_type=urn:ietf:params:oauth:grant-type:device_code", glc.ClientId, deviceCode)

	client := &http.Client{}

	for {
		// Make the token request
		resp, err := client.Post(url, "application/x-www-form-urlencoded", bytes.NewBufferString(formData))
		if err != nil {
			return nil, fmt.Errorf("error polling for token: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}

		switch resp.StatusCode {
		case http.StatusOK: // Token retrieved successfully
			var tokenResponse TokenResponse
			err = json.Unmarshal(body, &tokenResponse)
			if err != nil {
				return nil, fmt.Errorf("error unmarshalling token response: %w", err)
			}
			return &tokenResponse, nil

		case http.StatusBadRequest: // Handle known errors like `authorization_pending` or `slow_down`
			var errorResponse struct {
				Error            string `json:"error"`
				ErrorDescription string `json:"error_description"`
			}
			err = json.Unmarshal(body, &errorResponse)
			if err != nil {
				return nil, fmt.Errorf("error unmarshalling error response: %w", err)
			}

			// Handle specific error types
			switch errorResponse.Error {
			case "authorization_pending":
				// Continue polling
			case "slow_down":
				// log.Println("Received slow_down signal. Increasing polling interval.")
				interval += 5 * time.Second // Adjust polling interval
			case "access_denied":
				return nil, errors.New("access denied by the user")
			case "expired_token":
				return nil, errors.New("device code has expired")
			default:
				return nil, fmt.Errorf("unexpected error: %s - %s", errorResponse.Error, errorResponse.ErrorDescription)
			}

		default: // Handle unexpected status codes
			return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		}

		// Wait for the specified interval before the next request
		time.Sleep(interval)
	}
}

func (glc *GitlabClient) InitDeviceFlow() *TokenResponse {
	deviceResp, err := glc.RequestDeviceAuthorization()
	if err != nil {
		glc.logger.Fatal("Error requesting device authorization: %w", err)
	}
	code := strings.Split(deviceResp.VerificationUriComplete, "=")[1]
	glc.logger.Info(`
Welcome to Gl Auth
Visit the following URL to authorize the device: %s
Make sure to verify the code is correct.
code is %s
	`, deviceResp.VerificationUriComplete, code)

	tokenResp, err := glc.PullToken(deviceResp.DeviceCode, time.Duration(deviceResp.Interval)*time.Second)
	if err != nil {
		glc.logger.Fatal("Error pulling token: %w", err)
	}
	glc.token = tokenResp.AccessToken
	return tokenResp
}
