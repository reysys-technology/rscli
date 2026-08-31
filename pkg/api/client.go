// Package api is the one place rscli talks to a Reysys deployment.
//
// Every command shares it, so authentication, TLS and error handling are
// defined once. Previously each command carried its own copy of the token
// exchange and its own http.Client, and both copies disabled TLS verification.
package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/reysys-technology/rscli/pkg"
)

// Client is an authenticated handle on one deployment. Build it with New and
// reuse it; the bearer token is fetched once and kept for the process lifetime,
// which is a single CI step.
type Client struct {
	cfg   pkg.Config
	http  *http.Client
	token string
}

// New returns a client that has already authenticated, so a credential problem
// surfaces immediately rather than on the first upload.
func New(ctx context.Context, cfg pkg.Config) (*Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &Client{
		cfg:  cfg,
		http: &http.Client{Transport: transport, Timeout: cfg.HTTPTimeout},
	}
	token, err := client.fetchToken(ctx)
	if err != nil {
		return nil, err
	}
	client.token = token
	return client, nil
}

// tokenTimeout bounds the credential exchange separately from an upload. The
// two have nothing in common: a token request is a few hundred bytes and either
// answers quickly or is not going to. Waiting the upload timeout on an
// unreachable identity provider would turn a wrong hostname into a ten-minute
// hang in someone's pipeline.
const tokenTimeout = 60 * time.Second

// fetchToken runs the OAuth2 client-credentials grant against Keycloak. This is
// the same exchange the scanner agent performs; the backend reads the account
// from the token's azp claim.
func (c *Client) fetchToken(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, tokenTimeout)
	defer cancel()

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.cfg.ClientID)
	form.Set("client_secret", c.cfg.ClientSecret)

	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("building token request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.http.Do(request)
	if err != nil {
		return "", fmt.Errorf("reaching the identity provider at %s: %w", c.cfg.TokenURL, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("reading token response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		// Deliberately does not echo the body verbatim: a failed token request
		// can quote back the client_id, and CI logs are widely readable.
		return "", fmt.Errorf(
			"authentication failed (HTTP %d) — check RS_CLIENT_ID and RS_CLIENT_SECRET",
			response.StatusCode)
	}

	var parsed struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("decoding token response: %w", err)
	}
	if parsed.AccessToken == "" {
		return "", fmt.Errorf("identity provider returned no access token")
	}
	return parsed.AccessToken, nil
}

// PostJSON sends body to an endpoint path (for example "/trivy.json.ingest")
// and returns the raw response body. A non-2xx status is an error carrying the
// server's message, which is what a pipeline log needs to be actionable.
func (c *Client) PostJSON(ctx context.Context, path string, body []byte) ([]byte, error) {
	endpoint := c.cfg.BaseURL + path
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building request for %s: %w", path, err)
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", "application/json")

	response, err := c.http.Do(request)
	if err != nil {
		// A timeout here is not the same as an unreachable server, and saying so
		// matters: the ingest is synchronous, so when the client gives up the
		// server is usually still working and may well finish. The scan is then
		// in the console while the pipeline was told the upload failed.
		if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
			return nil, fmt.Errorf(
				"%s did not answer within %s. Large reports take longer to store than to send, "+
					"and the upload may still have completed on the server — check the console "+
					"before assuming it did not. Raise RS_HTTP_TIMEOUT (a duration such as 20m) "+
					"if this recurs: %w",
				endpoint, c.cfg.HTTPTimeout, err)
		}
		return nil, fmt.Errorf("reaching %s: %w", endpoint, err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response from %s: %w", path, err)
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return responseBody, fmt.Errorf("%s returned HTTP %d: %s",
			path, response.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return responseBody, nil
}
