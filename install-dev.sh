#!/usr/bin/env bash
# Install Teer dari source lokal (untuk developer / contributor).
# Jalankan dari root repo: ./install-dev.sh
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_NAME="teer"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
LEGACY_DIR="/usr/local/bin"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info()  { echo -e "${GREEN}[teer]${NC} $*"; }
warn()  { echo -e "${YELLOW}[teer]${NC} $*"; }
error() { echo -e "${RED}[teer]${NC} $*" >&2; exit 1; }

# --- cek prasyarat ---
command -v go    >/dev/null 2>&1 || error "Go tidak ditemukan. Install dari https://go.dev/dl/"
command -v bun   >/dev/null 2>&1 || error "Bun tidak ditemukan. Install dari https://bun.sh/docs/installation"
command -v task  >/dev/null 2>&1 || error "Taskfile CLI tidak ditemukan. Install: go install github.com/go-task/task/v3/cmd/task@latest"
command -v wails3 >/dev/null 2>&1 || error "wails3 CLI tidak ditemukan. Install: go install github.com/wailsapp/wails/v3/cmd/wails3@latest"

# WebKit (Linux saja)
if [[ "$(uname -s)" == "Linux" ]]; then
  if ! ldconfig -p 2>/dev/null | grep -q "libwebkitgtk-6\|libwebkit2gtk-6"; then
    warn "WebKitGTK 6.0 tidak terdeteksi. Jalankan:"
    warn "  sudo apt install libgtk-4-dev libwebkitgtk-6.0-dev   # Debian/Ubuntu"
    warn "  sudo dnf install gtk4-devel webkitgtk6.0-devel        # Fedora"
    warn "Lanjutkan build tetap dijalankan..."
  fi
fi

# --- pindah ke root repo ---
cd "$REPO_DIR"

info "Build dari source: $REPO_DIR"

# --- build ---
task build

BIN_SRC="$REPO_DIR/bin/Teer"
[ -f "$BIN_SRC" ] || BIN_SRC="$REPO_DIR/bin/teer"
[ -f "$BIN_SRC" ] || error "Binary tidak ditemukan di bin/ setelah build. Cek output task build."

# --- install binary ---
mkdir -p "$INSTALL_DIR"
rm -f "$INSTALL_DIR/$BIN_NAME"

# Migrasi: hapus binary lama di /usr/local/bin jika ada
if [ -f "$LEGACY_DIR/$BIN_NAME" ]; then
  sudo rm -f "$LEGACY_DIR/$BIN_NAME"
  info "Binary lama di $LEGACY_DIR dihapus (migrasi ke $INSTALL_DIR)"
fi

cp "$BIN_SRC" "$INSTALL_DIR/$BIN_NAME"
chmod +x "$INSTALL_DIR/$BIN_NAME"
info "Binary dipasang di $INSTALL_DIR/$BIN_NAME"

# --- pastikan INSTALL_DIR ada di PATH ---
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  for rc in "$HOME/.bashrc" "$HOME/.zshrc"; do
    if [ -f "$rc" ]; then
      if ! grep -qF "$INSTALL_DIR" "$rc"; then
        echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$rc"
        info "PATH ditambahkan ke $rc"
      fi
    fi
  done
  warn "PATH diperbarui. Jalankan: source ~/.bashrc  (atau buka terminal baru)"
fi

# --- install desktop entry (Linux) ---
if [[ "$(uname -s)" == "Linux" ]]; then
  DATA_HOME="${XDG_DATA_HOME:-$HOME/.local/share}"
  APPS_DIR="$DATA_HOME/applications"
  ICONS_DIR="$DATA_HOME/icons/hicolor/128x128/apps"

  mkdir -p "$APPS_DIR" "$ICONS_DIR"

  # Salin ikon dari source
  cp "$REPO_DIR/build/appicon.png" "$ICONS_DIR/teer.png"

  cat > "$APPS_DIR/teer.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=Teer
Comment=Terminal Workspace Manager
Exec=$INSTALL_DIR/$BIN_NAME
Icon=$ICONS_DIR/teer.png
Categories=Development;Utility;
Terminal=false
StartupNotify=true
StartupWMClass=teer
Version=1.0
EOF
  chmod +x "$APPS_DIR/teer.desktop"

  command -v update-desktop-database >/dev/null 2>&1 \
    && update-desktop-database "$APPS_DIR" >/dev/null 2>&1 || true

  # Shortcut di Desktop
  DESKTOP_DIR="$HOME/Desktop"
  if [ -f "$HOME/.config/user-dirs.dirs" ]; then
    val="$(grep "^XDG_DESKTOP_DIR" "$HOME/.config/user-dirs.dirs" 2>/dev/null \
          | cut -d= -f2- | tr -d '"' | sed "s|\$HOME|$HOME|")"
    [ -n "$val" ] && DESKTOP_DIR="$val"
  fi
  if [ -d "$DESKTOP_DIR" ]; then
    cp "$APPS_DIR/teer.desktop" "$DESKTOP_DIR/teer.desktop"
    chmod +x "$DESKTOP_DIR/teer.desktop"
    command -v gio >/dev/null 2>&1 \
      && gio set "$DESKTOP_DIR/teer.desktop" metadata::trusted true 2>/dev/null || true
    info "Shortcut desktop: $DESKTOP_DIR/teer.desktop"
  fi

  info "Desktop entry: $APPS_DIR/teer.desktop"
fi

info "Selesai! Jalankan: teer"
