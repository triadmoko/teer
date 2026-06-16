// Lapisan domain: aturan tata letak panggung terminal (PRD §11).

/** Mode tata letak panggung: "tabs" (satu aktif) atau "grid" (ubin). */
export type LayoutMode = "tabs" | "grid";

/** Batas & default tata letak grid. */
export const GRID = {
  MIN_ROW_H: 180,
  MAX_ROW_H: 800,
  DEFAULT_ROW_H: 320,
  MIN_COLS: 1,
  MAX_COLS: 4,
  DEFAULT_COLS: 2,
} as const;

/** Pilihan jumlah kolom yang tersedia di UI grid. */
export const COL_CHOICES = [1, 2, 3, 4] as const;

/** Jepit tinggi baris grid ke rentang yang diizinkan. */
export function clampRowH(h: number): number {
  return Math.min(GRID.MAX_ROW_H, Math.max(GRID.MIN_ROW_H, h));
}
