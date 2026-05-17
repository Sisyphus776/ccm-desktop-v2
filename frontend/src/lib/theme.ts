export const themes = [
  { id: 'lacquer', name: 'Lacquer 漆器', nameCN: '漆器' },
  { id: 'alabaster', name: 'Alabaster 石膏', nameCN: '雪花石膏' },
  { id: 'verdigris', name: 'Verdigris 铜绿', nameCN: '铜绿' },
  { id: 'noir-rose', name: 'Noir Rose 暗玫', nameCN: '暗玫瑰' },
  { id: 'ceramic', name: 'Ceramic 陶瓷', nameCN: '陶瓷' },
] as const;

export type ThemeId = typeof themes[number]['id'];

export function getTheme(): ThemeId {
  return (localStorage.getItem('ccm-theme') as ThemeId) || 'lacquer';
}

export function setTheme(id: ThemeId) {
  localStorage.setItem('ccm-theme', id);
  if (id === 'lacquer') {
    document.documentElement.removeAttribute('data-theme');
  } else {
    document.documentElement.setAttribute('data-theme', id);
  }
}

export function initTheme() {
  const theme = getTheme();
  if (theme === 'lacquer') {
    document.documentElement.removeAttribute('data-theme');
  } else {
    document.documentElement.setAttribute('data-theme', theme);
  }
}
