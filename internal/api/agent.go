package api

import "fmt"

// EnrollRequest is the payload for POST /api/agent/enroll (one-time token).
type EnrollRequest struct {
	Token    string `json:"token"`
	Hostname string `json:"hostname"`
	Version  string `json:"version,omitempty"`
	OS       string `json:"os,omitempty"`
	Arch     string `json:"arch,omitempty"`
	ServerID int    `json:"serverId,omitempty"`
}

// EnrollResponse carries the freshly minted long-lived agent key (once).
type EnrollResponse struct {
	AgentID     int    `json:"agentId"`
	ServerID    int    `json:"serverId"`
	ServerName  string `json:"serverName"`
	WorkspaceID int    `json:"workspaceId"`
	AgentKey    string `json:"agentKey"`
}

// EnrollAgent consumes a one-time enrollment token and binds this host to a chamber.
func EnrollAgent(hub string, req EnrollRequest) (*EnrollResponse, error) {
	var out EnrollResponse
	if err := request(hub, "", "POST", "/api/agent/enroll", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Metrics is a machine snapshot the agent reports on each heartbeat (Linux only;
// zeros elsewhere). Ram/disk are in bytes, matching the SSH probe's storage.
type Metrics struct {
	CpuPct    float64 `json:"cpuPct"`
	Load1     float64 `json:"load1"`
	Load5     float64 `json:"load5"`
	Load15    float64 `json:"load15"`
	RamUsed   float64 `json:"ramUsed"`
	RamTotal  float64 `json:"ramTotal"`
	DiskUsed  float64 `json:"diskUsed"`
	DiskTotal float64 `json:"diskTotal"`
	NetRx     float64 `json:"netRx"`
	NetTx     float64 `json:"netTx"`
}

// HeartbeatRequest is the payload for POST /api/agent/heartbeat.
type HeartbeatRequest struct {
	Hostname string   `json:"hostname,omitempty"`
	Version  string   `json:"version,omitempty"`
	OS       string   `json:"os,omitempty"`
	Arch     string   `json:"arch,omitempty"`
	Metrics  *Metrics `json:"metrics,omitempty"`
}

// HeartbeatCommand is a pending command the hub hands back on a heartbeat.
type HeartbeatCommand struct {
	ID     int            `json:"id"`
	Type   string         `json:"type"`
	Params map[string]any `json:"params"`
}

// CommandResultRequest reports the outcome of an executed command.
type CommandResultRequest struct {
	Status string         `json:"status"`
	Result map[string]any `json:"result,omitempty"`
	Error  string         `json:"error,omitempty"`
}

// ReportCommandResult POSTs the outcome of a command back to the hub.
func ReportCommandResult(hub, key string, commandID int, req CommandResultRequest) error {
	return request(hub, key, "POST", fmt.Sprintf("/api/agent/command/%d/result", commandID), req, nil)
}

// HeartbeatResponse is the hub's reply to a heartbeat.
type HeartbeatResponse struct {
	OK       bool               `json:"ok"`
	ServerID int                `json:"serverId"`
	Commands []HeartbeatCommand `json:"commands"`
}

// AgentHeartbeat phones home with the agent key and returns any pending commands.
func AgentHeartbeat(hub, key string, req HeartbeatRequest) (*HeartbeatResponse, error) {
	var out HeartbeatResponse
	if err := request(hub, key, "POST", "/api/agent/heartbeat", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TerminalIoRequest is one agent → hub fast-poll for a live terminal session.
type TerminalIoRequest struct {
	Output   string `json:"output,omitempty"`
	ExitCode *int   `json:"exitCode,omitempty"`
	Closed   bool   `json:"closed,omitempty"`
}

// Resize describes a requested PTY size change.
type Resize struct {
	Rows int `json:"rows"`
	Cols int `json:"cols"`
}

// TerminalIoResponse is the hub's reply: input to feed the PTY + optional resize/close.
type TerminalIoResponse struct {
	Input  string  `json:"input,omitempty"`
	Resize *Resize `json:"resize,omitempty"`
	Close  bool    `json:"close,omitempty"`
}

// TerminalIo exchanges one fast-poll for an agent-backed terminal session.
func TerminalIo(hub, key string, sessionID int, req TerminalIoRequest) (*TerminalIoResponse, error) {
	var out TerminalIoResponse
	if err := request(hub, key, "POST", fmt.Sprintf("/api/agent/terminal/%d/io", sessionID), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
