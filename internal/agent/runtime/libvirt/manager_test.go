package libvirt

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestDomainName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"myvm", "goship-myvm"},
		{"test-vm", "goship-test-vm"},
		{"", "goship-"},
	}

	for _, tc := range tests {
		result := domainName(tc.input)
		if result != tc.expected {
			t.Errorf("domainName(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestUserFacingName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"goship-myvm", "myvm"},
		{"goship-test-vm", "test-vm"},
		{"goship-", ""},
		{"other-vm", "other-vm"}, // no prefix to strip
	}

	for _, tc := range tests {
		result := userFacingName(tc.input)
		if result != tc.expected {
			t.Errorf("userFacingName(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestVmDir(t *testing.T) {
	result, err := vmDir("/home/user/.goship", "myvm")
	if err != nil {
		t.Fatalf("vmDir() error = %v", err)
	}
	expected := filepath.Join("/home/user/.goship", "vms", "myvm")
	if result != expected {
		t.Errorf("vmDir() = %q, want %q", result, expected)
	}
}

func TestDiskPath(t *testing.T) {
	result, err := diskPath("/home/user/.goship", "myvm")
	if err != nil {
		t.Fatalf("diskPath() error = %v", err)
	}
	expected := filepath.Join("/home/user/.goship", "vms", "myvm", "disk.qcow2")
	if result != expected {
		t.Errorf("diskPath() = %q, want %q", result, expected)
	}
}

func TestVMPathsRejectUnsafeNames(t *testing.T) {
	for _, name := range []string{"", ".", "..", "../registry", "a/../../target", "/tmp/x", "a//b", "bad\x00name"} {
		t.Run(name, func(t *testing.T) {
			if _, err := vmDir(t.TempDir(), name); err == nil {
				t.Fatalf("vmDir(%q) should fail", name)
			}
			if _, err := diskPath(t.TempDir(), name); err == nil {
				t.Fatalf("diskPath(%q) should fail", name)
			}
		})
	}
}

func TestVMPathsPreserveLinuxBackslashNames(t *testing.T) {
	if _, err := vmDir(t.TempDir(), `a\b`); err != nil {
		t.Fatalf("vmDir rejected a Linux filename containing a backslash: %v", err)
	}
}

func TestVMManagerRejectsUnsafeNameBeforeFilesystemUse(t *testing.T) {
	root := t.TempDir()
	sentinel := filepath.Join(root, "sentinel")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}

	m := &VMManager{dataDir: filepath.Join(root, "data")}
	if _, err := m.Create(CreateVMOptions{Name: "../..", BaseImage: filepath.Join(root, "missing")}); err == nil {
		t.Fatal("Create should reject an unsafe name")
	}
	if _, err := m.Destroy("../..", false); err == nil {
		t.Fatal("Destroy should reject an unsafe name")
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("sentinel changed after rejected operations: %v", err)
	}
}

func TestRemoveVMDirDoesNotFollowFinalSymlink(t *testing.T) {
	dataDir := t.TempDir()
	vmsRoot := filepath.Join(dataDir, "vms")
	if err := os.MkdirAll(vmsRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	sentinel := filepath.Join(outside, "sentinel")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(vmsRoot, "linked")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	m := &VMManager{dataDir: dataDir}
	if err := m.removeVMDir("linked"); err != nil {
		t.Fatalf("removeVMDir() error = %v", err)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("removeVMDir followed the symlink: %v", err)
	}
}

func TestRemoveVMDirRejectsSymlinkedVMRoot(t *testing.T) {
	dataDir := t.TempDir()
	outside := t.TempDir()
	victim := filepath.Join(outside, "victim")
	if err := os.MkdirAll(victim, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(victim, "sentinel")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dataDir, "vms")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	m := &VMManager{dataDir: dataDir}
	if err := m.removeVMDir("victim"); err == nil {
		t.Fatal("removeVMDir should reject a VM root symlink escaping dataDir")
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("removeVMDir followed the VM root symlink: %v", err)
	}
}

func TestGenerateUUID(t *testing.T) {
	uuid1, err := GenerateUUID()
	if err != nil {
		t.Fatalf("GenerateUUID failed: %v", err)
	}

	uuid2, err := GenerateUUID()
	if err != nil {
		t.Fatalf("GenerateUUID failed: %v", err)
	}

	// Verify v4 UUID format: 8-4-4-4-12 hex chars
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !uuidPattern.MatchString(uuid1) {
		t.Errorf("UUID %q does not match v4 format", uuid1)
	}
	if !uuidPattern.MatchString(uuid2) {
		t.Errorf("UUID %q does not match v4 format", uuid2)
	}

	// Verify uniqueness
	if uuid1 == uuid2 {
		t.Errorf("two generated UUIDs should be different, got %s both times", uuid1)
	}
}

func TestEnsurePathTraversable(t *testing.T) {
	// Create a temp dir hierarchy with a restricted parent.
	root := t.TempDir()
	restricted := filepath.Join(root, "restricted")
	inner := filepath.Join(restricted, "inner", "deep")

	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	// Remove o+x from the restricted dir.
	if err := os.Chmod(restricted, 0o700); err != nil {
		t.Fatal(err)
	}

	// Verify it's not traversable.
	info, _ := os.Stat(restricted)
	if info.Mode().Perm()&0o001 != 0 {
		t.Fatal("expected restricted dir to not have o+x")
	}

	// Fix it.
	if err := ensurePathTraversable(inner); err != nil {
		t.Fatalf("ensurePathTraversable failed: %v", err)
	}

	// Verify o+x was added.
	info, _ = os.Stat(restricted)
	if info.Mode().Perm()&0o001 == 0 {
		t.Error("expected restricted dir to have o+x after ensurePathTraversable")
	}
}
