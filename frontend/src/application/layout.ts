// Lapisan aplikasi: state & aksi tata letak panggung terminal.

import { writable } from "svelte/store";
import { GRID, clampRowH, type LayoutMode } from "@domain/layout";

/** Mode tata letak panggung: "tabs" (satu aktif) atau "grid" (ubin). */
export const layoutMode = writable<LayoutMode>("tabs");
/** Jumlah kolom saat mode grid (1–4), dikontrol manual oleh pengguna. */
export const gridCols = writable<number>(GRID.DEFAULT_COLS);
/** Tinggi tiap baris terminal grid dalam px; bila total melebihi panggung,
 *  area grid bisa di-scroll vertikal. */
export const gridRowH = writable<number>(GRID.DEFAULT_ROW_H);

/** Setel tinggi baris grid, dijepit ke rentang yang diizinkan. */
export function setRowH(h: number): void {
  gridRowH.set(clampRowH(h));
}

/** ID sesi yang sedang fullscreen di mode grid; null = tidak ada. */
export const fullscreenSessionId = writable<string | null>(null);
