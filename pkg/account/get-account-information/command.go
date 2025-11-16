package getaccountinformation

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"cli/pkg/config"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:  "get-account-information",
	RunE: run,
}

func run(cmd *cobra.Command, args []string) error {
	url := "https://localhost:9670/account/get-account-information"

	// Create HTTP client with TLS config (for localhost with self-signed cert)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

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

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("RS-Secret-ID", config.RsSecretID)
	req.Header.Set("RS-Secret", config.RsSecret)

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
