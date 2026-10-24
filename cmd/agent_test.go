package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/smoolzone/odux-agent/internal/api"
)

func TestExecuteCommandShell(t *testing.T) {
	status, result, errMsg := executeCommand(&api.HeartbeatCommand{
		ID: 1, Type: "shell", Params: map[string]any{"cmd": "echo hello"},
	})
	if status != "succeeded" {
		t.Fatalf("expected succeeded, got %q (%s)", status, errMsg)
	}
	if out, _ := result["stdout"].(string); out != "hello\n" {
		t.Fatalf("expected stdout %q, got %q", "hello\n", out)
	}
	if code, _ := result["exitCode"].(int); code != 0 {
		t.Fatalf("expected exitCode 0, got %d", code)
	}
}

func TestExecuteCommandShellFailure(t *testing.T) {
	status, result, errMsg := executeCommand(&api.HeartbeatCommand{
		ID: 2, Type: "shell", Params: map[string]any{"cmd": "exit 3"},
	})
	if status != "failed" {
		t.Fatalf("expected failed, got %q", status)
	}
	if code, _ := result["exitCode"].(int); code != 3 {
		t.Fatalf("expected exitCode 3, got %d", code)
	}
	if !strings.Contains(errMsg, "exit status 3") {
		t.Fatalf("expected error to mention exit status, got %q", errMsg)
	}
}

func TestExecuteCommandMissingCmd(t *testing.T) {
	status, _, errMsg := executeCommand(&api.HeartbeatCommand{ID: 3, Type: "shell"})
	if status != "failed" {
		t.Fatalf("expected failed, got %q", status)
	}
	if !strings.Contains(errMsg, "cmd") {
		t.Fatalf("expected error about missing cmd, got %q", errMsg)
	}
}

func TestExecuteCommandUnsupported(t *testing.T) {
	status, _, errMsg := executeCommand(&api.HeartbeatCommand{ID: 4, Type: "reboot"})
	if status != "failed" {
		t.Fatalf("expected failed, got %q", status)
	}
	if !strings.Contains(errMsg, "unsupported") {
		t.Fatalf("expected unsupported error, got %q", errMsg)
	}
}

func TestExecuteDeployFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "app.conf")

	status, result, errMsg := executeDeploy(map[string]any{
		"files": []any{map[string]any{"path": path, "content": "hello"}},
	})
	if status != "succeeded" {
		t.Fatalf("expected succeeded, got %q (%s)", status, errMsg)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected file written: %v", err)
	}
	if string(b) != "hello" {
		t.Fatalf("expected content hello, got %q", string(b))
	}
	files, _ := result["files"].([]string)
	if len(files) != 1 || files[0] != path {
		t.Fatalf("unexpected files result: %v", result["files"])
	}
}
