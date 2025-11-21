package uploadtrivycontainerimagescan

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/reysys-technology/rscli/pkg/config"

	"github.com/spf13/cobra"
)

var scanFilePath string

var Command = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "upload-trivy-container-image-scan",
		RunE: run,
	}
	cmd.Flags().StringVarP(&scanFilePath, "file", "f", "", "Path to the Trivy JSON scan result file (required)")
	return cmd
}()

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
	// Read the Trivy scan JSON file
	scanData, err := os.ReadFile(scanFilePath)
	if err != nil {
		return fmt.Errorf("failed to read scan file: %w", err)
	}

	// Validate that it's valid JSON
	var scanJSON json.RawMessage
	if err := json.Unmarshal(scanData, &scanJSON); err != nil {
		return fmt.Errorf("invalid JSON in scan file: %w", err)
	}

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

	// Now make the API request to upload the scan
	url := fmt.Sprintf("%s/trivy/upload-trivy-container-image-scan", config.BaseURL)

	// Create POST request with the scan data
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(scanData))
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

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Print success response
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Scan uploaded successfully!\n")
	fmt.Printf("Response:\n%s\n", string(body))

	return nil
}
