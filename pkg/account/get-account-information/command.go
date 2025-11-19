package getaccountinformation

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/reysys-technology/rscli/pkg/config"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:  "get-account-information",
	RunE: run,
}

// TokenResponse represents the response from the token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// getSessionToken obtains a JWT session token using the secret credentials
func getSessionToken(client *http.Client) (string, error) {
	url := fmt.Sprintf("%s/token/get-session-token", config.BaseURL)

	// Prepare request body with secret_id and secret
	requestBody := map[string]string{
		"secret_id": config.RsSecretID,
		"secret":    config.RsSecret,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token request: %w", err)
	}

	// Create POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send token request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

func run(cmd *cobra.Command, args []string) error {
	// Create HTTP client with TLS config (for localhost with self-signed cert)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Get session token first
	accessToken, err := getSessionToken(client)
	if err != nil {
		return fmt.Errorf("failed to get session token: %w", err)
	}

	// Now make the actual API request
	url := fmt.Sprintf("%s/account/get-account-information", config.BaseURL)

	// Prepare request body (empty for now, modify as needed)
	requestBody := map[string]interface{}{}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set Authorization header with Bearer token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Print response
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response:\n%s\n", string(body))

	return nil
}
