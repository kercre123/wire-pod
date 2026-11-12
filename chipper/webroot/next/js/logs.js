// wire-pod keeps the last 50 interesting log lines in memory. The home cards read the last thing each
// robot heard from them, and the Activity tab filters them per robot.
import { useEffect, useState } from "./html.js";
import { getText } from "./api.js";

const HEARD = /^(\d{4})\.(\d\d)\.(\d\d) (\d\d):(\d\d):(\d\d): Intent matched: .*?transcribed text: '(.*)', device: (\S+)/;

// Returns { esn: { text, at: Date } } with the latest line per robot.
export function parseHeard(logText) {
  const out = {};
  for (const line of String(logText).split("\n")) {
    const m = HEARD.exec(line);
    if (!m) continue;
    const at = new Date(+m[1], +m[2] - 1, +m[3], +m[4], +m[5], +m[6]);
    out[m[8].toLowerCase()] = { text: m[7], at };
  }
  return out;
}

export function useHeard(everyMs = 8000) {
  const [heard, setHeard] = useState({});
  useEffect(() => {
    let alive = true;
    let timer;
    const run = async () => {
      if (!document.hidden) {
        try {
          const t = await getText("/api/get_logs", { timeout: 6000 });
          if (alive) setHeard(parseHeard(t));
        } catch (e) {
          // the log is a nice-to-have; leave the cards as they are
        }
      }
      if (alive) timer = setTimeout(run, everyMs);
    };
    run();
    return () => { alive = false; clearTimeout(timer); };
  }, []);
  return heard;
}

export function linesFor(logText, esn) {
  const needle = esn.toLowerCase();
  return String(logText).split("\n").filter((l) => l.toLowerCase().includes(needle)).join("\n");
}
