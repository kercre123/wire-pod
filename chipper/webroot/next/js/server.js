// Server-wide facts the shell and the home screen show: config and version.
import { useEffect } from "./html.js";
import { createStore, useStore } from "./store.js";
import { getJSON } from "./api.js";

export const serverStore = createStore({ config: null, configError: false, version: null, versionState: "idle" });
let started = false;

export async function loadConfig() {
  try {
    serverStore.set({ config: await getJSON("/api/get_config", { timeout: 8000 }), configError: false });
  } catch (e) {
    serverStore.set({ configError: true });
  }
}

// get_version_info asks GitHub, so it fails with no internet. That is not an error worth showing loudly.
export async function loadVersion() {
  serverStore.set({ versionState: "loading" });
  try {
    serverStore.set({ version: await getJSON("/api/get_version_info", { timeout: 15000 }), versionState: "ok" });
  } catch (e) {
    serverStore.set({ versionState: "unavailable" });
  }
}

export function useServerInfo() {
  const s = useStore(serverStore);
  useEffect(() => {
    if (started) return;
    started = true;
    loadConfig();
    loadVersion();
  }, []);
  return s;
}

// "v1.0.3" or "commit abc1234" for source builds
export function versionLabel(v) {
  if (!v) return "";
  if (v.fromsource) return v.installedcommit ? `commit ${v.installedcommit}` : "";
  return v.installedversion || "";
}
