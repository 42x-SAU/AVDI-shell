import { ThemeProvider } from '@/app/ThemeProvider'

/**
 * Корневые провайдеры приложения (тема и т.д.).
 * Роутер подключается в main.jsx, чтобы NavigationProgress имел доступ к контексту роутера.
 */
export function AppProviders({ children }) {
  return <ThemeProvider>{children}</ThemeProvider>
}
