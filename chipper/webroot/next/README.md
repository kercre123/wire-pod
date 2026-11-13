# WirePod web UI

Served at `/next/` (it opens at `/` only when the server config says `"ui": {"default": "next"}`). The original UI is the default and is at `/classic`, and every old URL
(`/index.html`, `/setup.html`, `/initial.html`, `/sdkapp/...`) keeps working.

- Preact + htm as plain ES modules, vendored in `vendor/` (see `vendor/README.md`). No build step,
  no Node, no CDN. Fonts and icons are local too, so it works with no internet.
- `index.html` has the import map; `app.js` is the shell and router; `js/views/` has one file per screen.
- `js/api.js` is the only place that talks to wire-pod, and turns errors into plain sentences.
- Hash routes: `#/`, `#/robot/<esn>/<tab>`, `#/settings/<section>`, `#/add`, `#/welcome`.
- Bluetooth pairing is not reimplemented: `#/add` opens the classic Bot Setup panel (`/classic#bot-setup`)
  and picks the robot up when it appears.
- Robot status comes from one cached call, `GET /api-sdk/robots_status`, so an offline robot never slows the page.
- Built for a Pi: the server caches and gzips `/next` (see `pkg/wirepod/config-ws/nextassets.go`; `vendor/` and `fonts/`
  get `?v=<hash>` URLs written in as they are served, so they are cached for a year and an upgrade still busts them).
  Polling goes through `js/poll.js`: only while the tab is visible and the panel is on screen, doubling its wait
  (up to a minute) while it fails. A screen that does not show robot status does not poll it.
