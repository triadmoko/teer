#!/usr/bin/env bash
set -euo pipefail

REPO="triadmoko/teer"
BIN_NAME="teer"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
LEGACY_DIR="/usr/local/bin"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()    { echo -e "${GREEN}[teer]${NC} $*"; }
warn()    { echo -e "${YELLOW}[teer]${NC} $*"; }
error()   { echo -e "${RED}[teer]${NC} $*" >&2; exit 1; }

download_file() {
  local url="$1" dest="$2"
  if command -v curl &>/dev/null; then
    curl -fsSL "$url" -o "$dest"
  else
    wget -qO "$dest" "$url"
  fi
}

desktop_dir() {
  local dir="$HOME/Desktop"
  if [ -f "$HOME/.config/user-dirs.dirs" ]; then
    local val
    val="$(grep "^XDG_DESKTOP_DIR" "$HOME/.config/user-dirs.dirs" 2>/dev/null | cut -d= -f2- | tr -d '"' | sed "s|\$HOME|$HOME|")"
    [ -n "$val" ] && dir="$val"
  fi
  echo "$dir"
}

install_linux_desktop() {
  local exec_path="${INSTALL_DIR}/${BIN_NAME}"
  local data_home="${XDG_DATA_HOME:-$HOME/.local/share}"
  local apps_dir="$data_home/applications"
  local icons_dir="$data_home/icons/hicolor/128x128/apps"
  local desktop_file="$apps_dir/teer.desktop"
  local icon_path="$icons_dir/teer.png"

  mkdir -p "$apps_dir" "$icons_dir"

  if ! download_file "https://raw.githubusercontent.com/${REPO}/${VERSION}/build/appicon.png" "$icon_path" 2>/dev/null; then
    download_file "https://raw.githubusercontent.com/${REPO}/main/build/appicon.png" "$icon_path" 2>/dev/null \
      || warn "Ikon desktop tidak dapat diunduh; launcher tetap dibuat tanpa ikon kustom."
  fi

  cat > "$desktop_file" <<EOF
[Desktop Entry]
Type=Application
Name=Teer
Comment=Terminal Workspace Manager
Exec=${exec_path}
Icon=${icon_path}
Categories=Development;Utility;
Terminal=false
StartupNotify=true
StartupWMClass=teer
Version=1.0
EOF

  chmod +x "$desktop_file"

  if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "$apps_dir" >/dev/null 2>&1 || true
  fi

  local user_desktop
  user_desktop="$(desktop_dir)"
  if [ -d "$user_desktop" ]; then
    cp "$desktop_file" "$user_desktop/teer.desktop"
    chmod +x "$user_desktop/teer.desktop"
    if command -v gio >/dev/null 2>&1; then
      gio set "$user_desktop/teer.desktop" metadata::trusted true 2>/dev/null || true
    fi
    info "Shortcut desktop: $user_desktop/teer.desktop"
  else
    warn "Folder desktop tidak ditemukan; launcher hanya di $desktop_file"
  fi
}

install_macos_desktop() {
  local app_dest="/Applications/Teer.app"
  local user_desktop
  user_desktop="$(desktop_dir)"
  if [ -d "$user_desktop" ]; then
    ln -sf "$app_dest" "$user_desktop/Teer.app"
    info "Shortcut desktop: $user_desktop/Teer.app"
  fi
}

# --- parse args (alternatif dari env TEER_VERSION; aman saat di-pipe via `bash -s --`) ---
while [ $# -gt 0 ]; do
  case "$1" in
    -v|--version)
      [ -n "${2:-}" ] || error "Opsi $1 butuh argumen versi, mis. --version v0.1.0"
      TEER_VERSION="$2"
      shift 2
      ;;
    --version=*)
      TEER_VERSION="${1#*=}"
      shift
      ;;
    -h|--help)
      echo "Usage: install.sh [--version <tag>]"
      echo "  --version, -v <tag>   Pasang versi tertentu (mis. v0.1.0)"
      echo "  Bisa juga lewat env: TEER_VERSION=v0.1.0"
      exit 0
      ;;
    *)
      error "Argumen tidak dikenal: $1"
      ;;
  esac
done

