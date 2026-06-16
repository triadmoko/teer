// Lapisan aplikasi: dialog dalam-app pengganti window.prompt/confirm bawaan.
//
// WebKitGTK (webview Wails di Linux) tidak mendukung prompt()/confirm() —
// keduanya mengembalikan null/false tanpa menampilkan apa pun. Kita pakai
// modal Svelte sendiri yang mengembalikan Promise.

import { writable } from "svelte/store";

export type DialogKind = "prompt" | "confirm";

export interface DialogState {
  kind: DialogKind;
  title: string;
  defaultValue: string;
  placeholder: string;
  confirmLabel: string;
  danger: boolean;
  resolve: (value: string | boolean | null) => void;
}

export const dialog = writable<DialogState | null>(null);

/** Tampilkan input teks. Resolve ke string (OK) atau null (batal). */
export function promptDialog(
  title: string,
  defaultValue = "",
  placeholder = "",
): Promise<string | null> {
  return new Promise((resolve) => {
    dialog.set({
      kind: "prompt",
      title,
      defaultValue,
      placeholder,
      confirmLabel: "Simpan",
      danger: false,
      resolve: (v) => resolve(v as string | null),
    });
  });
}

/** Tampilkan konfirmasi. Resolve ke true (ya) atau false (batal). */
export function confirmDialog(
  title: string,
  opts: { confirmLabel?: string; danger?: boolean } = {},
): Promise<boolean> {
  return new Promise((resolve) => {
    dialog.set({
      kind: "confirm",
      title,
      defaultValue: "",
      placeholder: "",
      confirmLabel: opts.confirmLabel ?? "Ya",
      danger: opts.danger ?? false,
      resolve: (v) => resolve(v === true),
    });
  });
}
