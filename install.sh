#!/bin/sh
set -eu

REPO="sporkops/cli"
BINARY="spork"
INSTALL_DIR="/usr/local/bin"
COSIGN_PUBKEY_URL="https://raw.githubusercontent.com/${REPO}/main/cosign.pub"

err() {
    echo "Error: $*" >&2
    exit 1
}

require_cmd() {
    command -v "$1" >/dev/null 2>&1 || err "required tool '$1' is not installed"
}

main() {
    require_cmd curl
    require_cmd tar
    require_cmd uname
    require_cmd mktemp
    require_cmd grep
    require_cmd awk

    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)

    case "$os" in
        linux)  os="linux" ;;
        darwin) os="darwin" ;;
        *) err "unsupported OS: $os" ;;
    esac

    case "$arch" in
        x86_64|amd64)  arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) err "unsupported architecture: $arch" ;;
    esac

    # Resolve the latest release tag. curl -f turns HTTP errors into exit codes
    # so a 404 or 5xx does not silently degrade to an empty tag.
    release_json=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest") \
        || err "could not fetch latest release metadata from GitHub"
    latest=$(printf '%s' "$release_json" | grep '"tag_name"' | head -n 1 | sed -E 's/.*"v([^"]+)".*/\1/')
    [ -n "$latest" ] || err "could not parse latest release tag"

    archive="spork_${os}_${arch}.tar.gz"
    base_url="https://github.com/${REPO}/releases/download/v${latest}"
    echo "Downloading spork v${latest} for ${os}/${arch}..."

    tmpdir=$(mktemp -d)
    trap 'rm -rf "$tmpdir"' EXIT

    curl -fsSL "${base_url}/${archive}"      -o "${tmpdir}/${archive}"      || err "downloading ${archive}"
    curl -fsSL "${base_url}/checksums.txt"   -o "${tmpdir}/checksums.txt"   || err "downloading checksums.txt"

    verify_checksum "${tmpdir}" "${archive}"
    verify_cosign   "${tmpdir}"

    tar -xzf "${tmpdir}/${archive}" -C "$tmpdir"
    [ -f "${tmpdir}/${BINARY}" ] || err "archive did not contain ${BINARY}"

    if [ -w "$INSTALL_DIR" ]; then
        mv "$tmpdir/$BINARY" "$INSTALL_DIR/$BINARY"
    else
        echo "Installing to $INSTALL_DIR (requires sudo)..."
        sudo mv "$tmpdir/$BINARY" "$INSTALL_DIR/$BINARY"
    fi
    chmod +x "$INSTALL_DIR/$BINARY"

    echo ""
    echo "✓ spork installed successfully (v${latest})"
    echo ""
    echo "Get started:"
    echo "  spork login"
    echo "  spork monitor add https://yoursite.com"
    echo ""
    echo "Docs: https://sporkops.com/docs"
}

# verify_checksum requires the matching line in checksums.txt, refuses to
# proceed when no sha256 tool is available, and dies on mismatch. Previously
# this function would warn-and-continue in any of those cases, defeating the
# integrity check entirely.
verify_checksum() {
    dir="$1"
    archive="$2"

    if command -v sha256sum >/dev/null 2>&1; then
        sha_cmd="sha256sum"
    elif command -v shasum >/dev/null 2>&1; then
        sha_cmd="shasum -a 256"
    else
        err "no sha256 tool available (install coreutils or perl); refusing to install without integrity verification"
    fi

    expected=$(grep " ${archive}\$" "${dir}/checksums.txt" | awk '{print $1}')
    [ -n "$expected" ] || err "no checksum entry for ${archive} in checksums.txt; refusing to install without integrity verification"

    actual=$(${sha_cmd} "${dir}/${archive}" | awk '{print $1}')
    if [ "$actual" != "$expected" ]; then
        echo "  Expected: $expected" >&2
        echo "  Got:      $actual"   >&2
        err "checksum verification failed for ${archive}"
    fi
}

# verify_cosign checks a keyless cosign signature over checksums.txt when
# cosign is available and SPORK_SKIP_COSIGN is unset. The signature is
# optional for now (opt-in via the tool's presence) so installs still work on
# systems without cosign; set SPORK_REQUIRE_COSIGN=1 in hardened environments
# to make it mandatory.
verify_cosign() {
    dir="$1"
    if [ "${SPORK_SKIP_COSIGN:-0}" = "1" ]; then
        return 0
    fi
    if ! command -v cosign >/dev/null 2>&1; then
        if [ "${SPORK_REQUIRE_COSIGN:-0}" = "1" ]; then
            err "SPORK_REQUIRE_COSIGN=1 but cosign is not installed"
        fi
        return 0
    fi

    # Prefer the release's own .sig/.pem over the repo's pinned key so we
    # can rotate signing keys without re-releasing the installer. The CI
    # pipeline uploads these alongside checksums.txt.
    sig_url="${base_url}/checksums.txt.sig"
    cert_url="${base_url}/checksums.txt.pem"

    if curl -fsSL "$sig_url"  -o "${dir}/checksums.txt.sig"  2>/dev/null \
    && curl -fsSL "$cert_url" -o "${dir}/checksums.txt.pem"  2>/dev/null; then
        echo "Verifying release signature with cosign..."
        COSIGN_EXPERIMENTAL=1 cosign verify-blob \
            --certificate "${dir}/checksums.txt.pem" \
            --signature   "${dir}/checksums.txt.sig" \
            --certificate-identity-regexp "^https://github\\.com/${REPO}/\\.github/workflows/release\\.yml@" \
            --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
            "${dir}/checksums.txt" >/dev/null 2>&1 \
            || err "cosign signature verification failed"
    elif [ "${SPORK_REQUIRE_COSIGN:-0}" = "1" ]; then
        err "SPORK_REQUIRE_COSIGN=1 but this release has no cosign signature"
    fi
}

main
