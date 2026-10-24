#!/bin/sh
# Odux CLI installer — downloads, verifies (signature + SHA-256), and installs
# the `odux` binary.
#
#   curl -fsSL https://odux.io/downloads/install-cli.sh | sh
#
# (The hub installer — "Odux itself" in local mode — is https://odux.io/install.)
#
# Artifacts are served from the public mirror (odux.io/downloads), populated only
# from the signed GitHub release. Every artifact is verified against SHA256SUMS,
# whose signature (SHA256SUMS.sig) is checked with the release public key
# embedded below — so a compromised mirror cannot deliver tampered code.
#
# Environment (all optional):
#   ODUX_VERSION            version to install, e.g. 0.1.3  (default: latest)
#   ODUX_INSTALL_DIR        install directory  (default: /usr/local/bin as root, else ~/.odux/bin)
#   ODUX_BASE_URL           override the artifact base URL  (mirrors / local testing)
#   ODUX_ALLOW_UNSIGNED     set 1 to permit an unsigned release (NOT recommended)
set -e

REPO="${ODUX_REPO:-smoolzone/odux-agent}"
DEFAULT_BASE="https://github.com/smoolzone/odux-agent/releases"
BASE_URL="${ODUX_BASE_URL:-$DEFAULT_BASE}"
VERSION="${ODUX_VERSION:-latest}"

# Release signing public key (ECDSA P-256) — see scripts/keys/README.md.
ODUX_PUBKEY='-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEuBJZUDO61Yr9uSTeJwDt1AOYvDGX
gTUq1yz/jrlQbGdmuxwiB2tDJALbTP844SNZhbrFqTkUuKMe67Z0EPpPag==
-----END PUBLIC KEY-----'

# --- platform ---------------------------------------------------------------
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$os" in
  linux|darwin) : ;;
  *) echo "odux installer: unsupported OS '$os' (need linux or darwin)" >&2; exit 1 ;;
esac
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "odux installer: unsupported architecture '$arch'" >&2; exit 1 ;;
esac

# --- resolve the version ----------------------------------------------------
if [ "$VERSION" = "latest" ]; then
  echo "odux installer: resolving the latest release…"
  # A mirror serves a plain-text `latest` file; GitHub needs the API (its
  # `/releases/latest` URL returns HTML). Only accept a value that looks like a version.
  VERSION="$(curl -fsSL "${BASE_URL}/latest" 2>/dev/null | head -n 1 | tr -d '[:space:]' || true)"
  case "$VERSION" in
    v[0-9]*|[0-9]*) : ;;
    *) VERSION="" ;;
  esac
  if [ -z "$VERSION" ]; then
    VERSION="$(
      curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep -o '"tag_name": *"[^"]*"' | head -n 1 | cut -d '"' -f 4
    )"
  fi
  [ -n "$VERSION" ] || { echo "odux installer: could not resolve the latest release" >&2; exit 1; }
fi
case "$VERSION" in
  v*) : ;;
  *) VERSION="v${VERSION}" ;;
esac

asset="odux_${VERSION#v}_${os}_${arch}.tar.gz"
# GitHub releases serve under /download/<tag>/; the mirror serves flat.
case "$BASE_URL" in
  *github.com*) REL_BASE="${BASE_URL}/download/${VERSION}" ;;
  *)            REL_BASE="${BASE_URL}" ;;
esac
url="${REL_BASE}/${asset}"
sums_url="${REL_BASE}/SHA256SUMS"
sig_url="${REL_BASE}/SHA256SUMS.sig"

# --- destination ------------------------------------------------------------
if [ -n "$ODUX_INSTALL_DIR" ]; then
  dest="$ODUX_INSTALL_DIR"
elif [ "$(id -u)" = "0" ]; then
  dest="/usr/local/bin"
else
  dest="$HOME/.odux/bin"
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT HUP INT TERM

echo "odux installer: downloading ${url}"
curl -fsSL "$url" -o "$tmpdir/${asset}"

# --- signature verification (provenance) ------------------------------------
# SHA256SUMS is signed in CI; verifying it against the embedded public key makes
# the checksum trustworthy even if the download mirror was tampered with.
sums="$(curl -fsSL "$sums_url" 2>/dev/null || true)"
[ -n "$sums" ] || { echo "odux installer: could not fetch SHA256SUMS" >&2; exit 1; }
printf '%s\n' "$sums" > "$tmpdir/SHA256SUMS"

if curl -fsSL "$sig_url" -o "$tmpdir/SHA256SUMS.sig" 2>/dev/null; then
  printf '%s\n' "$ODUX_PUBKEY" > "$tmpdir/odux-release.pub"
  if openssl dgst -sha256 -verify "$tmpdir/odux-release.pub" \
       -signature "$tmpdir/SHA256SUMS.sig" "$tmpdir/SHA256SUMS" >/dev/null 2>&1; then
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

# --- SHA-256 verification ---------------------------------------------------
echo "odux installer: verifying SHA-256 checksum…"
expected="$(printf '%s\n' "$sums" | awk -v a="$asset" '$2 == a { print $1; exit }')"
[ -n "$expected" ] || { echo "odux installer: no checksum found for ${asset}" >&2; exit 1; }
if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$tmpdir/${asset}" | awk '{print $1}')"
else
  actual="$(shasum -a 256 "$tmpdir/${asset}" | awk '{print $1}')"
fi
[ "$actual" = "$expected" ] || {
  echo "odux installer: checksum mismatch" >&2
  echo "  expected: ${expected}" >&2
  echo "  got:      ${actual}" >&2
  exit 1
}
echo "odux installer: checksum OK"

# --- install ----------------------------------------------------------------
mkdir -p "$dest"
tar -xzf "$tmpdir/${asset}" -C "$dest" odux
chmod 0755 "$dest/odux"

echo
echo "odux ${VERSION} installed to ${dest}/odux"
case ":$PATH:" in
  *":${dest}:"*) : ;;
  *) echo "hint: add ${dest} to your PATH, e.g.:  export PATH=\"${dest}:\$PATH\"" ;;
esac
echo "next:  odux doctor  ·  odux init  ·  odux castle"
