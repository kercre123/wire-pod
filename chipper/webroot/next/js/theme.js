// Applies appearance prefs to <html>: theme, text size, font and accent color.

// Same order as the robot's eye color presets (0 to 6). `light` is used on light themes, `dark` on dark ones.
export const ACCENTS = [
  { id: "teal", name: "Teal", swatch: "#19C6B3", light: "#09716B", dark: "#3FD3C3" },
  { id: "orange", name: "Orange", swatch: "#F5A524", light: "#9E4807", dark: "#FFA45C" },
  { id: "yellow", name: "Yellow", swatch: "#F2D64B", light: "#766000", dark: "#F2D64B" },
  { id: "lime", name: "Lime", swatch: "#7BD65B", light: "#33690B", dark: "#9BE05A" },
  { id: "sapphire", name: "Blue", swatch: "#4A9BFF", light: "#0B63B5", dark: "#6CB6FF" },
  { id: "purple", name: "Purple", swatch: "#8B7CF6", light: "#6A4BC4", dark: "#B79CFF" },
  { id: "green", name: "Green", swatch: "#52D08A", light: "#16703F", dark: "#5ED08F" },
];

const root = document.documentElement;
const darkQuery = window.matchMedia ? window.matchMedia("(prefers-color-scheme: dark)") : null;

export function isDark(prefs) {
  if (prefs.theme === "dark") return true;
  if (prefs.theme === "light") return false;
  return !!(darkQuery && darkQuery.matches);
}

// What an eye color looks like on screen. `eye` is the status call's { preset, custom, hue, sat }.
export function eyeColor(eye) {
  if (!eye) return ACCENTS[0].swatch;
  if (eye.custom) return `hsl(${Math.round(eye.hue * 360)} ${Math.round(eye.sat * 100)}% 55%)`;
  return (ACCENTS[eye.preset] || ACCENTS[0]).swatch;
}

function accentFor(prefs, robotEye) {
  const dark = isDark(prefs);
  if (prefs.accentFollowsRobot && robotEye) {
    if (robotEye.custom) {
      const h = Math.round(robotEye.hue * 360);
      return dark ? { c: `hsl(${h} 75% 65%)`, ink: "#0B1517" } : { c: `hsl(${h} 70% 30%)`, ink: "#FFFFFF" };
    }
    const a = ACCENTS[robotEye.preset] || ACCENTS[0];
    return dark ? { c: a.dark, ink: "#0B1517" } : { c: a.light, ink: "#FFFFFF" };
  }
  const a = ACCENTS.find((x) => x.id === prefs.accent) || ACCENTS[0];
  return dark ? { c: a.dark, ink: "#0B1517" } : { c: a.light, ink: "#FFFFFF" };
}

export function applyTheme(prefs, robotEye) {
  if (prefs.theme === "light" || prefs.theme === "dark") root.setAttribute("data-theme", prefs.theme);
  else root.removeAttribute("data-theme");
  if (prefs.size === "large") root.setAttribute("data-size", "large");
  else root.removeAttribute("data-size");
  if (prefs.font === "retro") root.setAttribute("data-font", "retro");
  else root.removeAttribute("data-font");
  const isDefault = !prefs.accentFollowsRobot && prefs.accent === "teal";
  if (isDefault) {
    root.removeAttribute("data-accent");
    ["--accent", "--accent-ink"].forEach((p) => root.style.removeProperty(p));
  } else {
    const { c, ink } = accentFor(prefs, robotEye);
    root.setAttribute("data-accent", "custom");
    root.style.setProperty("--accent", c);
    root.style.setProperty("--accent-ink", ink);
  }
  const meta = document.querySelector('meta[name="theme-color"]');
  if (meta) meta.setAttribute("content", isDark(prefs) ? "#0B1315" : "#F2F5F5");
}

export function watchSystemTheme(cb) {
  if (!darkQuery || !darkQuery.addEventListener) return () => {};
  darkQuery.addEventListener("change", cb);
  return () => darkQuery.removeEventListener("change", cb);
}
