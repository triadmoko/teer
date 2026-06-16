// Lapisan aplikasi: helper internal untuk set sesi yang sudah di-mount.

import { opened } from "./stores";

/** Tandai sebuah sesi sebagai terbuka (terminalnya di-mount). */
export function open(id: string): void {
  opened.update((s) => {
    if (s.has(id)) return s;
    const n = new Set(s);
    n.add(id);
    return n;
  });
}

/** Hapus tanda terbuka sebuah sesi (terminalnya dilepas). */
export function closeOpened(id: string): void {
  opened.update((s) => {
    if (!s.has(id)) return s;
    const n = new Set(s);
    n.delete(id);
    return n;
  });
}
