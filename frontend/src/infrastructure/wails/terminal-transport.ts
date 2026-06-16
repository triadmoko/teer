// Adapter Wails untuk aliran event terminal (output & exit PTY).
// Komponen UI berlangganan lewat sini, bukan ke Events.On mentah.

import { Events } from "@wailsio/runtime";
import { b64ToBytes } from "../encoding";

export const terminalTransport = {
  /**
   * Berlangganan output PTY sebuah sesi; callback menerima byte mentah
   * (hasil decode base64). Mengembalikan fungsi untuk berhenti berlangganan.
   */
  onOutput(sessionId: string, cb: (bytes: Uint8Array) => void): () => void {
    return Events.On(`session:${sessionId}:out`, (ev) => {
      cb(b64ToBytes(ev.data as string));
    });
  },

  /**
   * Berlangganan event berakhirnya proses sesi; callback menerima kode exit.
   * Mengembalikan fungsi untuk berhenti berlangganan.
   */
  onExit(sessionId: string, cb: (code: number) => void): () => void {
    return Events.On(`session:${sessionId}:exit`, (ev) => {
      const code = (ev.data as { code?: number })?.code ?? 0;
      cb(code);
    });
  },
};
