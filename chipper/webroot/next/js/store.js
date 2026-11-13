// Small shared state: robot status polling, toasts, confirm dialogs and appearance prefs.
import { useEffect, useReducer } from "./html.js";
import { fetchRobotStatuses, describeError } from "./api.js";
import { nextDelay } from "./backoff.js";

export function createStore(initial) {
  let state = initial;
  const subs = new Set();
  return {
    get: () => state,
    set(patch) {
      state = { ...state, ...(typeof patch === "function" ? patch(state) : patch) };
      subs.forEach((f) => f());
    },
    subscribe(f) {
      subs.add(f);
      return () => subs.delete(f);
    },
  };
}

export function useStore(store) {
  const [, bump] = useReducer((n) => n + 1, 0);
  useEffect(() => store.subscribe(bump), [store]);
  return store.get();
}

// ---------- robots: one cached status call, polled ----------

export const robotsStore = createStore({ loaded: false, robots: [], error: null });
let pollers = 0;
let timer = null;
let inflight = false;
let fails = 0; // polls in a row that failed, or found every robot offline and nothing new: each doubles the wait

// What a robot card shows. The server stamps every online probe with fresh times that nothing displays, so
// they are left out: a poll that finds nothing new changes nothing, and views are not re-rendered.
const shown = (r) => JSON.stringify(r.state === "online" ? { ...r, checked_at: 0, last_seen: 0 } : { ...r, checked_at: 0 });

// Merge a fresh poll into what we have:
//  - a robot whose displayed data did not change keeps its object (same identity), so nothing keyed on it reloads;
//  - one failed or unfinished check does not flip an online robot to offline: it takes two in a row
//    (a probe that merely timed out once would otherwise blank the robot's screens and reload them).
const missed = new Map(); // esn -> checked_at of the one failed check we are holding back for an online robot
function mergeRobots(prev, next) {
  const old = new Map(prev.map((r) => [r.esn, r]));
  const merged = next.map((r) => {
    const was = old.get(r.esn);
    if (r.state === "online") missed.delete(r.esn);
    else if (was && was.state === "online") {
      if (r.state === "checking") return was;
      // the same cached failure seen again is still one failure; a different, later check that also fails is real
      if (!missed.has(r.esn) || missed.get(r.esn) === r.checked_at) { missed.set(r.esn, r.checked_at); return was; }
    }
    return was && shown(was) === shown(r) ? was : r;
  });
  for (const esn of [...missed.keys()]) if (!next.some((r) => r.esn === esn)) missed.delete(esn);
  const same = merged.length === prev.length && merged.every((r, i) => r === prev[i]);
  return same ? prev : merged;
}

async function tick(recheck) {
  if (inflight) return;
  inflight = true;
  clearTimeout(timer);
  try {
    const next = await fetchRobotStatuses({ timeout: 10000, recheck });
    const prev = robotsStore.get();
    const robots = mergeRobots(prev.robots, next);
    if (!prev.loaded || prev.error || robots !== prev.robots) robotsStore.set({ loaded: true, robots, error: null });
    fails = next.length > 0 && next.every((r) => r.state === "offline") && robots === prev.robots ? fails + 1 : 0;
  } catch (e) {
    fails++;
    robotsStore.set({ loaded: true, error: describeError(e).message });
  } finally {
    inflight = false;
  }
  schedule();
}

function schedule() {
  clearTimeout(timer);
  if (pollers === 0 || document.hidden) return;
  const checking = robotsStore.get().robots.some((r) => r.state === "checking");
  timer = setTimeout(() => { if (!document.hidden) tick(); }, nextDelay(checking ? 1200 : missed.size > 0 ? 2500 : 5000, fails)); // a tab hidden while waiting does not poll
}

document.addEventListener("visibilitychange", () => {
  if (!document.hidden && pollers > 0) { fails = 0; tick(); }
});

export function refreshRobots() {
  fails = 0;
  return tick(true);
}

// Components that show robot status call this; polling runs while at least one is mounted.
// watch: false only reads (and loads once if nothing has), for screens that do not show robot status.
export function useRobots({ watch = true } = {}) {
  const s = useStore(robotsStore);
  useEffect(() => {
    if (!watch) { if (!robotsStore.get().loaded) tick(); return; }
    pollers++;
    if (pollers === 1 || !robotsStore.get().loaded) tick();
    else schedule();
    return () => {
      pollers--;
      if (pollers === 0) clearTimeout(timer);
    };
  }, []);
  return s;
}

export function robotName(r) {
  return r ? r.name || `Vector ${r.esn}` : "Robot";
}

// For titles that show the serial number next to the name: a robot with no custom name is just "Vector".
export function robotTitle(r) {
  return r ? r.name || "Vector" : "Robot";
}

// ---------- last-known data ----------
// Screens remember what they last loaded. Coming back shows it at once and refreshes quietly, so a spinner
// only ever appears the first time something is opened.
const lastKnown = new Map();
export const cacheGet = (key) => lastKnown.get(key);
export const cachePut = (key, value) => { lastKnown.set(key, value); return value; };

// ---------- toasts ----------

export const toastStore = createStore({ items: [] });
let toastId = 0;

// toast("Saved") or toast("Memory deleted", { undo: fn }) or toast("Couldn't save", { kind: "bad" })
export function toast(message, { undo, kind = "ok", ms, persist = false } = {}) {
  const id = ++toastId;
  const item = { id, message, undo, kind, persist };
  toastStore.set((s) => ({ items: [...s.items.slice(-2), item] }));
  setTimeout(() => dismissToast(id), ms || (undo ? 7000 : kind === "bad" ? 7000 : 3500));
  return id;
}

// Moving to another screen clears what the last one said, except messages meant for the next one.
export function clearToasts() {
  toastStore.set((s) => (s.items.some((t) => !t.persist) ? { items: s.items.filter((t) => t.persist) } : s));
}

export function dismissToast(id) {
  toastStore.set((s) => ({ items: s.items.filter((t) => t.id !== id) }));
}

// ---------- confirm dialog ----------

export const confirmStore = createStore({ current: null });

// await confirmDialog({ title, body, confirmLabel, danger }) -> true | false
export function confirmDialog(opts) {
  return new Promise((resolve) => {
    confirmStore.set({ current: { ...opts, resolve } });
  });
}

// ---------- appearance ----------

const PREFS_KEY = "wirepod-next-prefs";
export const DEFAULT_PREFS = { theme: "system", accent: "teal", accentFollowsRobot: false, size: "default", font: "default" };

function loadPrefs() {
  try {
    return { ...DEFAULT_PREFS, ...JSON.parse(localStorage.getItem(PREFS_KEY) || "{}") };
  } catch (e) {
    return { ...DEFAULT_PREFS };
  }
}

export const prefsStore = createStore({ prefs: loadPrefs() });

export function setPref(patch) {
  const prefs = { ...prefsStore.get().prefs, ...patch };
  try {
    localStorage.setItem(PREFS_KEY, JSON.stringify(prefs));
  } catch (e) {
    // private mode or blocked storage: the choice still applies until the page closes
  }
  prefsStore.set({ prefs });
}
