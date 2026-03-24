#!/usr/bin/env bash
#
# Go-Beats Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/rolniuq/go-beats/main/install.sh | bash
#
set -euo pipefail

REPO="rolniuq/go-beats"
APP_NAME="Go-Beats"
INSTALL_DIR="/Applications"

info()  { printf "\033[1;34m==> %s\033[0m\n" "$*"; }
ok()    { printf "\033[1;32m==> %s\033[0m\n" "$*"; }
err()   { printf "\033[1;31mError: %s\033[0m\n" "$*" >&2; exit 1; }

# ── Detect OS ────────────────────────────────────────────────────────────

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Darwin) ;;
  *) err "Go-Beats desktop app is only available for macOS. For Linux/Windows, use: go install github.com/${REPO}/cmd/go-beats@latest" ;;
esac

case "$ARCH" in
  arm64)  ARCH_LABEL="arm64" ;;
  x86_64) ARCH_LABEL="amd64" ;;
  *)      err "Unsupported architecture: $ARCH" ;;
esac

# ── Fetch latest release ─────────────────────────────────────────────────

info "Detecting latest Go-Beats release..."

RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
RELEASE_JSON="$(curl -fsSL "$RELEASE_URL")" || err "Failed to fetch release info. Check your internet connection."

TAG="$(echo "$RELEASE_JSON" | grep '"tag_name"' | head -1 | sed 's/.*: *"//;s/".*//')"
[ -n "$TAG" ] || err "Could not determine latest release tag."

info "Latest version: $TAG"

# ── Download .app bundle ─────────────────────────────────────────────────

ASSET_NAME="Go-Beats_${TAG#v}_darwin_${ARCH_LABEL}.app.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET_NAME}"

TMPDIR_INSTALL="$(mktemp -d)"
trap 'rm -rf "$TMPDIR_INSTALL"' EXIT

info "Downloading ${ASSET_NAME}..."
curl -fSL --progress-bar -o "${TMPDIR_INSTALL}/${ASSET_NAME}" "$DOWNLOAD_URL" \
  || err "Failed to download ${DOWNLOAD_URL}. This release may not have a desktop build yet."

# ── Extract and install ──────────────────────────────────────────────────

info "Extracting..."
tar -xzf "${TMPDIR_INSTALL}/${ASSET_NAME}" -C "$TMPDIR_INSTALL"

# Find the .app bundle in extracted files
APP_BUNDLE="$(find "$TMPDIR_INSTALL" -maxdepth 2 -name '*.app' -type d | head -1)"
[ -n "$APP_BUNDLE" ] || err "Could not find .app bundle in archive."

info "Installing ${APP_NAME}.app to ${INSTALL_DIR}..."
if [ -d "${INSTALL_DIR}/${APP_NAME}.app" ]; then
  rm -rf "${INSTALL_DIR}/${APP_NAME}.app"
fi
cp -R "$APP_BUNDLE" "${INSTALL_DIR}/${APP_NAME}.app"

# ── Remove quarantine (downloaded from internet) ─────────────────────────

xattr -rd com.apple.quarantine "${INSTALL_DIR}/${APP_NAME}.app" 2>/dev/null || true

# ── Done ─────────────────────────────────────────────────────────────────

ok "Go-Beats has been installed to ${INSTALL_DIR}/${APP_NAME}.app"
ok "You can find it in Launchpad, Spotlight, or run: open -a Go-Beats"
echo ""
echo "  To uninstall:  rm -rf /Applications/Go-Beats.app"
echo ""
