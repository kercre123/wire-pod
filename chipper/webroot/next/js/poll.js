// Polling that is easy on a small computer: it only runs while the tab is visible and the panel it feeds is
// on screen, and a poll that keeps failing waits twice as long each time (up to a minute).
import { useEffect, useRef, useState } from "./html.js";
import { nextDelay } from "./backoff.js";

// True while the tab is showing and, if a ref is given, its element is on screen.
export function useVisible(ref) {
  const [tab, setTab] = useState(!document.hidden);
  const [seen, setSeen] = useState(!ref);
  useEffect(() => {
    const f = () => setTab(!document.hidden);
    document.addEventListener("visibilitychange", f);
    return () => document.removeEventListener("visibilitychange", f);
  }, []);
  useEffect(() => {
    if (!ref || !ref.current || typeof IntersectionObserver !== "function") { setSeen(true); return; }
    const io = new IntersectionObserver((es) => setSeen(es[es.length - 1].isIntersecting));
    io.observe(ref.current);
    return () => io.disconnect();
  }, []);
  return tab && seen;
}

// Calls fn now and then every `every` ms while `enabled`. fn throws (or rejects) to say it failed.
export function usePoll(fn, every, { enabled = true, max = 60000, deps = [] } = {}) {
  const latest = useRef(fn);
  latest.current = fn;
  useEffect(() => {
    if (!enabled) return;
    let alive = true, timer, fails = 0;
    const run = async () => {
      try { await latest.current(); fails = 0; } catch (e) { fails++; }
      if (alive) timer = setTimeout(run, nextDelay(every, fails, max));
    };
    run();
    return () => { alive = false; clearTimeout(timer); };
  }, [enabled, ...deps]);
}
