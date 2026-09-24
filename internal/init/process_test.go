package gsinit

import "testing"

func TestProcessLogPathRejectsUnsafeAppName(t *testing.T) {
	path, err := processLogPath("web")
	if err != nil {
		t.Fatalf("processLogPath(web): %v", err)
	}
	if path != "/var/log/goship-web.log" {
		t.Fatalf("processLogPath(web) = %q", path)
	}
	if _, err := processLogPath(`a\b`); err != nil {
		t.Fatalf("processLogPath rejected a Linux filename containing a backslash: %v", err)
	}

	for _, name := range []string{".", "..", "../../outside", "/tmp/x", "a/b", "a//b", "bad\x00name"} {
		if path, err := processLogPath(name); err == nil {
			t.Fatalf("processLogPath(%q) = %q, want error", name, path)
		}
	}
}
