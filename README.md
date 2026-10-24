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

## How it works

See [`PROTOCOL.md`](PROTOCOL.md) for the exact requests the agent makes and the
commands it accepts (`shell`, `deploy`, `terminal`).

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

