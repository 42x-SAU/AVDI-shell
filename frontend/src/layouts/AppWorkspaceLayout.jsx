import { Outlet } from 'react-router-dom'
import { NavigationProgress } from '@/components/common/NavigationProgress'
import { Header } from '@/components/layout/Header'
import '@/layouts/AppWorkspaceLayout.css'

/**
 * Каркас рабочей зоны (/app): общая шапка сайта без маркетингового подвала,
 * чтобы панель мониторинга занимала доступную высоту экрана.
 */
export function AppWorkspaceLayout() {
  return (
    <>
      <a href="#app-workspace-main" className="skip-link">
        Перейти к содержимому панели
      </a>
      <NavigationProgress />
      <Header />
      <main className="app-workspace__main" id="app-workspace-main" role="main">
        <Outlet />
      </main>
    </>
  )
}
