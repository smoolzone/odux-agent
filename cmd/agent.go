package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/spf13/cobra"

	"github.com/smoolzone/odux-agent/internal/agent"
	"github.com/smoolzone/odux-agent/internal/api"
)

var (
	flagAgentToken    string
	flagAgentHub      string
	flagAgentServerID int
	flagAgentInterval int
	flagAgentHostname string
	flagAgentVersion  string
	flagAgentMode     string
)

// monitorRefusal is what the hub is told when a monitor-only agent refuses a command.
const monitorRefusal = "refused: this chamber's agent runs in monitor-only mode (it reports machine stats and " +
	"runs nothing). Only someone with root on the chamber can change that: sudo odux agent mode full"

// agentCmd is the outbound connector: it phones home over TLS so the hub never
// needs inbound access to a client network.
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run the outbound connector that phones home to the hub",
	Long:  "The agent is the designed management path: install it on a chamber, enroll with a one-time token, and it phones home over TLS. No inbound SSH/firewall holes required.",
}

var agentEnrollCmd = &cobra.Command{
	Use:   "enroll",
	Short: "Enroll this host as an agent using a one-time token",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := agent.Load()
		if err != nil {
			return err
		}
		hub := flagAgentHub
		if hub == "" {
			hub = cfg.Hub
		}
		if hub == "" {
			hub = os.Getenv("ODUX_HUB")
		}
		if hub == "" {
			hub = prompt("Hub URL", "https://app.odux.io")
		}
		if err := api.CheckHub(hub); err != nil {
			return err
		}
		mode := cfg.EffectiveMode()
		if flagAgentMode != "" {
			m, err := agent.ParseMode(flagAgentMode)
			if err != nil {
				return err
			}
			mode = m
		}
		token := flagAgentToken
		if token == "" {
			token = os.Getenv("ODUX_ENROLL_TOKEN")
		}
		if token == "" {
			t, err := promptSecret("Enrollment token (en_…)")
			if err != nil {
				return err
			}
			token = strings.TrimSpace(t)
		}
		hostname := flagAgentHostname
		if hostname == "" {
			hostname, _ = os.Hostname()
		}
		version := flagAgentVersion
		if version == "" {
			version = Version
		}

		res, err := api.EnrollAgent(hub, api.EnrollRequest{
			Token:    token,
			Hostname: hostname,
			Version:  version,
			OS:       runtime.GOOS,
			Arch:     runtime.GOARCH,
			ServerID: flagAgentServerID,
		})
		if err != nil {
			return err
		}

		cfg.Hub = hub
		cfg.AgentID = res.AgentID
		cfg.ServerID = res.ServerID
		cfg.Key = res.AgentKey
		cfg.Mode = mode
		if err := agent.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("%s enrolled as agent %d for chamber %q (key stored encrypted at %s)\n",
			emoji(check), res.AgentID, res.ServerName, agent.Path())
		fmt.Println(describeMode(mode))
		return nil
	},
}

var agentRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the heartbeat loop (foreground daemon)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := agent.Load()
		if err != nil {
			return err
		}
		if cfg.Key == "" || cfg.Hub == "" {
			return fmt.Errorf("not enrolled — run `odux agent enroll` first")
		}
		interval := flagAgentInterval
		if interval <= 0 {
			interval = 5
		}
		hostname := flagAgentHostname
		if hostname == "" {
			hostname, _ = os.Hostname()
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

		fmt.Println(describeMode(cfg.EffectiveMode()))

		beat := func() error {
			// Re-read the mode every beat, so `odux agent mode` takes effect without a restart.
			if fresh, err := agent.Load(); err == nil && fresh.Mode != cfg.Mode {
				cfg.Mode = fresh.Mode
				fmt.Println(describeMode(cfg.EffectiveMode()))
			}
			req := api.HeartbeatRequest{
				Hostname: hostname,
				Version:  Version,
				OS:       runtime.GOOS,
				Arch:     runtime.GOARCH,
			}
			if m := agent.CollectMetrics(); m.RamTotal > 0 {
				req.Metrics = &m
			}
			res, err := api.AgentHeartbeat(cfg.Hub, cfg.Key, req)
			if err != nil {
				return err
			}
			if !flagQuiet {
				fmt.Printf("%s heartbeat ok (chamber %d)%s\n",
					emoji(check), res.ServerID, pendingHint(res))
			}
			processCommands(cfg, res.Commands)
			return nil
		}

		// First beat immediately so enrollment is visible right away.
		if err := beat(); err != nil {
			return err
		}
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := beat(); err != nil {
					fmt.Fprintf(os.Stderr, "%s heartbeat failed: %v\n", emoji(cross), err)
				}
			case <-sig:
				fmt.Println("stopping agent")
				return nil
			}
		}
	},
}

