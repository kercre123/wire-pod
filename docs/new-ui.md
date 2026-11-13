# The new WirePod web UI

A fresh front page for wire-pod: one place to see your robots, give them personalities and memories, drive them, and manage the server. It opens at **`/next`**, and `/` now redirects there. The original UI is still at **`/classic`**.

![The robots home: three robots with battery, personality and the last thing each heard](images/robots-home.png)

It's plain HTML and JavaScript served by wire-pod itself. There's nothing to install, no Node at runtime, and no internet needed: the fonts, icons and the small [Preact](https://preactjs.com) library are vendored in the repo.

## Try it

Open `http://<your wire-pod address>:8080/next/` (the same address you already use, plus `/next/`). A fresh install starts with a short setup; an existing one lands on your robots.

## A quick tour

**Robots** (home). One card per robot with its name, serial number, online state, battery, personality, and the last thing it heard. Robots that are off don't slow the page down: it shows what it last knew and keeps checking quietly. A strip along the bottom summarises the server (AI, voice, weather, version).

**A robot's page** has five tabs.

| Tab | What it does |
|---|---|
| **Personality** | The robot's name and personality, with a live preview of what the AI is told. See [AI features](ai-features.md). |
| **Memories** | What the robot remembers (add, delete, clear), plus the faces it knows. |
| **Settings** | Things stored on the robot: volume, eye colour, button behaviour, Alexa sign-in, language, clock, units, time zone, weather location. Each change saves as soon as you make it. |
| **Controls** | Camera, a drive pad, lift and head, "say something", play a sound, quick actions (explore, dance, fetch the cube, go to the charger, sleep) and mirror mode. |
| **Activity** | Recent conversations, lifetime stats, mood, photos and a technical log filtered to that robot. |

![Controls: camera, quick actions, say something and a drive pad](images/controls.png)

**Settings** (whole server) has five sections: **AI & knowledge**, **Voice & commands** (speech language and your custom commands), **Integrations** (weather), **Server & updates**, and **Appearance**.

**Add a robot** and the **first-run setup** walk you through connection, language, pairing and naming. Bluetooth pairing itself is still the tried-and-tested flow from the classic UI: the new page sends you to it (`/classic#bot-setup`) and picks the robot up when it appears.

### If AI is off

Personality and Memories stay visible but locked, with a friendly "turn on AI" panel. Your data is kept. One click on **Turn on AI** takes you to the right settings and brings you back to the robot.

## Keyboard

The whole UI works without a mouse, and is built to be friendly to screen readers.

- **Tab** moves through everything in a sensible order. A **Skip to content** link appears first.
- When you move to a new screen, focus goes to its heading so a screen reader announces where you are.
- On a robot's tabs, **Left / Right** move between tabs; **Home / End** jump to the first or last.
- **Escape** closes menus and dialogs and puts focus back where it was.
- On **Controls**, switch on **Keyboard control** to drive with the keys: **W A S D** to move, **R / F** for the lift, **T / G** for the head. Let go to stop. The drive buttons also work with **Space** or **Enter** held down.
- Unsaved changes ask before you close the tab, and deleted memories can be undone for a few seconds.

## Phone and tablet

It's designed phone-first and scales up. Below about 900 px wide the side rail becomes a bottom tab bar, cards stack, buttons get larger tap targets, and the Settings sections become a list you drill into. Use it from the couch with the robot on your lap.

<p>
<img src="images/phone-home.png" alt="The robots home on a phone" width="300">
<img src="images/phone-memories-dark.png" alt="A Memories tab on a phone, in the dark theme" width="300">
</p>

Appearance (under **Settings > Appearance**) is saved on each device: light, dark or follow-the-system theme, an accent colour (or let it follow the robot's eye colour), larger text, and a retro pixel font if you miss the classic look.

## Where did the classic UI go?

Nowhere. It's at **`/classic`**, linked from the bottom of the side rail ("Open classic UI"), and every old address (`/index.html`, `/setup.html`, `/initial.html`, `/sdkapp/...`) still works, so bookmarks and scripts keep working. The classic page also has a link to try the new interface.

Anything the new UI doesn't do yet is still there: for example Bluetooth pairing and the OSKR/developer-robot setup live in the classic UI.

## For developers

The code is in `chipper/webroot/next` (see its `README.md`). No build step: edit a file, refresh the page. `js/api.js` is the only place that talks to wire-pod. Run the small logic tests with `node --test chipper/webroot/next/test/*.mjs`.
