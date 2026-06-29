import type { Terminal } from "@xterm/xterm";
import type { FitAddon } from "@xterm/addon-fit";
import type { SearchAddon } from "@xterm/addon-search";
import type { SerializeAddon } from "@xterm/addon-serialize";

export type TerminalEntry = {
  term: Terminal;
  fit: FitAddon;
  search: SearchAddon;
  serialize: SerializeAddon;
  el: HTMLDivElement;
  // Mutable ref agar listener onSessionOutput (yang tetap aktif saat parking)
  // dan snapshot interval (di-reset tiap checkout) bisa berbagi dirty flag.
  dirtyRef: { value: boolean };
  started: boolean;
};

const registry = new Map<string, TerminalEntry>();

function ensureParking(): HTMLDivElement {
  const ID = "__t_park__";
  let el = document.getElementById(ID) as HTMLDivElement | null;
  if (!el) {
    el = document.createElement("div");
    el.id = ID;
    // Off-screen, zero-size, tak terlihat — xterm buffer tetap hidup di memori
    // tapi canvas tidak di-resize ke 0×0 karena ResizeObserver diputus saat park.
    el.style.cssText =
      "position:fixed;left:-9999px;top:0;width:0;height:0;" +
      "overflow:hidden;pointer-events:none;visibility:hidden;";
    document.body.appendChild(el);
  }
  return el;
}

export function park(sessionId: string, entry: TerminalEntry): void {
  registry.set(sessionId, entry);
  ensureParking().appendChild(entry.el);
}

export function checkout(sessionId: string): TerminalEntry | undefined {
  const entry = registry.get(sessionId);
  if (entry) registry.delete(sessionId);
  return entry;
}
