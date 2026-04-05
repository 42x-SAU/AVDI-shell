import { Outlet } from 'react-router-dom'
import { NavigationProgress } from '@/components/common/NavigationProgress'
import { Footer } from '@/components/layout/Footer'
import { Header } from '@/components/layout/Header'
import '@/layouts/MainLayout.css'

/**
 * Общий каркас SPA: шапка, контент страниц, подвал.
 */
export function MainLayout() {
  return (
    <>
      <a href="#main-content" className="skip-link">
        Перейти к содержимому
      </a>
      <NavigationProgress />
      <Header />
      <main className="main-layout__content" id="main-content" role="main">
        <Outlet />
      </main>
      <Footer />
    </>
  )
}
