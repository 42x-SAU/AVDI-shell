import { AppProviders } from '@/app/AppProviders'
import { AppRouter } from '@/router/AppRouter'

/**
 * Корневой компонент: провайдеры и конфигурация маршрутов.
 */
export default function App() {
  return (
    <AppProviders>
      <AppRouter />
    </AppProviders>
  )
}
