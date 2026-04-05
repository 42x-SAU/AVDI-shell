import { useCallback, useEffect, useMemo, useState } from 'react'
import { ThemeContext } from '@/context/themeContext'
import { STORAGE_KEYS, THEME } from '@/utils/constants'

/**
 * Определяет эффективную тему: при «system» учитывается prefers-color-scheme.
 */
function resolveEffectiveTheme(stored) {
  if (stored === THEME.LIGHT || stored === THEME.DARK) return stored
  if (typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    return THEME.DARK
  }
  return THEME.LIGHT
}

export function ThemeProvider({ children }) {
  const [theme, setThemeState] = useState(() => {
    if (typeof window === 'undefined') return THEME.SYSTEM
    try {
      return localStorage.getItem(STORAGE_KEYS.THEME) || THEME.SYSTEM
    } catch {
      return THEME.SYSTEM
    }
  })

  const effectiveTheme = useMemo(() => resolveEffectiveTheme(theme), [theme])

  useEffect(() => {
    document.documentElement.dataset.theme = effectiveTheme
  }, [effectiveTheme])

  useEffect(() => {
    if (theme !== THEME.SYSTEM) return undefined
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const handler = () => {
      document.documentElement.dataset.theme = resolveEffectiveTheme(THEME.SYSTEM)
    }
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [theme])

  const setTheme = useCallback((next) => {
    setThemeState(next)
    try {
      localStorage.setItem(STORAGE_KEYS.THEME, next)
    } catch {
      /* игнорируем недоступность localStorage */
    }
  }, [])

  const toggleTheme = useCallback(() => {
    setThemeState((prev) => {
      const eff = resolveEffectiveTheme(prev)
      const next = eff === THEME.DARK ? THEME.LIGHT : THEME.DARK
      try {
        localStorage.setItem(STORAGE_KEYS.THEME, next)
      } catch {
        /* пусто */
      }
      return next
    })
  }, [])

  const value = useMemo(
    () => ({
      theme,
      effectiveTheme,
      setTheme,
      toggleTheme,
    }),
    [theme, effectiveTheme, setTheme, toggleTheme],
  )

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}
