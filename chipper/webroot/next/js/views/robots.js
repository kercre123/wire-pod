// Home: one card per robot, with a one-line server strip.
import { html, useEffect } from "../html.js";
import { Icon } from "../icons.js";
import { href } from "../router.js";
import { useRobots, robotName, robotTitle } from "../store.js";
import { Face, StatusPill, Battery, Loading, Callout, pageTitle } from "../ui.js";
import { useServerInfo, versionLabel } from "../server.js";
import { useHeard } from "../logs.js";
import { aiSummary, languageLabel, timeAgo } from "../data.js";

function RobotCard({ robot, heard }) {
  const name = robotName(robot);
  const offline = robot.state === "offline";
  let foot;
  if (offline) {
    foot = robot.last_seen ? `Last seen ${timeAgo(new Date(robot.last_seen * 1000))}` : "Not seen since WirePod started";
  } else if (heard) {
    foot = `“${heard.text}” · ${timeAgo(heard.at)}`;
  } else {
    foot = "Nothing heard yet";
  }
  return html`<div class="rcard">
    <div class="rc-top">
      <${Face} robot=${robot} size=${48} />
      <div class="rc-id"><b><a class="stretch" href=${href.robot(robot.esn)} style="color:inherit;text-decoration:none">${robotTitle(robot)}</a></b><span class="mono">${robot.esn}</span></div>
      <${StatusPill} robot=${robot} />
    </div>
    ${robot.state === "checking" ? html`<div class="bat"><span class="cell none"></span><span>Checking…</span></div>` : html`<${Battery} battery=${robot.battery} />`}
    ${offline
      ? html`<${Callout} icon="info"><span>${name} isn’t answering. Check that it’s powered on and on the same Wi-Fi as this computer. You can still edit its personality and memories.</span><//>`
      : html`<p class="persona">${robot.personality || html`<span class="sub">No personality set yet. Open ${name} to give it one.</span>`}</p>`}
    <div class="rc-foot"><span>${foot}</span></div>
  </div>`;
}

function Strip({ server }) {
  const cfg = server.config;
  const kg = cfg && cfg.knowledge;
  const stt = cfg && cfg.STT;
  const weather = cfg && cfg.weather;
  const v = server.version;
  const item = (label, dot, text, to) => html`<a href=${href.settings(to)}><span class="lab">${label}</span><span class="val"><span class=${"dot " + dot}></span><span>${text}</span></span></a>`;
  return html`<section class="strip" aria-label="Server">
    <div class="head"><${Icon} name="server" size=${18} />Server</div>
    ${item("AI", kg && kg.enable ? "" : "mute", cfg ? aiSummary(kg) : "…", "ai")}
    ${item("Voice", "", stt ? (stt.provider === "vosk" || stt.provider === "whisper.cpp" ? languageLabel(stt.language) : stt.provider) : "…", "voice")}
    ${item("Weather", weather && weather.enable ? "" : "mute", cfg ? (weather && weather.enable ? "Set up" : "Not set up") : "…", "integrations")}
    ${item("Version", v && v.avail ? "warn" : v ? "" : "mute", v ? (v.avail ? "Update ready" : versionLabel(v) || "Up to date") : server.versionState === "loading" ? "Checking…" : "Can’t check", "server")}
  </section>`;
}

export function RobotsHome() {
  const { loaded, robots, error } = useRobots();
  const server = useServerInfo();
  const heard = useHeard();
  useEffect(() => pageTitle("Robots"), []);
  const reachable = robots.filter((r) => r.state === "online").length;
  const checking = robots.some((r) => r.state === "checking");
  const sub = !loaded ? "" : robots.length === 0 ? "No robots yet"
    : `${robots.length} robot${robots.length === 1 ? "" : "s"} on this network${checking ? "" : ` · ${reachable} reachable`}`;
  return html`<div class="page">
    <div class="top">
      <div><h1>Your robots</h1><p class="sub" aria-live="polite">${sub}</p></div>
      <a class="btn pri" href=${href.add}><${Icon} name="plus" size=${18} />Add a robot</a>
    </div>
    ${error && html`<${Callout} icon="alert" kind="bad"><span>${error}</span><//>`}
    ${!loaded && html`<${Loading}>Looking for your robots…<//>`}
    ${loaded && robots.length === 0 && !error && html`<div class="card offline-note">
      <h2>No robots yet</h2>
      <p class="sub">Add your Vector to give it a voice, a memory and a personality. It takes a few minutes and a Bluetooth connection.</p>
      <a class="btn pri" href=${href.add}><${Icon} name="plus" size=${18} />Add a robot</a>
    </div>`}
    ${robots.length > 0 && html`<div class="cards">${robots.map((r) => html`<${RobotCard} key=${r.esn} robot=${r} heard=${heard[r.esn]} />`)}</div>`}
    <${Strip} server=${server} />
    ${robots.length > 0 && html`<p class="sub" style="font-size:.8125rem">Tip: open a robot to change its name, personality and memories.</p>`}
  </div>`;
}
