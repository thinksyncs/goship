package registry_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/guilhermebr/goship/internal/registry"
)

func TestNew(t *testing.T) {
	reg, err := registry.New(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reg.Handler() == nil {
		t.Fatal("expected non-nil handler")
	}

	if reg.StoreDir() == "" {
		t.Fatal("expected non-empty store dir")
	}
}

func TestRegistryV2Endpoint(t *testing.T) {
	reg, err := registry.New(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ts := httptest.NewServer(reg.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v2/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestCleanupProject(t *testing.T) {
	reg, err := registry.New(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Cleanup non-existent project should not error.
	if err := reg.CleanupProject("nonexistent"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := os.RemoveAll(reg.StoreDir()); err != nil {
		t.Fatal(err)
	}
	if err := reg.CleanupProject("nonexistent"); err != nil {
		t.Fatalf("cleanup with a missing storage root: %v", err)
	}
}

func TestCleanupProjectRejectsTraversal(t *testing.T) {
	dataDir := t.TempDir()
	reg, err := registry.New(dataDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sentinel := filepath.Join(dataDir, "sentinel")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := reg.CleanupProject("../../../sentinel"); err == nil {
		t.Fatal("expected traversal name to be rejected")
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("sentinel changed after rejected cleanup: %v", err)
	}
}

func TestCleanupProjectRejectsSymlinkedRegistryRoot(t *testing.T) {
	dataDir := t.TempDir()
	reg, err := registry.New(dataDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := os.RemoveAll(reg.StoreDir()); err != nil {
		t.Fatal(err)
	}

	outside := t.TempDir()
	victim := filepath.Join(outside, "goship-victim")
	if err := os.MkdirAll(victim, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(victim, "sentinel")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, reg.StoreDir()); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	if err := reg.CleanupProject("victim"); err == nil {
		t.Fatal("CleanupProject should reject a registry root symlink escaping dataDir")
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("CleanupProject followed the registry root symlink: %v", err)
	}
}
