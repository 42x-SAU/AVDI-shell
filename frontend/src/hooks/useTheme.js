import { useContext } from 'react'
import { ThemeContext } from '@/context/themeContext'

/** Хук темы: обёртка над ThemeContext для удобного импорта из @/hooks */
export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) {
    throw new Error('useTheme должен вызываться внутри ThemeProvider')
  }
  return ctx
}
