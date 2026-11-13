// WirePod web UI entry point: shell (rail on desktop, tab bar on phones), routing, appearance.
import { html, render, useEffect, useRef } from "./js/html.js";
import { Icon } from "./js/icons.js";
import { useRoute, href, go } from "./js/router.js";
import { useStore, prefsStore, useRobots, clearToasts } from "./js/store.js";
import { applyTheme, watchSystemTheme } from "./js/theme.js";
import { Toasts, ConfirmHost, pageTitle } from "./js/ui.js";
import { useServerInfo, versionLabel } from "./js/server.js";
import { RobotsHome } from "./js/views/robots.js";
import { RobotPage } from "./js/views/robot.js";
import { Settings } from "./js/views/settings.js";
import { Wizard } from "./js/views/wizard.js";

function Logo() {
  return html`<span class="logo" aria-hidden="true"><i></i><i></i></span>`;
}

function Rail({ section, robotCount, server }) {
  const v = versionLabel(server.version);
  const update = server.version && server.version.avail;
  return html`<aside class="rail">
    <div class="brand"><${Logo} />WirePod</div>
    <nav aria-label="Main">
      <a class="nv" href=${href.robots} aria-current=${section === "robots" ? "page" : null}><${Icon} name="robots" />Robots${robotCount > 0 && html`<span class="cnt">${robotCount}</span>`}</a>
      <a class="nv" href=${href.settings()} aria-current=${section === "settings" ? "page" : null}><${Icon} name="sliders" />Settings</a>
    </nav>
    <div class="rail-foot">
      <div class="srv"><b><span class=${"dot" + (server.configError ? " bad" : "")}></span>${server.configError ? "Server not answering" : "Server running"}</b>
        ${(v || update) && html`<span>${v}${update && html` · <a href=${href.settings("server")}>Update ready</a>`}</span>`}</div>
      <a class="classic" href="/classic">Open classic UI</a>
    </div>
  </aside>`;
}

function TabBar({ section }) {
  return html`<nav class="tabbar" aria-label="Main">
    <a href=${href.robots} aria-current=${section === "robots" ? "page" : null}><${Icon} name="robots" size=${24} />Robots</a>
    <a href=${href.settings()} aria-current=${section === "settings" ? "page" : null}><${Icon} name="sliders" size=${24} />Settings</a>
  </nav>`;
}

function NotFound() {
  useEffect(() => pageTitle("Not found"), []);
  return html`<div class="page"><h1>Page not found</h1><p class="sub">That address doesn’t lead anywhere. <a href=${href.robots}>Back to your robots</a>.</p></div>`;
}

function App() {
  const route = useRoute();
  const { prefs } = useStore(prefsStore);
  const { robots } = useRobots({ watch: false }); // each screen that shows robot status polls for itself
  const server = useServerInfo();
  const mainRef = useRef(null);

  // toasts belong to the screen that raised them
  useEffect(() => { window.addEventListener("hashchange", clearToasts); return () => window.removeEventListener("hashchange", clearToasts); }, []);

  // a brand new install goes through first-run setup
  const needsSetup = !!(server.config && !server.config.pastinitialsetup);
  useEffect(() => { if (needsSetup && route.name !== "welcome") go(href.welcome, { replace: true }); }, [needsSetup, route.name]);

  const robot = route.name === "robot" ? robots.find((r) => r.esn === route.esn.toLowerCase()) : null;
  const eye = robot && robot.eye;
  useEffect(() => { applyTheme(prefs, eye); }, [prefs, eye]);
  const eyeNow = useRef(eye);
  eyeNow.current = eye;
  useEffect(() => watchSystemTheme(() => applyTheme(prefsStore.get().prefs, eyeNow.current)), []);

  // After moving to a new screen, put keyboard and screen-reader focus on its heading.
  const first = useRef(true);
  const key = route.name === "robot" ? `robot:${route.esn}` : route.name;
  useEffect(() => {
    if (first.current) { first.current = false; return; }
    const h1 = mainRef.current && mainRef.current.querySelector("h1");
    if (h1) { h1.setAttribute("tabindex", "-1"); h1.focus({ preventScroll: false }); }
    window.scrollTo(0, 0);
  }, [key]);

  if (route.name === "welcome" || route.name === "add") {
    return html`<${Wizard} first=${route.name === "welcome"} /><${Toasts} /><${ConfirmHost} />`;
  }
  const section = route.name === "settings" ? "settings" : "robots";
  let view;
  switch (route.name) {
    case "robots": view = html`<${RobotsHome} />`; break;
    case "robot": view = html`<${RobotPage} esn=${route.esn} tab=${route.tab} />`; break;
    case "settings": view = html`<${Settings} section=${route.section} back=${route.back} />`; break;
    default: view = html`<${NotFound} />`;
  }
  return html`<div class="app">
    <${Rail} section=${section} robotCount=${robots.length} server=${server} />
    <div>
      <header class="mbar"><${Logo} />WirePod</header>
      <main id="main" class="main" ref=${mainRef}>${view}</main>
    </div>
    <${TabBar} section=${section} />
    <${Toasts} /><${ConfirmHost} />
  </div>`;
}

render(html`<${App} />`, document.getElementById("root"));
