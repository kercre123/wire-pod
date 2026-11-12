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
