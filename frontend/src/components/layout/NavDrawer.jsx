import { useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { Link } from 'react-router-dom'
import { AppSettingsMenu } from '@/components/ui/AppSettingsMenu'
import { ROUTES } from '@/utils/constants'
import '@/components/layout/NavDrawer.css'

/**
 * Выезжающая панель навигации (мобильное и десктопное бургер-меню).
 * Закрытие по Escape и клику по подложке.
 */
export function NavDrawer({ id, isOpen, onClose }) {
  const panelRef = useRef(null)

  useEffect(() => {
    if (!isOpen) return undefined
    const onKey = (e) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKey)
    const prevOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      document.removeEventListener('keydown', onKey)
      document.body.style.overflow = prevOverflow
    }
  }, [isOpen, onClose])

  useEffect(() => {
    if (isOpen && panelRef.current) {
      const link = panelRef.current.querySelector('a, button')
      if (link) link.focus()
    }
  }, [isOpen])

  /* Портал в body: иначе backdrop-filter у предка (шапка) ломает position:fixed и весь оверлей */
  const drawer = (
    <div
      className={`nav-drawer ${isOpen ? 'nav-drawer--open' : ''}`}
      aria-hidden={!isOpen}
    >
      <button
        type="button"
        className="nav-drawer__backdrop"
        aria-label="Закрыть меню"
        tabIndex={isOpen ? 0 : -1}
        onClick={onClose}
      />
      <aside
        ref={panelRef}
        id={id}
        className="nav-drawer__panel"
        role="dialog"
        aria-modal="true"
        aria-label="Навигация по разделам"
      >
        <div className="nav-drawer__head">
          <span className="nav-drawer__title">Меню</span>
          <button type="button" className="nav-drawer__close" onClick={onClose} aria-label="Закрыть меню">
            ×
          </button>
        </div>
        <nav className="nav-drawer__nav" aria-label="Основная навигация">
          <Link to={ROUTES.HOME} className="nav-drawer__link" onClick={onClose}>
            Главная
          </Link>
          <span
            className="nav-drawer__link nav-drawer__link--disabled nav-drawer__link--with-soon"
            aria-disabled="true"
          >
            <span className="nav-drawer__link-label">Админ-панель</span>
            <span className="nav-drawer__soon">скоро</span>
          </span>
        </nav>
        <div className="nav-drawer__footer">
          {/* Плашка приложения над входом */}
          <div className="nav-drawer__app-promo">
            <div className="nav-drawer__app-promo-icon" aria-hidden>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="22"
                height="22"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
                <path d="M8 21h8" />
                <path d="M12 17v4" />
              </svg>
            </div>
            <div className="nav-drawer__app-promo-copy">
              <span className="nav-drawer__app-promo-title">AVDI App</span>
              <span className="nav-drawer__app-promo-subtitle">
                Удобный сервис с агентом - теперь в приложении
              </span>
            </div>
            <button
              type="button"
              className="nav-drawer__app-install"
              disabled
              aria-label="Установить приложение"
            >
              Установить
            </button>
          </div>
          <div className="nav-drawer__footer-row">
            <Link to={ROUTES.LOGIN} className="nav-drawer__login" onClick={onClose}>
              Войти в аккаунт
            </Link>
            <AppSettingsMenu className="nav-drawer__settings-btn" />
          </div>
        </div>
      </aside>
    </div>
  )

  return typeof document !== 'undefined' ? createPortal(drawer, document.body) : null
}
