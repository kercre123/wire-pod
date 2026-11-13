// Quick actions. Vector's own behaviours (explore, dance, sleep...) cannot run while the SDK holds behavior
// control, so those hand control back first. "Go to charger" has a direct call that works either way.
import { sdk, describeError } from "./api.js";

export const QUICK = [
  { label: "Explore", intent: "explore_start", verb: "explore" },
  { label: "Dance", intent: "intent_imperative_dance", verb: "dance" },
  { label: "Fetch cube", intent: "intent_imperative_fetchcube", verb: "fetch the cube" },
  { label: "Go to charger", intent: "intent_system_charger", verb: "go to the charger", held: "drive_on_charger" },
  { label: "Sleep", intent: "intent_system_sleep", verb: "go to sleep", icon: "moon" },
];

const wait = (ms) => new Promise((r) => setTimeout(r, ms));

// deps: { esn, name, controlled, release, toast, call = sdk, pause = wait }
export async function runQuick(q, { esn, name, controlled, release, toast, call = sdk, pause = wait }) {
  try {
    if (controlled && q.held) {
      const res = JSON.parse(await call(q.held, esn));
      if (res.result === "BEHAVIOR_COMPLETE_STATE") toast(`${q.label}: sent to ${name}`);
      else toast(`${name} can’t ${q.verb} right now. It may already be there.`, { kind: "bad" });
      return;
    }
    if (controlled) {
      await release();
      toast(`Released controls so ${name} can ${q.verb}`);
      await pause(800); // the release reaches the robot within half a second
    }
    await call("cloud_intent", esn, { intent: q.intent });
    toast(`${q.label}: sent to ${name}`);
  } catch (e) {
    toast(`${name} didn’t take “${q.label}”. ${describeError(e, name).message}`, { kind: "bad" });
  }
}
