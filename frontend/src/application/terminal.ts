// Lapisan aplikasi: use case I/O terminal aktif (dipakai komponen Terminal).
// Menjembatani xterm di presentation dengan adapter PTY di infrastructure.

import { mergeEnv, type SessionDef } from "@domain/models";
import { sessionGateway, terminalTransport } from "@infrastructure/wails";

/** Bangkitkan PTY untuk sesi dengan env workspace+sesi tergabung (FR-13/18). */
export function startSession(
  session: SessionDef,
  wsEnv: Record<string, string>,
  wsCwd: string,
  size: { cols: number; rows: number },
): Promise<void> {
  return sessionGateway.start({
    id: session.id,
    shell: session.shell ?? "",
    cwd: session.cwd || wsCwd,
    env: mergeEnv(wsEnv, session.env),
    startupCommand: session.startupCommand ?? "",
    cols: size.cols,
    rows: size.rows,
  });
}

/** Kirim input keyboard/paste ke stdin shell (FR-18). */
export const writeSession = (id: string, data: string): Promise<void> =>
  sessionGateway.write(id, data);

/** Sesuaikan ukuran PTY mengikuti ukuran pane (FR-17). */
export const resizeSession = (
  id: string,
  cols: number,
  rows: number,
): Promise<void> => sessionGateway.resize(id, cols, rows);

/** Berlangganan output PTY sebuah sesi (byte mentah). */
export const onSessionOutput = (
  id: string,
  cb: (bytes: Uint8Array) => void,
): (() => void) => terminalTransport.onOutput(id, cb);

/** Berlangganan event berakhirnya proses sesi. */
export const onSessionExit = (
  id: string,
  cb: (code: number) => void,
): (() => void) => terminalTransport.onExit(id, cb);
