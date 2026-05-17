export const themes = [
  { id: 'lacquer', name: 'Lacquer 漆器', nameCN: '漆器' },
  { id: 'alabaster', name: 'Alabaster 石膏', nameCN: '雪花石膏' },
  { id: 'slate', name: 'Slate 钛空', nameCN: '钛空' },
  { id: 'photon', name: 'Photon 光量子', nameCN: '光量子' },
  { id: 'obsidian', name: 'Obsidian 黑曜', nameCN: '黑曜' },
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
