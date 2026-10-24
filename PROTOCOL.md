# Odux agent protocol

The `odux` agent runs on a chamber (a customer's server) and phones home to the Odux
hub over HTTPS. It **runs as root** and **opens no inbound port** — every interaction
is initiated by the agent over an outbound TLS connection.

## Authentication

- **Enrollment** presents a **single-use enrollment token** (`en_…`) minted by the hub.
- In exchange the hub returns a **per-chamber agent key** (`ag_…`). The agent stores it
  locally and sends it as a `Bearer` token on every later request.
- The key stays valid until the chamber is unenrolled.

## Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/api/agent/enroll` | Exchange a one-time token for an agent key |
| POST | `/api/agent/heartbeat` | Phone home every 5 s; receive pending commands |
| POST | `/api/agent/command/:id/result` | Report the outcome of an executed command |
| POST | `/api/agent/terminal/:sessionId/io` | Fast-poll (~150 ms) one interactive terminal session |

### `POST /api/agent/enroll`

Request:

```json
{
  "token": "en_…",
  "hostname": "web-01",
  "version": "0.2.0",
  "os": "linux",
  "arch": "amd64",
  "serverId": 0
}
```

Response (the key is returned exactly once):

```json
{ "agentId": 3, "serverId": 12, "serverName": "web-01", "workspaceId": 1, "agentKey": "ag_…" }
```

### `POST /api/agent/heartbeat`

Sent every 5 seconds. Carries an optional machine snapshot (Linux only; zeros elsewhere).

```json
{
  "hostname": "web-01",
  "version": "0.2.0",
  "os": "linux",
  "arch": "amd64",
  "metrics": {
    "cpuPct": 12.5,
    "load1": 0.4,
    "load5": 0.3,
    "load15": 0.2,
    "ramUsed": 1073741824,
    "ramTotal": 4294967296,
    "diskUsed": 21474836480,
    "diskTotal": 107374182400,
    "netRx": 0,
    "netTx": 0
  }
}
```

Response:

```json
{
  "ok": true,
  "serverId": 12,
  "commands": [
    { "id": 7, "type": "shell", "params": { "cmd": "systemctl restart nginx" } }
  ]
}
```

### `POST /api/agent/command/:id/result`

```json
{ "status": "succeeded", "result": { "exitCode": 0, "stdout": "…", "stderr": "…" }, "error": "" }
```

### `POST /api/agent/terminal/:sessionId/io`

Each poll sends newly captured PTY output and receives any input to feed the PTY.

Request:

```json
{ "output": "…", "exitCode": null, "closed": false }
```

Response:

```json
{ "input": "…", "resize": { "rows": 24, "cols": 80 }, "close": false }
```

## What the hub can ask the agent to do

Commands arrive in the `commands` array of a heartbeat response. Each has a `type`:

### `shell`

Run a shell command (`/bin/sh -c`) and return its exit code, stdout and stderr.

```json
{ "id": 7, "type": "shell", "params": { "cmd": "systemctl restart nginx" } }
```

### `deploy`

Apply a policy-derived spec: install packages, write files, write environment values.

```json
{
  "id": 8,
  "type": "deploy",
  "params": {
    "packages": ["nginx"],
    "files": [{ "path": "/etc/nginx/conf.d/app.conf", "content": "server { … }" }],
    "env": [{ "key": "DB_PASSWORD", "value": "…" }]
  }
}
```

### `terminal`

Open an interactive PTY (`/bin/bash`, falling back to `/bin/sh`). Unlike a one-shot
command, the session is streamed over `/api/agent/terminal/:sessionId/io`; the hub
passes a `sessionId` in the command params.

## Trust and safety

- The agent **runs as root** and executes whatever the hub asks. Enroll it only on
  servers you control and trust the hub to manage.
- The agent **opens no inbound port** — no SSH exposure, no firewall hole.
- All traffic is HTTPS. Enrollment tokens are single-use; agent keys are per-chamber.
