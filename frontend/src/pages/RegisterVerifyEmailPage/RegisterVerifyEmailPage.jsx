import { useState } from 'react'
import { Link, Navigate, useLocation } from 'react-router-dom'
import { ROUTES } from '@/utils/constants'
import { maskEmailForDisplay } from '@/utils/maskEmail'
import '@/pages/RegisterPage/RegisterPage.css'
import '@/pages/RegisterVerifyEmailPage/RegisterVerifyEmailPage.css'

/**
 * Шаг после регистрации: ввод кода из письма (email передаётся через state навигации).
 */
export function RegisterVerifyEmailPage() {
  const location = useLocation()
  const emailFromState = location.state?.email
  const [code, setCode] = useState('')

  if (!emailFromState || typeof emailFromState !== 'string') {
    return <Navigate to={ROUTES.REGISTER} replace />
  }

  const maskedEmail = maskEmailForDisplay(emailFromState)

  const handleSubmit = (e) => {
    e.preventDefault()
    // Заглушка: проверка кода на backend — позже
  }

  return (
    <div className="register-page">
      <section className="register-page__hero" aria-labelledby="verify-email-title">
        <div className="container register-page__inner">
          <div className="register-panel register-verify">
            <h1 id="verify-email-title" className="register-panel__title register-verify__title">
              Проверьте почту
            </h1>
            <p className="register-verify__lead">
              На электронную почту <span className="register-verify__email">{maskedEmail}</span> был
              выслан 6-значный код подтверждения. Не забудьте проверить папку «Спам».
            </p>

            <form className="register-panel__form register-verify__form" onSubmit={handleSubmit} noValidate>
              <div className="register-field">
                <label htmlFor="verify-email-code" className="register-field__label">
                  Код подтверждения
                </label>
                <input
                  id="verify-email-code"
                  name="code"
                  type="text"
                  className="register-field__input register-verify__code-input"
                  placeholder="123456"
                  inputMode="numeric"
                  autoComplete="one-time-code"
                  maxLength={6}
                  value={code}
                  onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                />
              </div>

              <div className="register-verify__row">
                <button type="button" className="register-verify__link-btn">
                  Не пришел код подтверждения?
                </button>
              </div>

              <div className="register-panel__actions">
                <Link to={ROUTES.LOGIN} className="btn btn--primary register-panel__submit">
                  Зарегистрироваться
                </Link>
              </div>
            </form>
          </div>
        </div>
      </section>
    </div>
  )
}
