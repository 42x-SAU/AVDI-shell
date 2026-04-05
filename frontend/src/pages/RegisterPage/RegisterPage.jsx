import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ROUTES } from '@/utils/constants'
import { getPasswordStrength } from '@/utils/passwordStrength'
import '@/pages/RegisterPage/RegisterPage.css'

/**
 * Страница регистрации: панель «Создание аккаунта» и поля формы.
 */
export function RegisterPage() {
  const navigate = useNavigate()
  const [login, setLogin] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [passwordRepeat, setPasswordRepeat] = useState('')
  /** Показ пароля по нажатию на «глаз» (не удержание). */
  const [showPassword, setShowPassword] = useState(false)
  const [showPasswordRepeat, setShowPasswordRepeat] = useState(false)
  const [consentAccepted, setConsentAccepted] = useState(false)

  const stopPasteAndDrop = (e) => {
    e.preventDefault()
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    if (!consentAccepted) return
    // Заглушка: отправка на backend — позже; затем переход к подтверждению почты
    navigate(ROUTES.REGISTER_VERIFY_EMAIL, { state: { email: email.trim() } })
  }

  const passwordStrength = getPasswordStrength(password)

  return (
    <div className="register-page">
      <section className="register-page__hero" aria-labelledby="register-panel-title">
        <div className="container register-page__inner">
          <div className="register-panel">
            <div className="register-panel__toolbar">
              <Link to={ROUTES.LOGIN} className="register-panel__back">
                <svg
                  className="register-panel__back-icon"
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
                  <path d="M19 12H5M12 19l-7-7 7-7" />
                </svg>
                Назад
              </Link>
            </div>
            <h1 id="register-panel-title" className="register-panel__title">
              Создание аккаунта
            </h1>

            <form className="register-panel__form" onSubmit={handleSubmit} noValidate>
              <div className="register-field">
                <label htmlFor="register-login" className="register-field__label">
                  Логин
                </label>
                <input
                  id="register-login"
                  name="login"
                  type="text"
                  className="register-field__input"
                  placeholder="Не более 10 символов"
                  maxLength={10}
                  autoComplete="username"
                  value={login}
                  onChange={(e) => setLogin(e.target.value)}
                />
              </div>

              <div className="register-field">
                <label htmlFor="register-email" className="register-field__label">
                  Эл. почта
                </label>
                <input
                  id="register-email"
                  name="email"
                  type="email"
                  className="register-field__input"
                  placeholder="yourname@example.ru"
                  autoComplete="email"
                  inputMode="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>

              <div className="register-field">
                <label htmlFor="register-phone" className="register-field__label">
                  Номер телефона <span className="register-field__optional">(необязательно)</span>
                </label>
                <input
                  id="register-phone"
                  name="phone"
                  type="tel"
                  className="register-field__input"
                  autoComplete="tel"
                  inputMode="tel"
                  placeholder="+7"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                />
              </div>

              <div className="register-field">
                <label htmlFor="register-password" className="register-field__label">
                  Пароль
                </label>
                <div className="register-field__control">
                  <input
                    id="register-password"
                    name="password"
                    type={showPassword ? 'text' : 'password'}
                    className="register-field__input register-field__input--with-reveal"
                    placeholder="Придумайте надежный пароль"
                    autoComplete="new-password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                  <button
                    type="button"
                    className="register-field__reveal"
                    aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'}
                    aria-pressed={showPassword}
                    onClick={() => setShowPassword((v) => !v)}
                  >
                    <svg
                      className="register-field__reveal-icon"
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
                <div
                  className="register-password-strength"
                  role="status"
                  aria-live="polite"
                  aria-label={
                    passwordStrength.level === 'empty'
                      ? 'Надежность пароля'
                      : `Надежность пароля: ${passwordStrength.label}`
                  }
                >
                  <div className="register-password-strength__track">
                    <div
                      className={`register-password-strength__fill register-password-strength__fill--${passwordStrength.level}`}
                    />
                  </div>
                  {passwordStrength.level !== 'empty' ? (
                    <span
                      className={`register-password-strength__label register-password-strength__label--${passwordStrength.level}`}
                    >
                      {passwordStrength.label}
                    </span>
                  ) : null}
                </div>
              </div>

              <div className="register-field">
                <label htmlFor="register-password-repeat" className="register-field__label">
                  Повторите пароль
                </label>
                <div className="register-field__control">
                  <input
                    id="register-password-repeat"
                    name="passwordRepeat"
                    type={showPasswordRepeat ? 'text' : 'password'}
                    className="register-field__input register-field__input--with-reveal"
                    placeholder="Даже не пытайтесь здесь схитрить"
                    autoComplete="new-password"
                    value={passwordRepeat}
                    onChange={(e) => setPasswordRepeat(e.target.value)}
                    onPaste={stopPasteAndDrop}
                    onDrop={stopPasteAndDrop}
                    onDragOver={stopPasteAndDrop}
                  />
                  <button
                    type="button"
                    className="register-field__reveal"
                    aria-label={showPasswordRepeat ? 'Скрыть пароль' : 'Показать пароль'}
                    aria-pressed={showPasswordRepeat}
                    onClick={() => setShowPasswordRepeat((v) => !v)}
                  >
                    <svg
                      className="register-field__reveal-icon"
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
              </div>

              <div className="register-consent">
                <input
                  id="register-consent"
                  name="consent"
                  type="checkbox"
                  className="register-consent__checkbox"
                  checked={consentAccepted}
                  onChange={(e) => setConsentAccepted(e.target.checked)}
                />
                <label htmlFor="register-consent" className="register-consent__text">
                  Я ознакомлен(-а) с{' '}
                  <Link to={ROUTES.LEGAL_PERSONAL_DATA} className="register-consent__link">
                    Правилами обработки персональных данных
                  </Link>{' '}
                  и{' '}
                  <Link to={ROUTES.LEGAL_PRIVACY} className="register-consent__link">
                    Политикой конфиденциальности
                  </Link>
                  .
                </label>
              </div>

              <div className="register-panel__actions">
                <button
                  type="submit"
                  className="btn btn--primary register-panel__submit"
                  disabled={!consentAccepted}
                >
                  Далее
                </button>
              </div>
            </form>
          </div>
        </div>
      </section>
    </div>
  )
}
