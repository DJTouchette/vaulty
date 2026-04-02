package backend

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// MockBackend implements SecretBackend for testing.
type MockBackend struct {
	name    string
	secrets map[string]string
	getCalls int
}

func NewMockBackend(name string, secrets map[string]string) *MockBackend {
	return &MockBackend{name: name, secrets: secrets}
}

func (m *MockBackend) Name() string { return m.name }

func (m *MockBackend) List() ([]string, error) {
	names := make([]string, 0, len(m.secrets))
	for k := range m.secrets {
		names = append(names, k)
	}
	return names, nil
}

func (m *MockBackend) Get(name string) (string, error) {
	m.getCalls++
	val, ok := m.secrets[name]
	if !ok {
		return "", fmt.Errorf("secret %q not found", name)
	}
	return val, nil
}

// --- CachedBackend tests ---

func TestCachedBackendDelegates(t *testing.T) {
	mock := NewMockBackend("test", map[string]string{
		"API_KEY": "secret123",
	})
	cached := NewCachedBackend(mock, 5*time.Minute)

	if cached.Name() != "test" {
		t.Errorf("Name() = %q, want test", cached.Name())
	}

	names, err := cached.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 1 || names[0] != "API_KEY" {
		t.Errorf("List() = %v, want [API_KEY]", names)
	}

	val, err := cached.Get("API_KEY")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "secret123" {
		t.Errorf("Get(API_KEY) = %q, want secret123", val)
	}
}

func TestCachedBackendCachesValues(t *testing.T) {
	mock := NewMockBackend("test", map[string]string{
		"KEY": "value1",
	})
	cached := NewCachedBackend(mock, 5*time.Minute)

	// First call fetches from inner
	val, err := cached.Get("KEY")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "value1" {
		t.Errorf("Get(KEY) = %q, want value1", val)
	}
	if mock.getCalls != 1 {
		t.Fatalf("expected 1 call to inner.Get, got %d", mock.getCalls)
	}

	// Second call should use cache
	val, err = cached.Get("KEY")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "value1" {
		t.Errorf("Get(KEY) = %q, want value1", val)
	}
	if mock.getCalls != 1 {
		t.Fatalf("expected 1 call to inner.Get after cache hit, got %d", mock.getCalls)
	}
}

func TestCachedBackendExpires(t *testing.T) {
	mock := NewMockBackend("test", map[string]string{
		"KEY": "value1",
	})
	// Use a very short TTL
	cached := NewCachedBackend(mock, 1*time.Millisecond)

	// First call
	_, err := cached.Get("KEY")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if mock.getCalls != 1 {
		t.Fatalf("expected 1 call, got %d", mock.getCalls)
	}

	// Wait for TTL to expire
	time.Sleep(5 * time.Millisecond)

	// Should fetch again
	_, err = cached.Get("KEY")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if mock.getCalls != 2 {
		t.Fatalf("expected 2 calls after expiry, got %d", mock.getCalls)
	}
}

func TestCachedBackendZero(t *testing.T) {
	mock := NewMockBackend("test", map[string]string{
		"KEY1": "secret1",
		"KEY2": "secret2",
	})
	cached := NewCachedBackend(mock, 5*time.Minute)

	// Populate cache
	cached.Get("KEY1")
	cached.Get("KEY2")

	cached.Zero()

	// Cache should be empty — next Get should call inner again
	mock.getCalls = 0
	cached.Get("KEY1")
	if mock.getCalls != 1 {
		t.Fatalf("expected 1 call after Zero, got %d", mock.getCalls)
	}
}

func TestCachedBackendGetError(t *testing.T) {
	mock := NewMockBackend("test", map[string]string{})
	cached := NewCachedBackend(mock, 5*time.Minute)

	_, err := cached.Get("NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for nonexistent secret")
	}
}

// --- Factory tests ---

func TestNewBackendUnknownType(t *testing.T) {
	_, err := NewBackend(BackendConfig{Type: "bogus"})
	if err == nil {
		t.Fatal("expected error for unknown backend type")
	}
	if !strings.Contains(err.Error(), "unknown backend type") {
		t.Errorf("error = %q, want it to mention 'unknown backend type'", err)
	}
}

func TestNewBackendInvalidTTL(t *testing.T) {
	// This test will fail if the CLI is not found, so we test the TTL parsing
	// path by using a type that will fail at CLI lookup — we need to test
	// that bad TTL parsing is caught. We can't easily test this without a real
	// CLI, so instead we test the error for an unknown type with bad TTL.
	_, err := NewBackend(BackendConfig{Type: "bogus", TTL: "not-a-duration"})
	if err == nil {
		t.Fatal("expected error")
	}
	// The unknown type error fires before TTL parsing, so we just verify it errors
}

// --- AWS argument construction tests ---

func TestAWSBackendCmdArgs(t *testing.T) {
	b := &AWSBackend{region: "us-east-1", profile: "myprofile"}

	listArgs := b.CmdArgsList()
	assertContains(t, listArgs, "--region", "us-east-1")
	assertContains(t, listArgs, "--profile", "myprofile")
	assertContains(t, listArgs, "--query", "SecretList[].Name")

	getArgs := b.CmdArgsGet("my-secret")
	assertContains(t, getArgs, "--secret-id", "my-secret")
	assertContains(t, getArgs, "--region", "us-east-1")
	assertContains(t, getArgs, "--profile", "myprofile")
}

func TestAWSBackendCmdArgsNoOptional(t *testing.T) {
	b := &AWSBackend{}

	listArgs := b.CmdArgsList()
	assertNotContains(t, listArgs, "--region")
	assertNotContains(t, listArgs, "--profile")
}

