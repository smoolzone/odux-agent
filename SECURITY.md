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

In **full** mode (the default) the `odux` agent executes commands as root and trusts the hub it
is enrolled with: treat a compromised hub as a compromised server. In **monitor** mode it
refuses every command and only reports machine stats, so a compromised hub learns those stats
and nothing more. The mode is set on the server (`sudo odux agent mode monitor`), never by the
hub.

The transport is TLS-only: the agent refuses a plain `http://` hub address (except
`localhost`), and every call to the hub times out after 30 seconds. The agent opens no
inbound port.