// agentUnit is the systemd unit that runs the agent as a daemon (auto-restart, on-boot).
const agentUnit = `[Unit]
Description=Odux agent (phones home to the hub over TLS)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/odux agent run
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`

var agentInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the agent as a systemd service (root)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := os.WriteFile("/etc/systemd/system/odux-agent.service", []byte(agentUnit), 0o644); err != nil {
			return fmt.Errorf("write unit failed (run as root): %w", err)
		}
		for _, c := range [][]string{
			{"systemctl", "daemon-reload"},
			{"systemctl", "enable", "--now", "odux-agent"},
		} {
			if out, err := exec.Command(c[0], c[1:]...).CombinedOutput(); err != nil {
				return fmt.Errorf("%s failed: %s", strings.Join(c, " "), strings.TrimSpace(string(out)))
			}
		}
		fmt.Printf("%s agent installed as systemd service (odux-agent)\n", emoji(check))
		return nil
	},
}

var agentUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the agent systemd service",
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = exec.Command("systemctl", "disable", "--now", "odux-agent").Run()
		_ = os.Remove("/etc/systemd/system/odux-agent.service")
		_ = exec.Command("systemctl", "daemon-reload").Run()
		fmt.Printf("%s agent service removed\n", emoji(check))
		return nil
	},
}

var agentModeCmd = &cobra.Command{
	Use:   "mode [full|monitor]",
	Short: "Show or set what this agent may do: full (run hub commands) or monitor (stats only)",
	Long: "full    — run the commands the hub sends (shell, deploy, terminal). The default.\n" +
		"monitor — only report machine stats; refuse every command from the hub.\n\n" +
		"The mode is stored on this machine (root-owned), so the hub cannot change it.\n" +
		"A running agent picks up a change on its next heartbeat; no restart needed.",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := agent.Load()
		if err != nil {
			return err
		}
		if len(args) == 0 {
			fmt.Println(describeMode(cfg.EffectiveMode()))
			return nil
		}
		mode, err := agent.ParseMode(args[0])
		if err != nil {
			return err
		}
		cfg.Mode = mode
		if err := agent.Save(cfg); err != nil {
			return fmt.Errorf("could not save the mode (run as root): %w", err)
		}
		fmt.Printf("%s %s\n", emoji(check), describeMode(mode))
		return nil
	},
}

// describeMode is the one-line explanation printed wherever the mode matters.
func describeMode(mode string) string {
	if mode == agent.ModeMonitor {
		return "mode: monitor-only — reports machine stats, refuses every command from the hub"
	}
	return "mode: full — runs the commands the hub sends (shell, deploy, terminal)"
}

// processCommands runs (full mode) or refuses (monitor-only) each command from a heartbeat.
func processCommands(cfg *agent.Config, cmds []api.HeartbeatCommand) {
	for _, c := range cmds {
		var err error
		if cfg.MonitorOnly() {
			err = refuseCommand(cfg.Hub, cfg.Key, c)
		} else {
			err = handleCommand(cfg.Hub, cfg.Key, c)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s command %d: %v\n", emoji(cross), c.ID, err)
		}
	}
}

// refuseCommand tells the hub a command was not run because this agent is monitor-only.
func refuseCommand(hub, key string, c api.HeartbeatCommand) error {
	fmt.Printf("%s refused command %d (%s): monitor-only mode\n", emoji(cross), c.ID, c.Type)
	return api.ReportCommandResult(hub, key, c.ID, api.CommandResultRequest{Status: "failed", Error: monitorRefusal})
}

func pendingHint(res *api.HeartbeatResponse) string {
	if len(res.Commands) == 0 {
		return ""
	}
	return fmt.Sprintf(" (%d pending command(s))", len(res.Commands))
}

