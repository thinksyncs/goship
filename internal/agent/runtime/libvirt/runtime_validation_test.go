package libvirt

import (
	"bytes"
	"context"
	"testing"

	"github.com/guilhermebr/goship/pkg/domain/entities"
)

func TestRuntimeRejectsUnsafeAppNamesBeforeInstanceLookup(t *testing.T) {
	rt := &Runtime{instances: make(map[string]*instanceInfo)}
	ctx := context.Background()
	unsafeName := "../../outside"

	tests := []struct {
		name string
		call func() error
	}{
		{name: "deploy", call: func() error {
			return rt.DeployApp(ctx, "missing", &entities.AppSpec{Name: unsafeName})
		}},
		{name: "stop", call: func() error { return rt.StopApp(ctx, "missing", unsafeName) }},
		{name: "remove", call: func() error { return rt.RemoveApp(ctx, "missing", unsafeName) }},
		{name: "upload", call: func() error {
			return rt.UploadBinary(ctx, "missing", unsafeName, "server", bytes.NewReader(nil), 0, "")
		}},
		{name: "logs", call: func() error {
			_, err := rt.GetAppLogs(ctx, "missing", unsafeName, 10)
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil || err.Error() == "instance not found: missing" {
				t.Fatalf("runtime did not reject the app name first: %v", err)
			}
		})
	}
}
