package commands

import "testing"

func TestServerProbeURL(t *testing.T) {
	tests := []struct {
		listenAddr string
		want       string
	}{
		{listenAddr: "127.0.0.1:8080", want: "http://127.0.0.1:8080"},
		{listenAddr: ":8080", want: "http://127.0.0.1:8080"},
		{listenAddr: "0.0.0.0:8080", want: "http://127.0.0.1:8080"},
		{listenAddr: "[::]:8080", want: "http://127.0.0.1:8080"},
		{listenAddr: "[::1]:8080", want: "http://[::1]:8080"},
		{listenAddr: "localhost:8080", want: "http://localhost:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.listenAddr, func(t *testing.T) {
			got, err := serverProbeURL(tt.listenAddr)
			if err != nil {
				t.Fatalf("serverProbeURL() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("serverProbeURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestServerProbeURLRejectsInvalidAddress(t *testing.T) {
	if _, err := serverProbeURL("localhost"); err == nil {
		t.Fatal("expected missing port to be rejected")
	}
}
