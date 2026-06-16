#!/usr/bin/env bash
set -euo pipefail

BIN_NAME="teer"
INSTALL_DIR="/usr/local/bin"
APP_DEST="/Applications/Teer.app"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[teer]${NC} $*"; }
warn()  { echo -e "${YELLOW}[teer]${NC} $*"; }
error() { echo -e "${RED}[teer]${NC} $*" >&2; exit 1; }

desktop_dir() {
  local dir="$HOME/Desktop"
  if [ -f "$HOME/.config/user-dirs.dirs" ]; then
    local val
    val="$(grep "^XDG_DESKTOP_DIR" "$HOME/.config/user-dirs.dirs" 2>/dev/null | cut -d= -f2- | tr -d '"' | sed "s|\$HOME|$HOME|")"
    [ -n "$val" ] && dir="$val"
  fi
  echo "$dir"
}

remove_path() {
  local path="$1"
  if [ ! -e "$path" ] && [ ! -L "$path" ]; then
    return 0
  fi
  if [ -w "$(dirname "$path")" ] 2>/dev/null || [ -w "$path" ] 2>/dev/null; then
    rm -rf "$path"
  else
    sudo rm -rf "$path"
  fi
  info "Dihapus: $path"
}

uninstall_linux() {
  local data_home="${XDG_DATA_HOME:-$HOME/.local/share}"
  local apps_dir="$data_home/applications"
  local icons_dir="$data_home/icons/hicolor/128x128/apps"
  local user_desktop
  user_desktop="$(desktop_dir)"

  remove_path "${INSTALL_DIR}/${BIN_NAME}"
  remove_path "$apps_dir/teer.desktop"
  remove_path "$user_desktop/teer.desktop"
  remove_path "$icons_dir/teer.png"

  if command -v update-desktop-database >/dev/null 2>&1 && [ -d "$apps_dir" ]; then
    update-desktop-database "$apps_dir" >/dev/null 2>&1 || true
  fi
}

uninstall_macos() {
  local user_desktop
  user_desktop="$(desktop_dir)"
  local desktop_link="$user_desktop/Teer.app"

  if [ -L "$desktop_link" ] && [ "$(readlink "$desktop_link")" = "$APP_DEST" ]; then
    remove_path "$desktop_link"
  elif [ -e "$desktop_link" ]; then
    warn "Lewati $desktop_link (bukan symlink dari installer Teer)"
  fi

  remove_path "${INSTALL_DIR}/${BIN_NAME}"
  remove_path "$APP_DEST"
}

purge_config() {
  local config_dir="${XDG_CONFIG_HOME:-$HOME/.config}/teer"
  remove_path "$config_dir"
}

OS="$(uname -s)"
case "$OS" in
  Linux)  uninstall_linux ;;
  Darwin) uninstall_macos ;;
  *)      error "Unsupported OS: $OS. Use uninstall.ps1 on Windows." ;;
esac

if [ "${TEER_PURGE_CONFIG:-}" = "1" ]; then
  purge_config
else
  warn "Config di ~/.config/teer tidak dihapus. Set TEER_PURGE_CONFIG=1 untuk ikut menghapus."
fi

info "Teer berhasil di-uninstall."
