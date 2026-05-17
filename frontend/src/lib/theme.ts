export const themes = [
  { id: 'oled', name: 'OLED Dark', nameCN: 'OLED 暗黑' },
  { id: 'frost', name: 'Frost', nameCN: '冰霜蓝' },
  { id: 'sepia', name: 'Sepia', nameCN: '暖褐纸' },
  { id: 'monochrome', name: 'Monochrome', nameCN: '极简灰' },
  { id: 'neon', name: 'Neon', nameCN: '霓虹夜' },
] as const;

export type ThemeId = typeof themes[number]['id'];

export function getTheme(): ThemeId {
  return (localStorage.getItem('ccm-theme') as ThemeId) || 'oled';
}

export function setTheme(id: ThemeId) {
  localStorage.setItem('ccm-theme', id);
  if (id === 'oled') {
    document.documentElement.removeAttribute('data-theme');
  } else {
    document.documentElement.setAttribute('data-theme', id);
  }
}

export function initTheme() {
  const theme = getTheme();
  if (theme === 'oled') {
    document.documentElement.removeAttribute('data-theme');
  } else {
    document.documentElement.setAttribute('data-theme', theme);
  }
}
