import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import type { ThemeMode } from '@/api/types'

const THEME_KEY = 'bv_theme'
const MOTION_KEY = 'bv_motion'

function systemTheme(): ThemeMode {
  if (typeof window === 'undefined') return 'light'

  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function storedTheme(): ThemeMode | null {
  if (typeof window === 'undefined') return null

  const value = window.localStorage.getItem(THEME_KEY)

  return value === 'dark' || value === 'light' ? value : null
}

export const useThemeStore = defineStore('theme', () => {
  const theme = ref<ThemeMode>(storedTheme() ?? systemTheme())
  const reducedMotion = ref(
    typeof window !== 'undefined'
      ? window.localStorage.getItem(MOTION_KEY) === 'reduced'
      : false,
  )
  const isDark = computed(() => theme.value === 'dark')

  function apply(): void {
    if (typeof document === 'undefined') return

    document.documentElement.classList.toggle('dark', theme.value === 'dark')
    document.documentElement.dataset.theme = theme.value
    document.documentElement.dataset.motion = reducedMotion.value ? 'reduced' : 'full'
    document
      .querySelector('meta[name="theme-color"]')
      ?.setAttribute('content', theme.value === 'dark' ? '#0d0d0b' : '#ffcc00')
  }

  function setTheme(value: ThemeMode): void {
    theme.value = value
    window.localStorage.setItem(THEME_KEY, value)
    apply()
  }

  function toggle(): void {
    setTheme(theme.value === 'dark' ? 'light' : 'dark')
  }

  function setReducedMotion(value: boolean): void {
    reducedMotion.value = value
    window.localStorage.setItem(MOTION_KEY, value ? 'reduced' : 'full')
    apply()
  }

  function init(): void {
    apply()

    const media = window.matchMedia('(prefers-color-scheme: dark)')
    media.addEventListener('change', (event) => {
      if (!window.localStorage.getItem(THEME_KEY)) {
        theme.value = event.matches ? 'dark' : 'light'
        apply()
      }
    })
  }

  return { theme, isDark, reducedMotion, init, setTheme, toggle, setReducedMotion }
})
