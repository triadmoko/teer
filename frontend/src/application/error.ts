// Lapisan aplikasi: penanganan error lintas use case.

import { writable } from "svelte/store";

/** Pesan error untuk ditampilkan di banner (mis. backend tidak tersambung). */
export const lastError = writable<string | null>(null);

const BACKEND_HINT =
  "Tidak terhubung ke backend Go. Jalankan dalam mode DESKTOP " +
  "(`task dev` lalu pakai jendela yang muncul, atau `./bin/teer`) — " +
  "bukan membuka URL di browser.";

/** Jalankan aksi infra; tangkap error agar tidak gagal diam-diam. */
export async function guard<T>(fn: () => Promise<T>): Promise<T | undefined> {
  try {
    const r = await fn();
    lastError.set(null);
    return r;
  } catch (e) {
    console.error("teer binding error:", e);
    lastError.set(BACKEND_HINT);
    return undefined;
  }
}
