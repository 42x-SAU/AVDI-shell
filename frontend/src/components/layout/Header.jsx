import { useId, useState } from 'react'
import { Link } from 'react-router-dom'
import logoImg from '@/assets/images/avdi-full.png'
import { NavDrawer } from '@/components/layout/NavDrawer'
import { ROUTES } from '@/utils/constants'
import '@/components/layout/Header.css'

/**
 * Шапка: бренд слева, кнопка меню справа. Переключение темы — в выезжающем меню.
 */
export function Header() {
  const [menuOpen, setMenuOpen] = useState(false)
  const drawerId = useId()

  return (
    <header className="site-header">
      <div className="site-header__bar container">
        <Link
          to={ROUTES.HOME}
          className="site-header__brand"
          aria-label="AVDI Shell — на главную страницу"
        >
          <img
            src={logoImg}
            alt=""
            width={171}
            height={44}
            className="site-header__logo"
            decoding="async"
          />
        </Link>
        <div className="site-header__actions">
          <button
            type="button"
            className="site-header__burger"
            aria-label="Открыть меню навигации"
            aria-expanded={menuOpen}
            aria-controls={drawerId}
            onClick={() => setMenuOpen(true)}
          >
            <span className="site-header__burger-line" aria-hidden="true" />
            <span className="site-header__burger-line" aria-hidden="true" />
            <span className="site-header__burger-line" aria-hidden="true" />
          </button>
        </div>
      </div>
      <NavDrawer id={drawerId} isOpen={menuOpen} onClose={() => setMenuOpen(false)} />
    </header>
  )
}