// handleCommand executes one dequeued command and reports the result back.
func handleCommand(hub, key string, c api.HeartbeatCommand) error {
	fmt.Printf("%s running command %d (%s)\n", emoji(sword), c.ID, c.Type)
	// Terminal sessions are long-lived PTYs streamed over the fast-poll io
	// endpoint. Run them in their own goroutine so the heartbeat loop stays
	// responsive; the session reports its command result when it ends.
	if c.Type == "terminal" {
		go runTerminalSession(hub, key, c)
		return nil
	}
	status, result, errMsg := executeCommand(&c)
	if err := api.ReportCommandResult(hub, key, c.ID, api.CommandResultRequest{
		Status: status,
		Result: result,
		Error:  errMsg,
	}); err != nil {
		return err
	}
	fmt.Printf("%s command %d %s\n", emoji(check), c.ID, status)
	return nil
}

// runTerminalSession opens a PTY shell on the chamber and bridges it to the hub's
// terminal io endpoint on a fast poll (~150ms): chamber output -> hub, hub input /
// resize / close -> the PTY. Ends when the shell exits or the hub asks to close.
func runTerminalSession(hub, key string, c api.HeartbeatCommand) {
	report := func(status, errMsg string) {
		_ = api.ReportCommandResult(hub, key, c.ID, api.CommandResultRequest{
			Status: status,
			Error:  errMsg,
		})
	}

	sid := 0
	if f, ok := c.Params["sessionId"].(float64); ok {
		sid = int(f)
	}
	if sid <= 0 {
		fmt.Fprintf(os.Stderr, "%s terminal %d: missing sessionId param\n", emoji(cross), c.ID)
		report("failed", "terminal command is missing a sessionId param")
		return
	}

	shell := "/bin/bash"
	if !commandExists(shell) {
		shell = "/bin/sh"
	}
	cmd := exec.Command(shell)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		report("failed", "failed to start pty shell: "+err.Error())
		return
	}
	defer ptmx.Close()
	go cmd.Wait() // reap the child once it exits

	// Reader goroutine: drain PTY output into a shared buffer; signal on EOF.
	var mu sync.Mutex
	var outBuf bytes.Buffer
	exited := make(chan struct{})
	go func() {
		buf := make([]byte, 8192)
		for {
			n, rerr := ptmx.Read(buf)
			if n > 0 {
				mu.Lock()
				outBuf.Write(buf[:n])
				mu.Unlock()
			}
			if rerr != nil {
				break
			}
		}
		close(exited)
	}()

	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mu.Lock()
			chunk := outBuf.String()
			outBuf.Reset()
			mu.Unlock()
			res, ioErr := api.TerminalIo(hub, key, sid, api.TerminalIoRequest{Output: chunk})
			if ioErr != nil {
				continue // transient network error — keep polling
			}
			if res.Input != "" {
				_, _ = ptmx.Write([]byte(res.Input))
			}
			if res.Resize != nil {
				_ = pty.Setsize(ptmx, &pty.Winsize{
					Rows: uint16(res.Resize.Rows),
					Cols: uint16(res.Resize.Cols),
				})
			}
			if res.Close {
				_ = cmd.Process.Kill()
				fmt.Printf("%s terminal %d closed by hub\n", emoji(check), c.ID)
				report("succeeded", "")
				return
			}
		case <-exited:
			mu.Lock()
			chunk := outBuf.String()
			outBuf.Reset()
			mu.Unlock()
			exitCode := 0
			if cmd.ProcessState != nil {
				exitCode = cmd.ProcessState.ExitCode()
			}
			if chunk != "" {
				_, _ = api.TerminalIo(hub, key, sid, api.TerminalIoRequest{Output: chunk})
			}
			ec := exitCode
			_, _ = api.TerminalIo(hub, key, sid, api.TerminalIoRequest{Closed: true, ExitCode: &ec})
			fmt.Printf("%s terminal %d shell exited (%d)\n", emoji(check), c.ID, exitCode)
			report("succeeded", "")
			return
		}
	}
}

// executeCommand runs a curated command type. Only "shell" for now; deploy/restart
// land on top of this once the policy + selector primitives exist.
func executeCommand(c *api.HeartbeatCommand) (string, map[string]any, string) {
	switch c.Type {
	case "shell":
		script, _ := c.Params["cmd"].(string)
		if script == "" {
			return "failed", nil, "shell command is missing a string 'cmd' param"
		}
		result, err := runShell(script)
		if err != nil {
			return "failed", result, err.Error()
		}
		return "succeeded", result, ""
	case "deploy":
		return executeDeploy(c.Params)
	default:
		return "failed", nil, fmt.Sprintf("unsupported command type %q", c.Type)
	}
}

