import { ref, watch, onMounted } from 'vue'

const STORAGE_KEY = 'eregen-admin-theme'

export function useTheme() {
  const isDark = ref(false)

  function getInitialTheme(): boolean {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored !== null) return stored === 'dark'
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  }

  function applyTheme(dark: boolean) {
    isDark.value = dark
    document.documentElement.setAttribute('data-theme', dark ? 'dark' : 'light')
    localStorage.setItem(STORAGE_KEY, dark ? 'dark' : 'light')
  }

  function toggle() {
    applyTheme(!isDark.value)
  }

  onMounted(() => {
    applyTheme(getInitialTheme())
  })

  watch(isDark, (dark) => {
    document.documentElement.setAttribute('data-theme', dark ? 'dark' : 'light')
  })

  return { isDark, toggle, applyTheme }
}
