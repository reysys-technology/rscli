package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// withEnv sets the credential pair every Load() needs plus whatever else the
// case is about, and points HOME at an empty directory so a real
// ~/.reysys/config.yaml on the developer's machine cannot change the result.
func withEnv(t *testing.T, extra map[string]string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RS_CLIENT_ID", "id")
	t.Setenv("RS_CLIENT_SECRET", "secret")
	for k, v := range extra {
		t.Setenv(k, v)
	}
}

func TestRefusesAPlaintextBaseURL(t *testing.T) {
	// Demonstrated 2026-08-22 against v1.1.0: this exact value delivered a live
	// bearer token to a plaintext listener while rscli reported success.
	withEnv(t, map[string]string{"RS_BASE_URL": "http://127.0.0.1:8099"})
	// Loopback is allowed on purpose — that is how the stack runs on a laptop.
	if _, err := Load(); err != nil {
		t.Fatalf("loopback http should be allowed for local development, got %v", err)
	}

	withEnv(t, map[string]string{"RS_BASE_URL": "http://attacker.example"})
	_, err := Load()
	if err == nil {
		t.Fatal("expected a plaintext non-loopback base URL to be refused")
	}
	if !strings.Contains(err.Error(), "attacker.example") || !strings.Contains(err.Error(), "must be https") {
		t.Fatalf("the error must name the host and the rule, got %q", err)
	}
}

func TestRefusesABaseURLThatIsNotAURL(t *testing.T) {
	withEnv(t, map[string]string{"RS_BASE_URL": "not-a-url"})
	if _, err := Load(); err == nil {
		t.Fatal("expected a base URL with no host to be refused")
	}
}

func TestStillRefusesAPlaintextTokenURL(t *testing.T) {
	withEnv(t, map[string]string{"RS_TOKEN_URL": "http://attacker.example/token"})
	if _, err := Load(); err == nil {
		t.Fatal("expected a plaintext token URL to be refused")
	}
}

func TestAcceptsTheDefaults(t *testing.T) {
	withEnv(t, nil)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("defaults must load, got %v", err)
	}
	if !strings.HasPrefix(cfg.BaseURL, "https://") || !strings.HasPrefix(cfg.TokenURL, "https://") {
		t.Fatalf("defaults must be https, got %q and %q", cfg.BaseURL, cfg.TokenURL)
	}
}

func TestIgnoresAConfigFileInTheWorkingDirectory(t *testing.T) {
	// A pull request must not be able to supply its own endpoints by committing
	// a config.yaml, because CI checks the branch out into the working directory.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"),
		[]byte("base_url: http://attacker.example\ntoken_url: http://attacker.example/token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	withEnv(t, nil)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("the working-directory config must be ignored, not fatal: %v", err)
	}
	if strings.Contains(cfg.BaseURL, "attacker") || strings.Contains(cfg.TokenURL, "attacker") {
		t.Fatalf("working-directory config.yaml was honoured: base=%q token=%q", cfg.BaseURL, cfg.TokenURL)
	}
}

func TestTheUploadTimeoutIsGenerousByDefault(t *testing.T) {
	// The old value was two minutes, and a real CI upload hit it: a 4.7 MB image
	// report left the ingest still working at 120 s while the API answered other
	// requests in under a second. The pipeline was told the upload failed and the
	// scan never reached the console, so the gate could not be enabled at all.
	withEnv(t, nil)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPTimeout != defaultHTTPTimeout {
		t.Fatalf("default timeout is %s, want %s", cfg.HTTPTimeout, defaultHTTPTimeout)
	}
	if cfg.HTTPTimeout <= 2*time.Minute {
		t.Fatalf("the default must exceed the value that was measured failing, got %s", cfg.HTTPTimeout)
	}
}

func TestTheUploadTimeoutIsConfigurable(t *testing.T) {
	withEnv(t, map[string]string{"RS_HTTP_TIMEOUT": "25m"})
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPTimeout != 25*time.Minute {
		t.Fatalf("RS_HTTP_TIMEOUT was ignored, got %s", cfg.HTTPTimeout)
	}
}

func TestAnUnusableTimeoutFallsBackRatherThanFailingTheUpload(t *testing.T) {
	// A mistyped timeout must not stop a pipeline from uploading its scan, and a
	// one-second ceiling must not turn every upload into a fake server error.
	for _, value := range []string{"nonsense", "1s", "0", "-5m", "10"} {
		withEnv(t, map[string]string{"RS_HTTP_TIMEOUT": value})
		cfg, err := Load()
		if err != nil {
			t.Fatalf("RS_HTTP_TIMEOUT=%q must not fail the command: %v", value, err)
		}
		if cfg.HTTPTimeout != defaultHTTPTimeout {
			t.Fatalf("RS_HTTP_TIMEOUT=%q gave %s, want the default", value, cfg.HTTPTimeout)
		}
	}
}
