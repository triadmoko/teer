// Lapisan domain: tipe & preset pengaturan tampilan terminal (FR-20).

export interface TerminalTheme {
  name: string;
  background: string;
  foreground: string;
  cursor: string;
  selectionBackground: string;
  black: string;
  red: string;
  green: string;
  yellow: string;
  blue: string;
  magenta: string;
  cyan: string;
  white: string;
  brightBlack: string;
}

export const THEMES: readonly TerminalTheme[] = [
  {
    name: "Default Dark",
    background: "#121214", foreground: "#d4d4d8", cursor: "#e4e4e7",
    selectionBackground: "#3f3f46",
    black: "#18181b", red: "#f87171", green: "#4ade80", yellow: "#facc15",
    blue: "#60a5fa", magenta: "#c084fc", cyan: "#22d3ee", white: "#d4d4d8",
    brightBlack: "#52525b",
  },
  {
    name: "Dracula",
    background: "#282a36", foreground: "#f8f8f2", cursor: "#f8f8f2",
    selectionBackground: "#44475a",
    black: "#21222c", red: "#ff5555", green: "#50fa7b", yellow: "#f1fa8c",
    blue: "#bd93f9", magenta: "#ff79c6", cyan: "#8be9fd", white: "#f8f8f2",
    brightBlack: "#6272a4",
  },
  {
    name: "One Dark",
    background: "#282c34", foreground: "#abb2bf", cursor: "#528bff",
    selectionBackground: "#3e4451",
    black: "#282c34", red: "#e06c75", green: "#98c379", yellow: "#e5c07b",
    blue: "#61afef", magenta: "#c678dd", cyan: "#56b6c2", white: "#abb2bf",
    brightBlack: "#5c6370",
  },
  {
    name: "Solarized Dark",
    background: "#002b36", foreground: "#839496", cursor: "#839496",
    selectionBackground: "#073642",
    black: "#073642", red: "#dc322f", green: "#859900", yellow: "#b58900",
    blue: "#268bd2", magenta: "#d33682", cyan: "#2aa198", white: "#eee8d5",
    brightBlack: "#586e75",
  },
  {
    name: "Tokyo Night",
    background: "#1a1b2e", foreground: "#c0caf5", cursor: "#c0caf5",
    selectionBackground: "#283457",
    black: "#15161e", red: "#f7768e", green: "#9ece6a", yellow: "#e0af68",
    blue: "#7aa2f7", magenta: "#bb9af7", cyan: "#7dcfff", white: "#a9b1d6",
    brightBlack: "#414868",
  },
] as const;

export const FONT_FAMILIES: readonly string[] = [
  'ui-monospace, "JetBrains Mono", Menlo, monospace',
  '"Cascadia Code", ui-monospace, monospace',
  '"Fira Code", ui-monospace, monospace',
  '"Hack", ui-monospace, monospace',
  'monospace',
] as const;

export const FONT_FAMILY_LABELS: readonly string[] = [
  "JetBrains Mono",
  "Cascadia Code",
  "Fira Code",
  "Hack",
  "Monospace (sistem)",
] as const;

export const FONT_SIZES: readonly number[] = [10, 11, 12, 13, 14, 15, 16, 18, 20] as const;

export interface TerminalSettings {
  fontSize: number;
  fontFamily: string;
  themeName: string;
}

export const DEFAULT_TERMINAL_SETTINGS: TerminalSettings = {
  fontSize: 13,
  fontFamily: FONT_FAMILIES[0],
  themeName: "Default Dark",
};
