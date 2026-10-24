# Security policy

## Reporting a vulnerability

Please do **not** open a public issue. Report privately via either:

- GitHub **private vulnerability reporting** (Security → Report a vulnerability), or
- email to **security@odux.io**.

You'll get a first reply within **3 working days**. Include the affected version, your
OS/architecture, and as much reproduction detail as you can.

## Supported versions

Only the **latest release** is patched. If you're on an older version, upgrade first.

## The agent runs as root

The `odux` agent executes commands as root and trusts the hub it is enrolled with.
Treat a compromised hub as a compromised server. The agent's threat model is: the hub
is trusted, the transport is TLS-only, and the agent opens no inbound port.
