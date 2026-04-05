import { Navigate, Route, Routes } from 'react-router-dom'
import { AppWorkspaceLayout } from '@/layouts/AppWorkspaceLayout'
import { MainLayout } from '@/layouts/MainLayout'
import { HomePage } from '@/pages/HomePage'
import { LoginPage } from '@/pages/LoginPage'
import { LegalPersonalDataPage } from '@/pages/LegalPersonalDataPage'
import { LegalPrivacyPage } from '@/pages/LegalPrivacyPage'
import { RegisterPage } from '@/pages/RegisterPage'
import { RegisterVerifyEmailPage } from '@/pages/RegisterVerifyEmailPage'
import { AppShellPage } from '@/pages/AppShellPage'
import { ROUTES } from '@/utils/constants'

/**
 * Объявление маршрутов: новые страницы добавлять сюда и в папку pages.
 * Маршрут /app вынесен в отдельный layout без подвала — см. AppWorkspaceLayout.
 */
export function AppRouter() {
  return (
    <Routes>
      <Route element={<AppWorkspaceLayout />}>
        <Route path={ROUTES.APP_SHELL} element={<AppShellPage />} />
      </Route>
      <Route element={<MainLayout />}>
        <Route path={ROUTES.HOME} element={<HomePage />} />
        <Route path={ROUTES.LOGIN} element={<LoginPage />} />
        <Route path={ROUTES.REGISTER} element={<RegisterPage />} />
        <Route path={ROUTES.REGISTER_VERIFY_EMAIL} element={<RegisterVerifyEmailPage />} />
        <Route path={ROUTES.LEGAL_PERSONAL_DATA} element={<LegalPersonalDataPage />} />
        <Route path={ROUTES.LEGAL_PRIVACY} element={<LegalPrivacyPage />} />
        <Route path="*" element={<Navigate to={ROUTES.HOME} replace />} />
      </Route>
    </Routes>
  )
}
