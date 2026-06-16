import { UpdaterService } from "@bindings/teer/internal/service";
import type { UpdateInfo } from "@bindings/teer/internal/service";

export type { UpdateInfo };

export const updaterGateway = {
  checkUpdate: (): Promise<UpdateInfo | null> => UpdaterService.CheckUpdate(),
  downloadAndApply: (downloadURL: string): Promise<void> =>
    UpdaterService.DownloadAndApply(downloadURL),
};
