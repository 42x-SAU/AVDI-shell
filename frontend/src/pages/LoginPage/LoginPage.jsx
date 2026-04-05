import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { hasLoginErrors, validateLoginForm } from '@/utils/validateLoginForm'
import { ROUTES } from '@/utils/constants'
import '@/pages/LoginPage/LoginPage.css'

/**
 * Страница входа: панель с полями учётных данных (подключение API — позже).
 */
/** Демо-учётные данные для перехода в рабочую область (без backend). */
const DEMO_LOGIN = 'Admin'
const DEMO_PASSWORD = 'privet123'

export function LoginPage() {
  const navigate = useNavigate()
  const [identifier, setIdentifier] = useState('')
  const [password, setPassword] = useState('')
  /** Ошибки полей после проверки (null — ошибки нет). */
  const [fieldErrors, setFieldErrors] = useState({ identifier: null, password: null })
  /** Пароль виден только пока зажата кнопка «глаз» (мышь или касание). */
  const [revealPassword, setRevealPassword] = useState(false)

  useEffect(() => {
    if (!revealPassword) return undefined
    const stopReveal = () => setRevealPassword(false)
    window.addEventListener('mouseup', stopReveal)
    window.addEventListener('touchend', stopReveal, { passive: true })
    return () => {
      window.removeEventListener('mouseup', stopReveal)
      window.removeEventListener('touchend', stopReveal)
    }
  }, [revealPassword])

  const handleRevealStart = (e) => {
    e.preventDefault()
    setRevealPassword(true)
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    const next = validateLoginForm(identifier, password)
    setFieldErrors(next)
    if (hasLoginErrors(next)) return
    if (identifier.trim() === DEMO_LOGIN && password === DEMO_PASSWORD) {
      navigate(ROUTES.APP_SHELL)
      return
    }
    // Заглушка: отправка на backend будет добавлена при готовности API
  }

  return (
    <div className="login-page">
      <section className="login-page__hero" aria-labelledby="login-page-title">
        <div className="container login-page__inner">
          <div className="login-panel">
            <header className="login-panel__head">
              <h1 id="login-page-title" className="login-panel__title">
                С возвращением!
              </h1>
              <p className="login-panel__lead">
                Введите логин или адрес электронной почты и пароль, чтобы продолжить.
              </p>
            </header>
            <form className="login-panel__form" onSubmit={handleSubmit} noValidate>
              <div className="login-field">
                <label htmlFor="login-identifier" className="login-field__label">
                  Логин или эл. почта
                </label>
                <input
                  id="login-identifier"
                  name="identifier"
                  type="text"
                  className={`login-field__input${fieldErrors.identifier ? ' login-field__input--error' : ''}`}
                  placeholder="Andrew или tuoot@example.ru"
                  autoComplete="username"
                  value={identifier}
                  onChange={(e) => {
                    setIdentifier(e.target.value)
                    if (fieldErrors.identifier) {
                      setFieldErrors((prev) => ({ ...prev, identifier: null }))
                    }
                  }}
                  aria-invalid={Boolean(fieldErrors.identifier)}
                  aria-describedby={fieldErrors.identifier ? 'login-identifier-error' : undefined}
                />
                {fieldErrors.identifier ? (
                  <p id="login-identifier-error" className="login-field__error" role="alert">
                    {fieldErrors.identifier}
                  </p>
                ) : null}
              </div>
              <div className="login-field">
                <label htmlFor="login-password" className="login-field__label">
                  Пароль
                </label>
                <div className="login-field__control">
                  <input
                    id="login-password"
                    name="password"
                    type={revealPassword ? 'text' : 'password'}
                    className={`login-field__input login-field__input--with-reveal${fieldErrors.password ? ' login-field__input--error' : ''}`}
                    placeholder="Минимум 8 символов"
                    autoComplete="current-password"
                    value={password}
                    onChange={(e) => {
                      setPassword(e.target.value)
                      if (fieldErrors.password) {
                        setFieldErrors((prev) => ({ ...prev, password: null }))
                      }
                    }}
                    aria-invalid={Boolean(fieldErrors.password)}
                    aria-describedby={fieldErrors.password ? 'login-password-error' : undefined}
                  />
                  <button
                    type="button"
                    className="login-field__reveal"
                    aria-label="Показать пароль при удержании"
                    onMouseDown={handleRevealStart}
                    onTouchStart={handleRevealStart}
                  >
                    <svg
                      className="login-field__reveal-icon"
                      xmlns="http://www.w3.org/2000/svg"
                      width="20"
                      height="20"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      aria-hidden
                    >
                      <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                      <circle cx="12" cy="12" r="3" />
                    </svg>
                  </button>
                </div>
                {fieldErrors.password ? (
                  <p id="login-password-error" className="login-field__error" role="alert">
                    {fieldErrors.password}
                  </p>
                ) : null}
                <div className="login-field__row">
                  <button type="button" className="login-forgot">
                    Забыли пароль?
                  </button>
                </div>
              </div>
              <div className="login-panel__actions">
                <button type="submit" className="btn btn--primary login-panel__submit">
                  Продолжить
                </button>
              </div>
              <p className="login-panel__register-line">
                Нет аккаунта?{' '}
                <Link to={ROUTES.REGISTER} className="login-panel__register-link">
                  Зарегистрируйтесь!
                </Link>
              </p>
            </form>
            <div className="login-panel__alt">
              <div className="login-divider">
                <span className="login-divider__line" aria-hidden />
                <span className="login-divider__text">или</span>
                <span className="login-divider__line" aria-hidden />
              </div>
              <div className="login-panel__alt-actions">
                <button
                  type="button"
                  className="btn btn--ghost login-panel__alt-btn login-panel__alt-btn--with-soon"
                >
                  <span className="login-panel__alt-btn-label">Войти с помощью Госуслуг</span>
                  <span className="login-panel__soon-badge">Скоро</span>
                </button>
                <button
                  type="button"
                  className="btn btn--ghost login-panel__alt-btn login-panel__alt-btn--with-soon"
                >
                  <span className="login-panel__alt-btn-label">Войти как организация</span>
                  <span className="login-panel__soon-badge">Скоро</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}
