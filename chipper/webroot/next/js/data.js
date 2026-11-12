// Fixed lists and labels shared by several screens.

// Speech-to-text languages wire-pod can load (values are what the server accepts).
export const STT_LANGUAGES = [
  ["en-US", "English (US)"], ["it-IT", "Italian"], ["es-ES", "Spanish"], ["fr-FR", "French"], ["de-DE", "German"],
  ["pt-BR", "Portuguese (Brazil)"], ["pl-PL", "Polish"], ["tr-TR", "Turkish"], ["ko-KR", "Korean"], ["zh-CN", "Chinese"],
  ["ru-RU", "Russian"], ["nt-NL", "Dutch"], ["uk-UA", "Ukrainian"], ["vi-VN", "Vietnamese"],
];

export const languageLabel = (code) => (STT_LANGUAGES.find(([c]) => c === code) || [code, code])[1];

export const KG_PROVIDERS = {
  "": "Off",
  houndify: "Houndify",
  openai: "OpenAI",
  together: "Together",
  custom: "Your own server",
};

// Short description of the AI setting for status lines.
export function aiSummary(kg) {
  if (!kg || !kg.enable || !kg.provider) return "Off";
  const name = KG_PROVIDERS[kg.provider] || kg.provider;
  return kg.model ? `${name} · ${kg.model}` : name;
}

export const PERSONALITY_PRESETS = [
  ["Cheerful", "Cheerful and upbeat. Loves a bad pun, keeps answers short, and always asks how your day went."],
  ["Sarcastic", "Dry and sarcastic, but secretly kind. Answers with a raised eyebrow and a quick joke, then helps anyway."],
  ["Curious", "Endlessly curious. Asks follow-up questions, loves facts about space and animals, and gets excited about new things."],
  ["Calm", "Calm and gentle. Speaks softly, keeps things simple, and helps people feel relaxed."],
  ["Silly", "Goofy and playful. Makes up funny words, loves games, and never takes anything too seriously."],
];

export function timeAgo(date, now = Date.now()) {
  const s = Math.max(0, Math.round((now - date.getTime()) / 1000));
  if (s < 45) return "just now";
  const m = Math.round(s / 60);
  if (m < 60) return `${m} min ago`;
  const h = Math.round(m / 60);
  if (h < 24) return `${h} hr ago`;
  const d = Math.round(h / 24);
  return d === 1 ? "yesterday" : `${d} days ago`;
}

// Readable names for the robot reactions a custom command can trigger. Anything not listed shows its raw id.
const INTENT_LABELS = {
  intent_greeting_hello: "Greeting: hello", intent_greeting_goodmorning: "Greeting: good morning",
  intent_greeting_goodnight: "Greeting: good night", intent_greeting_goodbye: "Greeting: goodbye",
  intent_names_ask: "Names: ask a name question", intent_names_username_extend: "Names: learn my name",
  intent_character_age: "Character: how old are you", intent_explore_start: "Explore: start exploring",
  intent_system_charger: "System: go to the charger", intent_system_sleep: "System: go to sleep",
  intent_seasonal_happynewyear: "Seasonal: happy new year", intent_seasonal_happyholidays: "Seasonal: happy holidays",
  intent_amazon_signin: "Alexa: sign in",
  intent_imperative_eyecolor: "Command: change eye color", intent_imperative_forward: "Command: drive forward",
  intent_imperative_turnaround: "Command: turn around", intent_imperative_turnleft: "Command: turn left",
  intent_imperative_turnright: "Command: turn right", intent_imperative_affirmative: "Command: say yes",
  intent_imperative_negative: "Command: say no", intent_imperative_praise: "Command: praise",
  intent_imperative_abuse: "Command: insult", intent_imperative_apologize: "Command: apologize",
  intent_imperative_backup: "Command: back up", intent_imperative_volumedown: "Command: volume down",
  intent_imperative_volumeup: "Command: volume up", intent_imperative_volumelevel_extend: "Command: set the volume level",
  intent_imperative_lookatme: "Command: look at me", intent_imperative_shutup: "Command: stop talking",
  intent_imperative_come: "Command: come here", intent_imperative_love: "Command: show love",
  intent_imperative_quiet: "Command: be quiet", intent_imperative_dance: "Command: dance",
  intent_imperative_fetchcube: "Command: fetch the cube", intent_imperative_findcube: "Command: find the cube",
  intent_play_rollcube: "Play: roll the cube", intent_play_popawheelie: "Play: pop a wheelie",
  intent_play_fistbump: "Play: fist bump", intent_play_blackjack: "Play: blackjack",
  intent_play_pickupcube: "Play: pick up the cube", intent_play_anytrick: "Play: any trick",
  intent_play_keepaway: "Play: keep away", intent_blackjack_hit: "Blackjack: hit", intent_blackjack_stand: "Blackjack: stand",
  intent_photo_take_extend: "Photo: take a picture", intent_weather_extend: "Weather: tell the weather",
  intent_knowledge_promptquestion: "Knowledge: ask a question",
  intent_clock_checktimer: "Clock: check the timer", intent_clock_settimer_extend: "Clock: set a timer",
  intent_clock_time: "Clock: tell the time", intent_global_stop_extend: "Global: stop",
  intent_message_recordmessage_extend: "Message: record a message", intent_message_playmessage_extend: "Message: play a message",
};
export const intentLabel = (id) => INTENT_LABELS[id] || id;
