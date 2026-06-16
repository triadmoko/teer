// Adapter Wails untuk siklus hidup PTY (SessionService).
// Membungkus binding RPC start/write/resize/close + daftar sesi berjalan.

import { SessionService } from "@bindings/teer/internal/service";
import type { StartOptions } from "@bindings/teer/internal/service";

/** Parameter untuk membangkitkan sebuah sesi PTY. */
export interface StartSessionInput {
  id: string;
  shell: string;
  cwd: string;
  env: Record<string, string>;
  startupCommand: string;
  cols: number;
  rows: number;
}

export const sessionGateway = {
  listRunning: () => SessionService.ListRunning(),

  start: (opts: StartSessionInput) =>
    SessionService.StartSession(opts as unknown as StartOptions),

  write: (id: string, data: string) => SessionService.WriteSession(id, data),

  resize: (id: string, cols: number, rows: number) =>
    SessionService.ResizeSession(id, cols, rows),

  close: (id: string) => SessionService.CloseSession(id),
};
