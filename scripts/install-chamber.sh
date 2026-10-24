#!/bin/sh
# Odux chamber installer — one line installs the CLI, enrolls the agent, and
# installs the systemd service.
#
#   curl -fsSL https://odux.io/install-chamber | sudo sh -s -- --token en_… --hub https://odux.io
#
# Every artifact is verified against SHA256SUMS, whose signature (SHA256SUMS.sig)
# is checked with the release public key embedded below — a compromised mirror
# cannot deliver tampered code.
#
# Environment (optional):
#   ODUX_VERSION            version to install (default: latest)
#   ODUX_BASE_URL           base URL for artifacts (default: https://odux.io/downloads)
#   ODUX_INSTALL_DIR        install directory     (default: /usr/local/bin)
#   ODUX_ALLOW_UNSIGNED     set 1 to permit an unsigned release (NOT recommended)
set -e

BASE_URL="${ODUX_BASE_URL:-https://github.com/smoolzone/odux-agent/releases/latest/download}"
DEST="${ODUX_INSTALL_DIR:-/usr/local/bin}"
VERSION="${ODUX_VERSION:-latest}"
TOKEN=""
HUB=""

# Release signing public key (ECDSA P-256) — see scripts/keys/README.md.
ODUX_PUBKEY='-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEuBJZUDO61Yr9uSTeJwDt1AOYvDGX
gTUq1yz/jrlQbGdmuxwiB2tDJALbTP844SNZhbrFqTkUuKMe67Z0EPpPag==
-----END PUBLIC KEY-----'

while [ $# -gt 0 ]; do
  case "$1" in
    --token) TOKEN="$2"; shift 2 ;;
    --hub) HUB="$2"; shift 2 ;;
    --base-url) BASE_URL="$2"; shift 2 ;;
    *) echo "odux installer: unknown argument '$1'" >&2; exit 2 ;;
  esac
done

if [ -z "$TOKEN" ]; then
  echo "odux installer: --token is required (e.g. --token en_…)" >&2
  exit 2
fi
HUB="${HUB:-https://odux.io}"

# --- platform ---------------------------------------------------------------
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "odux installer: unsupported architecture '$arch'" >&2; exit 1 ;;
esac
case "$os" in
  linux|darwin) : ;;
  *) echo "odux installer: unsupported OS '$os' (need linux or darwin)" >&2; exit 1 ;;
esac

# Enroll writes /etc/odux/agent.yaml and install writes a systemd unit → root.
if [ "$(id -u)" != "0" ]; then
  echo "odux installer: run as root, e.g.  … | sudo sh -s -- --token … --hub …" >&2
  exit 1
fi

# Resolve the version (the mirror publishes a plain-text `latest` file).
if [ "$VERSION" = "latest" ]; then
  VERSION="$(curl -fsSL "$BASE_URL/latest" 2>/dev/null | head -n 1 | tr -d '[:space:]' || true)"
  [ -n "$VERSION" ] || { echo "odux installer: could not resolve the latest version" >&2; exit 1; }
fi

asset="odux_${VERSION}_${os}_${arch}.tar.gz"
bin="$DEST/odux"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT HUP INT TERM

echo "odux installer: downloading $BASE_URL/$asset"
curl -fsSL "$BASE_URL/$asset" -o "$tmp/$asset"

# --- verify (signature -> checksum) -----------------------------------------
sums="$(curl -fsSL "$BASE_URL/SHA256SUMS" 2>/dev/null || true)"
[ -n "$sums" ] || { echo "odux installer: could not fetch SHA256SUMS" >&2; exit 1; }
printf '%s\n' "$sums" > "$tmp/SHA256SUMS"

if curl -fsSL "$BASE_URL/SHA256SUMS.sig" -o "$tmp/SHA256SUMS.sig" 2>/dev/null; then
  printf '%s\n' "$ODUX_PUBKEY" > "$tmp/odux-release.pub"
  if openssl dgst -sha256 -verify "$tmp/odux-release.pub" \
       -signature "$tmp/SHA256SUMS.sig" "$tmp/SHA256SUMS" >/dev/null 2>&1; then
    echo "odux installer: signature OK"
  else
    echo "odux installer: SIGNATURE INVALID — refusing to install" >&2
    exit 1
  fi
elif [ "${ODUX_ALLOW_UNSIGNED:-}" = "1" ]; then
  echo "odux installer: WARNING — installing UNSIGNED release (ODUX_ALLOW_UNSIGNED=1)" >&2
else
  echo "odux installer: UNSIGNED RELEASE — SHA256SUMS.sig missing; refusing to install" >&2
  echo "  (possible mirror tampering; override with ODUX_ALLOW_UNSIGNED=1 at your own risk)" >&2
  exit 1
fi

expected="$(printf '%s\n' "$sums" | awk -v a="$asset" '$2 == a { print $1; exit }')"
[ -n "$expected" ] || { echo "odux installer: no checksum found for ${asset}" >&2; exit 1; }
if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$tmp/$asset" | awk '{print $1}')"
else
  actual="$(shasum -a 256 "$tmp/$asset" | awk '{print $1}')"
fi
[ "$actual" = "$expected" ] || {
  echo "odux installer: checksum mismatch" >&2
  echo "  expected: ${expected}" >&2
  echo "  got:      ${actual}" >&2
  exit 1
}
echo "odux installer: checksum OK"

tar -xzf "$tmp/$asset" -C "$tmp" odux
mkdir -p "$DEST"
install -m 0755 "$tmp/odux" "$bin"
echo "odux installer: installed $bin"

echo "odux installer: enrolling agent…"
"$bin" agent enroll --token "$TOKEN" --hub "$HUB"

echo "odux installer: installing the agent service…"
"$bin" agent install

echo
echo "Odux agent is enrolled and running (systemd service: odux-agent)."
