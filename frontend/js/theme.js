const THEME_KEY = 'theme_preference';

function getCurrentTheme() {
  const saved = localStorage.getItem(THEME_KEY);
  if (saved) return saved;
  
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function setTheme(theme) {
  localStorage.setItem(THEME_KEY, theme);
  applyTheme(theme);
}

// Make toggleTheme global
window.toggleTheme = function() {
  const current = getCurrentTheme();
  const newTheme = current === 'dark' ? 'light' : 'dark';
  setTheme(newTheme);
  
  if (typeof plausible !== 'undefined') {
    plausible('Theme Toggle', { props: { theme: newTheme } });
  }
}

function applyTheme(theme) {
  document.documentElement.setAttribute('data-theme', theme);
  
  const themeToggle = document.getElementById('theme-toggle');
  if (themeToggle) {
    themeToggle.textContent = theme === 'dark' ? 'Light Mode' : 'Dark Mode';
  }
}

// Initialize theme on page load
applyTheme(getCurrentTheme());

// Listen for system preference changes
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
  if (!localStorage.getItem(THEME_KEY)) {
    applyTheme(e.matches ? 'dark' : 'light');
  }
});