# --- detect OS/arch ---
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  OS_KEY="linux" ;;
  Darwin) OS_KEY="darwin" ;;
  *)      error "Unsupported OS: $OS. Use install.ps1 on Windows." ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH_KEY="amd64" ;;
  aarch64|arm64) ARCH_KEY="arm64" ;;
  *) error "Unsupported architecture: $ARCH" ;;
esac

if [ "$OS_KEY" = "darwin" ]; then
  ASSET="${BIN_NAME}-macos-universal.zip"
  IS_ZIP=true
else
  ASSET="${BIN_NAME}-${OS_KEY}-${ARCH_KEY}"
  IS_ZIP=false
fi

# --- resolve version ---
if [ -n "${TEER_VERSION:-}" ]; then
  VERSION="$TEER_VERSION"
  info "Installing version: $VERSION"
else
  info "Fetching latest release..."
  if command -v curl &>/dev/null; then
    VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
      | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
  elif command -v wget &>/dev/null; then
    VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" \
      | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
  else
    error "curl or wget required"
  fi
  info "Latest version: $VERSION"
fi

[ -z "$VERSION" ] && error "Failed to determine version. Check your internet connection."

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

# --- download ---
TMP_FILE="$(mktemp)"
trap 'rm -f "$TMP_FILE"' EXIT

info "Downloading $ASSET from $DOWNLOAD_URL ..."
if command -v curl &>/dev/null; then
  curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"
else
  wget -qO "$TMP_FILE" "$DOWNLOAD_URL"
fi

if [ "$IS_ZIP" = "true" ]; then
  command -v unzip &>/dev/null || error "unzip is required for macOS install. Run: brew install unzip"
  TMP_DIR="$(mktemp -d)"
  trap 'rm -rf "$TMP_FILE" "$TMP_DIR"' EXIT
  unzip -q "$TMP_FILE" -d "$TMP_DIR"
  APP_SRC="$TMP_DIR/Teer.app"
  APP_DEST="/Applications/Teer.app"
  if [ -d "$APP_DEST" ]; then
    warn "Replacing existing $APP_DEST"
    rm -rf "$APP_DEST"
  fi
  cp -R "$APP_SRC" "$APP_DEST"
  # symlink binary to PATH
  if [ -w "$INSTALL_DIR" ]; then
    ln -sf "/Applications/Teer.app/Contents/MacOS/teer" "${INSTALL_DIR}/${BIN_NAME}"
  else
    sudo ln -sf "/Applications/Teer.app/Contents/MacOS/teer" "${INSTALL_DIR}/${BIN_NAME}"
  fi
  info "Installed Teer.app to /Applications"
  info "Binary symlinked at ${INSTALL_DIR}/${BIN_NAME}"
  install_macos_desktop
else
  chmod +x "$TMP_FILE"
  mkdir -p "$INSTALL_DIR"
  # Hapus binary lama (termasuk di lokasi legacy) agar tidak "Text file busy"
  rm -f "${INSTALL_DIR}/${BIN_NAME}"
  [ -f "${LEGACY_DIR}/${BIN_NAME}" ] && sudo rm -f "${LEGACY_DIR}/${BIN_NAME}" && \
    info "Binary lama di $LEGACY_DIR dihapus (migrasi ke $INSTALL_DIR)"
  mv "$TMP_FILE" "${INSTALL_DIR}/${BIN_NAME}"
  info "Installed at ${INSTALL_DIR}/${BIN_NAME}"

  # Pastikan INSTALL_DIR ada di PATH
  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    for rc in "$HOME/.bashrc" "$HOME/.zshrc"; do
      if [ -f "$rc" ] && ! grep -qF "$INSTALL_DIR" "$rc"; then
        echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$rc"
        info "PATH ditambahkan ke $rc"
      fi
    done
    warn "PATH diperbarui. Buka terminal baru atau jalankan: source ~/.bashrc"
  fi

  install_linux_desktop
fi

# --- Linux: install dependencies hint ---
if [ "$OS_KEY" = "linux" ]; then
  if ! ldconfig -p 2>/dev/null | grep -q "libwebkitgtk-6\|libwebkit2gtk-6"; then
    warn "Missing system dependency. Run:"
    warn "  sudo apt install libwebkitgtk-6.0-0  # Debian/Ubuntu"
    warn "  sudo dnf install webkitgtk6.0         # Fedora"
  fi
fi

info "Done! Run: teer"
