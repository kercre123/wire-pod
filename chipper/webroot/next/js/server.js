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

// Our own builds are named after the upstream release they are based on, like v1.2.18-lucas.5. The server
// compares that string with upstream's v1.2.18 and calls it an update; a build at or ahead of the upstream
// release is not behind it.
const triple = (v) => { const m = /^v?(\d+)\.(\d+)\.(\d+)/.exec(v || ""); return m ? m.slice(1).map(Number) : null; };
export function updateAvailable(v) {
  if (!v || !v.avail) return false;
  if (v.fromsource || !/-lucas\.\d+$/.test(v.installedversion || "")) return true;
  const have = triple(v.installedversion), latest = triple(v.currentversion);
  if (!have || !latest) return true;
  for (let i = 0; i < 3; i++) if (have[i] !== latest[i]) return have[i] < latest[i];
  return false;
}

// get_version_info asks GitHub, so it fails with no internet. That is not an error worth showing loudly.
export async function loadVersion() {
  serverStore.set({ versionState: "loading" });
  try {
    const v = await getJSON("/api/get_version_info", { timeout: 15000 });
    serverStore.set({ version: { ...v, avail: updateAvailable(v) }, versionState: "ok" });
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
