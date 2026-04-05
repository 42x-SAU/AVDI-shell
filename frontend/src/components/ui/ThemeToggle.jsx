import { useTheme } from '@/hooks/useTheme'
import { THEME } from '@/utils/constants'
import '@/components/ui/ThemeToggle.css'

/** Иконка солнца — показывается при светлой теме (переключение на тёмную). */
function IconSun() {
  return (
    <svg
      className="theme-toggle__icon-svg"
      width={24}
      height={24}
      viewBox="0 0 24 24"
      aria-hidden="true"
      focusable="false"
    >
      <circle cx="12" cy="12" r="4" fill="none" stroke="currentColor" strokeWidth="2" />
      <path
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"
      />
    </svg>
  )
}

/** Иконка луны — показывается при тёмной теме (переключение на светлую). */
function IconMoon() {
  return (
    <svg
      className="theme-toggle__icon-svg"
      width={24}
      height={24}
      viewBox="0 0 24 24"
      aria-hidden="true"
      focusable="false"
    >
      <path
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"
      />
    </svg>
  )
}

/**
 * Переключатель светлой/тёмной темы; aria-label зависит от текущего эффективного режима.
 * @param {'default' | 'icon'} [variant] — default: трек и подпись; icon: только кнопка с солнцем/луной.
 */
export function ThemeToggle({ className = '', variant = 'default' }) {
  const { effectiveTheme, toggleTheme } = useTheme()
  const isDark = effectiveTheme === THEME.DARK

  if (variant === 'icon') {
    return (
      <button
        type="button"
        className={`theme-toggle theme-toggle--icon ${className}`.trim()}
        onClick={toggleTheme}
        aria-label={isDark ? 'Включить светлую тему' : 'Включить тёмную тему'}
        aria-pressed={isDark}
      >
        {isDark ? <IconMoon /> : <IconSun />}
      </button>
    )
  }

  return (
    <button
      type="button"
      className={`theme-toggle ${className}`.trim()}
      onClick={toggleTheme}
      aria-label={isDark ? 'Включить светлую тему' : 'Включить тёмную тему'}
      aria-pressed={isDark}
    >
      <span className="theme-toggle__track" aria-hidden="true">
        <span className="theme-toggle__thumb" />
      </span>
      <span className="theme-toggle__text">{isDark ? 'Тёмная' : 'Светлая'}</span>
    </button>
  )
}
