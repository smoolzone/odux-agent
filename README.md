# Odux Agent

The Odux agent and `odux` CLI — one Go binary that runs on your servers, phones home
to the Odux hub over HTTPS, and executes the commands the hub sends it.

- **No inbound access.** The agent opens no port; it polls the hub every 5 seconds.
- **One binary.** `odux` is both the agent (`odux agent run`) and the operator CLI.
- **Readable before you trust it.** This runs as root on your servers — every line is here.

## Install

```bash
curl -fsSL https://odux.io/install | sh
```

Build from source (Go 1.22+):

```bash
make build      # -> ./odux
go test ./...
```

## Verify a release

Releases are signed with the Odux release key. Verify `SHA256SUMS` before trusting a
downloaded binary:

```bash
curl -fsSLO https://github.com/smoolzone/odux-agent/releases/latest/download/SHA256SUMS
curl -fsSLO https://github.com/smoolzone/odux-agent/releases/latest/download/SHA256SUMS.sig
openssl dgst -sha256 -verify scripts/keys/odux-release.pub -signature SHA256SUMS.sig SHA256SUMS
```

## Two modes: you choose what the agent may do

| Mode | What it does |
| --- | --- |
| **full** (default) | Reports machine stats and runs the commands the hub sends: `shell`, `deploy`, `terminal`. |
| **monitor** | Reports machine stats only. Every command from the hub is refused, and the hub is told why. |

Pick it when you install, or change it any time on the server:

```bash
curl -fsSL https://odux.io/install-chamber | sudo sh -s -- --token en_… --hub https://odux.io --mode monitor
sudo odux agent mode            # show the current mode
sudo odux agent mode full       # allow commands again (takes effect on the next heartbeat)
```

The mode is stored on the server, in a root-owned file (`/etc/odux/agent.yaml`). The hub
cannot change it: only someone with root on that machine can.

## What the agent sends

- **Every 5 seconds:** hostname, agent version, OS and architecture, and a machine snapshot:
  CPU %, load average, RAM used/total, disk used/total, network totals.
- **Only when you run something from the hub** (full mode): that command's exit code and
  output, or the output of a terminal session you opened.
- **Nothing else.** No files, logs, environment variables or process list. It opens no
  port, and it only talks to your hub, over HTTPS: a plain `http://` hub address is refused
  (except `localhost`, for development).

Every request it can make is listed in [`PROTOCOL.md`](PROTOCOL.md).

## Relationship to the Odux hub

The agent is open source (Apache-2.0) so you can read exactly what runs on your
servers. The [Odux hub](https://odux.io) — the control plane, panel and billing — is a
separate commercial product.

## Trademark

"Odux" is a trademark. Apache-2.0 grants you the code but no trademark rights: forks
and derivatives must use a different name and logo.

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md). For security issues, see
[`SECURITY.md`](SECURITY.md).

