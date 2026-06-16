
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
