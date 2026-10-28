package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/smoolzone/odux-agent/internal/agent"
	"github.com/smoolzone/odux-agent/internal/api"
)

// fakeHub records the command results the agent reports. A plain http:// hub is allowed
// here only because it is on 127.0.0.1 (CheckHub).
func fakeHub(t *testing.T) (*httptest.Server, func() []map[string]any) {
	t.Helper()
	var mu sync.Mutex
	var results []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/agent/command/") {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			mu.Lock()
			results = append(results, body)
			mu.Unlock()
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)
	return srv, func() []map[string]any {
		mu.Lock()
		defer mu.Unlock()
		return append([]map[string]any(nil), results...)
	}
}

func touchCommand(path string) api.HeartbeatCommand {
	return api.HeartbeatCommand{ID: 7, Type: "shell", Params: map[string]any{"cmd": "touch " + path}}
}

func TestMonitorOnlyRunsNothingAndTellsTheHubWhy(t *testing.T) {
	srv, results := fakeHub(t)
	marker := filepath.Join(t.TempDir(), "ran")
	cfg := &agent.Config{Hub: srv.URL, Key: "ag_test", Mode: agent.ModeMonitor}

	processCommands(cfg, []api.HeartbeatCommand{
		touchCommand(marker),
		{ID: 8, Type: "terminal", Params: map[string]any{"sessionId": float64(1)}},
	})

	if _, err := os.Stat(marker); err == nil {
		t.Fatal("monitor-only agent ran a shell command")
	}
	got := results()
	if len(got) != 2 {
		t.Fatalf("expected both commands reported as refused, got %d results", len(got))
	}
	for _, r := range got {
		if r["status"] != "failed" || !strings.Contains(r["error"].(string), "monitor-only") {
			t.Fatalf("expected a monitor-only refusal, got %v", r)
		}
	}
}

func TestFullModeRunsCommands(t *testing.T) {
	srv, results := fakeHub(t)
	marker := filepath.Join(t.TempDir(), "ran")
	for _, mode := range []string{agent.ModeFull, ""} { // "" = a config written before modes existed
		_ = os.Remove(marker)
		cfg := &agent.Config{Hub: srv.URL, Key: "ag_test", Mode: mode}
		processCommands(cfg, []api.HeartbeatCommand{touchCommand(marker)})
		if _, err := os.Stat(marker); err != nil {
			t.Fatalf("mode %q: full mode did not run the command", mode)
		}
	}
	if got := results(); len(got) != 2 || got[0]["status"] != "succeeded" {
		t.Fatalf("expected two successful results, got %v", got)
	}
}

func TestTheModeSurvivesSaveAndLoadAndDefaultsToFull(t *testing.T) {
	t.Setenv("ODUX_HOME", t.TempDir())
	t.Setenv("ODUX_HUB", "")
	cfg := &agent.Config{Hub: "https://hub.example", Key: "ag_x", Mode: agent.ModeMonitor}
	if err := agent.Save(cfg); err != nil {
		t.Fatal(err)
	}
	back, err := agent.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !back.MonitorOnly() || back.Key != "ag_x" {
		t.Fatalf("mode or key lost: %+v", back)
	}
	if (&agent.Config{}).EffectiveMode() != agent.ModeFull {
		t.Fatal("an unset mode must mean full (existing chambers keep working)")
	}
	if _, err := agent.ParseMode("root"); err == nil {
		t.Fatal("an unknown mode must be refused")
	}
	if m, _ := agent.ParseMode("monitor-only"); m != agent.ModeMonitor {
		t.Fatal("monitor-only is an alias for monitor")
	}
}
