package entities

import "testing"

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "myapp"},
		{name: "my-project"},
		{name: "model..v1"},
		{name: `a\b`},
		{name: "", wantErr: true},
		{name: ".", wantErr: true},
		{name: "..", wantErr: true},
		{name: "../registry", wantErr: true},
		{name: "a/../../target", wantErr: true},
		{name: "/tmp/x", wantErr: true},
		{name: "a//b", wantErr: true},
		{name: "bad\x00name", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectName(tt.name)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateProjectName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAppName(t *testing.T) {
	for _, name := range []string{"web", "api-server", "model..v1", `a\b`} {
		if err := ValidateAppName(name); err != nil {
			t.Fatalf("ValidateAppName(%q) error = %v", name, err)
		}
	}
	for _, name := range []string{"", ".", "..", "../../outside", "/tmp/x", "a/b", "a//b", "bad\x00name"} {
		if err := ValidateAppName(name); err == nil {
			t.Fatalf("ValidateAppName(%q) should fail", name)
		}
	}
}
