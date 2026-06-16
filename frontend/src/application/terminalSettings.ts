// Lapisan aplikasi: store & dialog pengaturan tampilan terminal (FR-20).
// Nilai disimpan di localStorage sehingga persisten antar restart.

import { writable, derived } from "svelte/store";
import {
  THEMES,
  DEFAULT_TERMINAL_SETTINGS,
  type TerminalTheme,
  type TerminalSettings,
} from "@domain/terminalSettings";

const LS_KEY = "teer:terminalSettings";

function load(): TerminalSettings {
  try {
    const raw = localStorage.getItem(LS_KEY);
    if (raw) return { ...DEFAULT_TERMINAL_SETTINGS, ...JSON.parse(raw) };
  } catch {
    /* abaikan */
  }
  return { ...DEFAULT_TERMINAL_SETTINGS };
}

const stored = load();

export const terminalFontSize = writable(stored.fontSize);
export const terminalFontFamily = writable(stored.fontFamily);
export const terminalThemeName = writable(stored.themeName);

export const terminalTheme = derived(
  terminalThemeName,
  ($name): TerminalTheme => THEMES.find((t) => t.name === $name) ?? THEMES[0],
);

// Persist ke localStorage saat ada perubahan.
derived(
  [terminalFontSize, terminalFontFamily, terminalThemeName],
  ([$fontSize, $fontFamily, $themeName]): TerminalSettings => ({
    fontSize: $fontSize,
    fontFamily: $fontFamily,
    themeName: $themeName,
  }),
).subscribe((v) => {
  try {
    localStorage.setItem(LS_KEY, JSON.stringify(v));
  } catch {
    /* abaikan */
  }
});

// ---- Dialog state ----

interface TerminalSettingsDialogState {
  resolve: () => void;
}

export const terminalSettingsDialog =
  writable<TerminalSettingsDialogState | null>(null);

export function openTerminalSettings(): Promise<void> {
  return new Promise((resolve) => {
    terminalSettingsDialog.set({ resolve });
  });
}
