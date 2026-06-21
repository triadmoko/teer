
export type LayoutMode = "tabs" | "grid";

export const GRID = {
  MIN_ROW_H: 180,
  MAX_ROW_H: 800,
  DEFAULT_ROW_H: 320,
  MIN_COLS: 1,
  MAX_COLS: 12,
  DEFAULT_COLS: 2,
} as const;

export const COL_CHOICES = [1, 2, 3, 4] as const;
export const COL_PRESET_CHOICES = [1, 2, 3, 4, 5, 6] as const;

export function clampRowH(h: number): number {
  return Math.min(GRID.MAX_ROW_H, Math.max(GRID.MIN_ROW_H, h));
}
