export const themes = [
  { id: 'oled', name: 'OLED Dark', nameCN: 'OLED 暗黑' },
  { id: 'light', name: 'Professional Light', nameCN: '专业亮白' },
  { id: 'paper', name: 'Warm Paper', nameCN: '暖纸米' },
  { id: 'midnight', name: 'Midnight Blue', nameCN: '午夜蓝' },
  { id: 'mineral', name: 'Mineral', nameCN: '矿物灰绿' },
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
