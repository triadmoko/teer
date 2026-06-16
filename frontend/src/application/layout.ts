
import { writable } from "svelte/store";
import { GRID, clampRowH, type LayoutMode } from "@domain/layout";

export const layoutMode = writable<LayoutMode>("tabs");

export const gridCols = writable<number>(GRID.DEFAULT_COLS);

export const gridRowH = writable<number>(GRID.DEFAULT_ROW_H);

export function setRowH(h: number): void {
  gridRowH.set(clampRowH(h));
}

export const fullscreenSessionId = writable<string | null>(null);