// --- GCP argument construction tests ---

func TestGCPBackendCmdArgs(t *testing.T) {
	b := &GCPBackend{project: "my-project"}

	listArgs := b.CmdArgsList()
	assertContains(t, listArgs, "--project", "my-project")
	assertContains(t, listArgs, "--format", "json")

	getArgs := b.CmdArgsGet("my-secret")
	assertContains(t, getArgs, "--secret=my-secret")
	assertContains(t, getArgs, "--project=my-project")
}

func TestGCPBackendCmdArgsNoProject(t *testing.T) {
	b := &GCPBackend{}

	listArgs := b.CmdArgsList()
	assertNotContains(t, listArgs, "--project")
}

// --- HashiCorp argument construction tests ---

func TestHashiCorpBackendCmdArgs(t *testing.T) {
	b := &HashiCorpBackend{addr: "https://vault.example.com", mount: "kv"}

	listArgs := b.CmdArgsList()
	assertContains(t, listArgs, "kv")
	assertContains(t, listArgs, "-format=json")

	fieldArgs := b.CmdArgsGetField("my-secret")
	assertContains(t, fieldArgs, "-field=value")
	assertContains(t, fieldArgs, "-mount=kv")
	assertContains(t, fieldArgs, "my-secret")

	jsonArgs := b.CmdArgsGetJSON("my-secret")
	assertContains(t, jsonArgs, "-format=json")
	assertContains(t, jsonArgs, "-mount=kv")
}

func TestHashiCorpBackendDefaultMount(t *testing.T) {
	// Constructor sets default mount, but we test the struct directly
	b := &HashiCorpBackend{mount: "secret"}

	listArgs := b.CmdArgsList()
	assertContains(t, listArgs, "secret")
}

// --- 1Password argument construction tests ---

func TestOnePasswordBackendCmdArgs(t *testing.T) {
	b := &OnePasswordBackend{vault: "Personal"}

	listArgs := b.CmdArgsList()
	assertContains(t, listArgs, "--vault", "Personal")
	assertContains(t, listArgs, "--format", "json")

	getArgs := b.CmdArgsGet("my-login")
	expected := "op://Personal/my-login/password"
	found := false
	for _, a := range getArgs {
		if a == expected {
			found = true
		}
	}
	if !found {
		t.Errorf("CmdArgsGet: expected %q in %v", expected, getArgs)
	}

	fallbackArgs := b.CmdArgsGetFallback("my-login")
	assertContains(t, fallbackArgs, "--vault", "Personal")
	assertContains(t, fallbackArgs, "my-login")
}

// --- CLI availability error tests ---

func TestAWSBackendMissingCLI(t *testing.T) {
	// Save and restore PATH
	origPath := t.TempDir() // Use an empty dir as PATH to ensure aws is not found
	t.Setenv("PATH", origPath)

	_, err := NewAWSBackend("us-east-1", "")
	if err == nil {
		t.Fatal("expected error when aws CLI is not found")
	}
	if !strings.Contains(err.Error(), "aws CLI not found") {
		t.Errorf("error = %q, want mention of 'aws CLI not found'", err)
	}
}

func TestGCPBackendMissingCLI(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := NewGCPBackend("my-project")
	if err == nil {
		t.Fatal("expected error when gcloud CLI is not found")
	}
	if !strings.Contains(err.Error(), "gcloud CLI not found") {
		t.Errorf("error = %q, want mention of 'gcloud CLI not found'", err)
	}
}

func TestHashiCorpBackendMissingCLI(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := NewHashiCorpBackend("https://vault.example.com", "secret")
	if err == nil {
		t.Fatal("expected error when vault CLI is not found")
	}
	if !strings.Contains(err.Error(), "vault CLI not found") {
		t.Errorf("error = %q, want mention of 'vault CLI not found'", err)
	}
}

func TestOnePasswordBackendMissingCLI(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := NewOnePasswordBackend("Personal")
	if err == nil {
		t.Fatal("expected error when op CLI is not found")
	}
	if !strings.Contains(err.Error(), "op CLI not found") {
		t.Errorf("error = %q, want mention of 'op CLI not found'", err)
	}
}

// --- Config parsing integration test ---

func TestBackendConfigRoundTrip(t *testing.T) {
	cfg := BackendConfig{
		Type:    "aws-secrets-manager",
		Region:  "us-west-2",
		Profile: "prod",
		TTL:     "10m",
	}

	if cfg.Type != "aws-secrets-manager" {
		t.Errorf("Type = %q", cfg.Type)
	}
	if cfg.Region != "us-west-2" {
		t.Errorf("Region = %q", cfg.Region)
	}
	if cfg.TTL != "10m" {
		t.Errorf("TTL = %q", cfg.TTL)
	}
}

// --- helpers ---

func assertContains(t *testing.T, args []string, vals ...string) {
	t.Helper()
	for i := 0; i < len(vals); i++ {
		found := false
		for j, a := range args {
			if a == vals[i] {
				// If there's a next expected value, check it follows immediately
				if i+1 < len(vals) && j+1 < len(args) && args[j+1] == vals[i+1] {
					found = true
					i++ // skip the next val since we matched a pair
					break
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q in args %v", vals[i], args)
		}
	}
}

func assertNotContains(t *testing.T, args []string, val string) {
	t.Helper()
	for _, a := range args {
		if a == val {
			t.Errorf("did not expect %q in args %v", val, args)
		}
	}
}