// executeDeploy applies a Policy-derived spec: packages, files, env.
func executeDeploy(params map[string]any) (string, map[string]any, string) {
	result := map[string]any{}

	if pkgs, ok := params["packages"].([]any); ok && len(pkgs) > 0 {
		names := make([]string, 0, len(pkgs))
		for _, p := range pkgs {
			if s, ok := p.(string); ok && s != "" {
				names = append(names, s)
			}
		}
		if len(names) > 0 {
			out, err := installPackages(names)
			if err != nil {
				return "failed", result, fmt.Sprintf("package install failed: %v", err)
			}
			result["packages"] = out
		}
	}

	if files, ok := params["files"].([]any); ok {
		var written []string
		for _, f := range files {
			fm, ok := f.(map[string]any)
			if !ok {
				continue
			}
			path, _ := fm["path"].(string)
			content, _ := fm["content"].(string)
			if path == "" {
				continue
			}
			if err := writeFile(path, content); err != nil {
				return "failed", result, fmt.Sprintf("write %s failed: %v", path, err)
			}
			written = append(written, path)
		}
		if len(written) > 0 {
			result["files"] = written
		}
	}

	if envs, ok := params["env"].([]any); ok && len(envs) > 0 {
		if err := writeEnv(envs); err != nil {
			return "failed", result, fmt.Sprintf("env write failed: %v", err)
		}
		result["env"] = "written"
	}

	return "succeeded", result, ""
}

// installPackages installs via the first available package manager.
func installPackages(pkgs []string) (string, error) {
	var name string
	switch {
	case commandExists("apt-get"):
		name = "apt-get"
	case commandExists("dnf"):
		name = "dnf"
	case commandExists("yum"):
		name = "yum"
	default:
		return "", fmt.Errorf("no supported package manager (apt-get/dnf/yum) found")
	}
	args := append([]string{"install", "-y"}, pkgs...)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	return string(out), err
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// writeFile writes a managed file, creating parent dirs (0644).
func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// writeEnv writes resolved env vars to /etc/odux/policy.env (0600).
func writeEnv(envs []any) error {
	var lines []string
	for _, e := range envs {
		m, ok := e.(map[string]any)
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		value, _ := m["value"].(string)
		if name == "" {
			continue
		}
		lines = append(lines, name+"="+value)
	}
	if err := os.MkdirAll("/etc/odux", 0o755); err != nil {
		return err
	}
	return os.WriteFile("/etc/odux/policy.env", []byte(strings.Join(lines, "\n")+"\n"), 0o600)
}

// runShell runs a script with a 5-minute timeout and captures combined output.
func runShell(script string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	result := map[string]any{
		"exitCode": 0,
		"stdout":   stdout.String(),
		"stderr":   stderr.String(),
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			result["exitCode"] = ee.ExitCode()
		} else {
			result["exitCode"] = -1
		}
		return result, err
	}
	return result, nil
}

func init() {
	agentCmd.AddCommand(agentEnrollCmd, agentRunCmd, agentInstallCmd, agentUninstallCmd, agentModeCmd)

	agentEnrollCmd.Flags().StringVar(&flagAgentToken, "token", "", "one-time enrollment token (or ODUX_ENROLL_TOKEN)")
	agentEnrollCmd.Flags().StringVar(&flagAgentHub, "hub", "", "hub URL (default: stored, then ODUX_HUB)")
	agentEnrollCmd.Flags().IntVar(&flagAgentServerID, "server-id", 0, "attach to an existing chamber id")
	agentEnrollCmd.Flags().StringVar(&flagAgentHostname, "hostname", "", "override hostname reported to the hub")
	agentEnrollCmd.Flags().StringVar(&flagAgentVersion, "version", "", "override agent version")
	agentEnrollCmd.Flags().StringVar(&flagAgentMode, "mode", "", "full (run hub commands, default) or monitor (stats only)")

	agentRunCmd.Flags().IntVar(&flagAgentInterval, "interval", 5, "heartbeat interval in seconds")
	agentRunCmd.Flags().StringVar(&flagAgentHostname, "hostname", "", "override hostname reported to the hub")

	rootCmd.AddCommand(agentCmd)
}
