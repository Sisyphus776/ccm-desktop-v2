import zh from './zh';
import en from './en';

const messages: Record<string, Record<string, string>> = { zh, en };

let currentLang = (navigator.language || 'zh').startsWith('zh') ? 'zh' : 'en';

export function t(key: string): string {
  return messages[currentLang]?.[key] || messages.en?.[key] || key;
}

export function setLang(lang: 'zh' | 'en') {
  currentLang = lang;
}

export function getLang(): string {
  return currentLang;
}
