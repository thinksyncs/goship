package registry

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/registry"

	"github.com/guilhermebr/goship/pkg/domain/entities"
)

// Registry wraps an OCI-compliant container registry backed by disk storage.
type Registry struct {
	handler  http.Handler
	dataDir  string
	storeDir string
}

// New creates a new Registry with disk-based blob storage under dataDir/registry/.
func New(dataDir string) (*Registry, error) {
	storeDir := filepath.Join(dataDir, "registry")
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create registry storage dir: %w", err)
	}

	blobHandler := registry.NewDiskBlobHandler(storeDir)
	handler := registry.New(registry.WithBlobHandler(blobHandler))

	return &Registry{
		handler:  handler,
		dataDir:  dataDir,
		storeDir: storeDir,
	}, nil
}

// Handler returns the http.Handler for the registry.
func (r *Registry) Handler() http.Handler {
	return r.handler
}

// StoreDir returns the path to the registry's storage directory.
func (r *Registry) StoreDir() string {
	return r.storeDir
}

// CleanupProject removes all registry storage for a project namespace.
// Project images are stored under the namespace "goship-<project>/".
func (r *Registry) CleanupProject(projectName string) error {
	if err := entities.ValidateProjectName(projectName); err != nil {
		return fmt.Errorf("invalid registry project name: %w", err)
	}
	namespace := fmt.Sprintf("goship-%s", projectName)
	root, err := os.OpenRoot(r.dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open registry storage root: %w", err)
	}
	defer func() { _ = root.Close() }()
	if err := root.RemoveAll(filepath.Join("registry", namespace)); err != nil {
		return fmt.Errorf("failed to clean registry project %q: %w", projectName, err)
	}
	return nil
}
