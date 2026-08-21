package pkg

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

// Config is everything rscli needs to talk to a Reysys deployment.
//
// Credentials are an OAuth2 client-credentials pair, minted in the console under
// Account Info ("provision an API client"). They are exchanged at the Keycloak
// token endpoint for a short-lived bearer token; there is no refresh token in
// this grant, so every invocation fetches its own.
//
// The older RS_SECRET_ID / RS_SECRET names are still accepted, because that is
// what existing pipelines have in their secret stores. They mean the same thing.
type Config struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
	TokenURL     string

	// InsecureSkipVerify disables TLS certificate verification. It exists for
	// developers running the stack locally behind a self-signed certificate and
	// must be asked for explicitly — earlier versions of this CLI hard-coded it
	// on, which silently disabled verification while carrying the secret.
	InsecureSkipVerify bool
}

const (
	defaultBaseURL  = "https://api.reysys.com"
	defaultTokenURL = "https://accounts.reysys.com/realms/accounts/protocol/openid-connect/token"
)

// Load reads configuration from, in decreasing precedence: environment
// variables, ./config.yaml, $HOME/.reysys/config.yaml, then defaults.
func Load() (Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	// Only the user's own home directory. Deliberately NOT the working directory:
	// in CI the working directory is the checked-out branch, so reading
	// ./config.yaml would let a pull request supply its own token_url and receive
	// the client credentials. A contributor who cannot edit the workflow could
	// still exfiltrate the secret by adding a file.
	v.AddConfigPath("$HOME/.reysys")

	v.SetEnvPrefix("RS")
	v.AutomaticEnv()

	for key, envs := range map[string][]string{
		"client_id":            {"RS_CLIENT_ID", "RS_SECRET_ID"},
		"client_secret":        {"RS_CLIENT_SECRET", "RS_SECRET"},
		"base_url":             {"RS_BASE_URL"},
		"token_url":            {"RS_TOKEN_URL"},
		"insecure_skip_verify": {"RS_INSECURE_SKIP_VERIFY"},
	} {
		for _, env := range envs {
			_ = v.BindEnv(key, env)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		if _, notFound := err.(viper.ConfigFileNotFoundError); !notFound {
			return Config{}, fmt.Errorf("reading config file: %w", err)
		}
	}

	cfg := Config{
		ClientID:           firstNonEmpty(v.GetString("client_id"), v.GetString("secret_id")),
		ClientSecret:       firstNonEmpty(v.GetString("client_secret"), v.GetString("secret")),
		BaseURL:            strings.TrimRight(firstNonEmpty(v.GetString("base_url"), defaultBaseURL), "/"),
		TokenURL:           firstNonEmpty(v.GetString("token_url"), defaultTokenURL),
		InsecureSkipVerify: v.GetBool("insecure_skip_verify"),
	}

	if err := cfg.checkTokenURL(); err != nil {
		return Config{}, err
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return Config{}, fmt.Errorf(
			"missing credentials: set RS_CLIENT_ID and RS_CLIENT_SECRET " +
				"(provision them in the console under Account Info), " +
				"or put client_id / client_secret in ~/.reysys/config.yaml")
	}
	return cfg, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// checkTokenURL refuses to send credentials somewhere implausible.
//
// The credentials are posted to TokenURL, so whoever controls that value
// receives them. It is settable by environment variable for self-hosted
// deployments, which is legitimate — but it must at least be https, and a
// plaintext endpoint is never acceptable for a client secret.
func (c Config) checkTokenURL() error {
	parsed, err := url.Parse(c.TokenURL)
	if err != nil {
		return fmt.Errorf("token URL %q is not a valid URL: %w", c.TokenURL, err)
	}
	if parsed.Host == "" {
		return fmt.Errorf("token URL %q has no host", c.TokenURL)
	}
	if parsed.Scheme != "https" {
		// Localhost over http is how the stack runs on a laptop.
		if parsed.Scheme == "http" && isLoopback(parsed.Hostname()) {
			return nil
		}
		return fmt.Errorf(
			"refusing to send credentials to %s over %s — the token URL must be https",
			parsed.Host, parsed.Scheme)
	}
	return nil
}

func isLoopback(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}
