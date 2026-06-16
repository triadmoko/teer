import { writable } from "svelte/store";
import { Events } from "@wailsio/runtime";
import { updaterGateway } from "@infrastructure/wails";
import type { UpdateInfo } from "@infrastructure/wails";

export type UpdateProgress = {
  stage: "downloading" | "applying" | "done";
  percent: number;
};

export const updateInfo = writable<UpdateInfo | null>(null);
export const updateProgress = writable<UpdateProgress | null>(null);
export const updateApplying = writable(false);

export async function checkUpdate(): Promise<void> {
  try {
    const info = await updaterGateway.checkUpdate();
    if (info?.available) {
      updateInfo.set(info);
    }
  } catch {
    // cek update gagal — silent, tidak ganggu user
  }
}

export function listenUpdateProgress(): () => void {
  return Events.On("updater:progress", (ev) => {
    const data = ev.data as UpdateProgress;
    updateProgress.set(data);
    if (data.stage === "done") {
      setTimeout(() => updateProgress.set(null), 800);
    }
  });
}

export async function applyUpdate(downloadURL: string): Promise<void> {
  updateApplying.set(true);
  updateInfo.set(null);
  try {
    await updaterGateway.downloadAndApply(downloadURL);
  } catch (e) {
    updateApplying.set(false);
    updateProgress.set(null);
    throw e;
  }
}

export function dismissUpdate(): void {
  updateInfo.set(null);
}
