import { useEffect, useId, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { useTheme } from '@/hooks/useTheme'
import { THEME } from '@/utils/constants'
import '@/components/ui/AppSettingsMenu.css'

/**
 * Кнопка открывает меню: переключение светлой/тёмной темы.
 */
export function AppSettingsMenu({ className = '' }) {
  const menuId = useId()
  const btnRef = useRef(null)
  const popoverRef = useRef(null)
  const [open, setOpen] = useState(false)
  const [pos, setPos] = useState({ top: 0, left: 0 })

  const { effectiveTheme, setTheme } = useTheme()
  const isDark = effectiveTheme === THEME.DARK

  useEffect(() => {
    if (!open) return undefined
    const onKey = (e) => {
      if (e.key === 'Escape') setOpen(false)
    }
    const onDown = (e) => {
      if (
        popoverRef.current?.contains(e.target) ||
        btnRef.current?.contains(e.target)
      ) {
        return
      }
      setOpen(false)
    }
    document.addEventListener('keydown', onKey)
    document.addEventListener('mousedown', onDown)
    return () => {
      document.removeEventListener('keydown', onKey)
      document.removeEventListener('mousedown', onDown)
    }
  }, [open])

  const computePopoverPos = () => {
    if (!btnRef.current) return { top: 8, left: 8 }
    const r = btnRef.current.getBoundingClientRect()
    const w = 260
    let left = r.right - w
    if (left < 8) left = 8
    if (left + w > window.innerWidth - 8) left = window.innerWidth - w - 8
    let top = r.bottom + 8
    if (top + 200 > window.innerHeight - 8) {
      top = Math.max(8, r.top - 8 - 200)
    }
    return { top, left }
  }

  const toggleThemeUi = () => {
    setTheme(isDark ? THEME.LIGHT : THEME.DARK)
  }

  return (
    <div className={`app-settings-menu ${className}`.trim()}>
      <button
        ref={btnRef}
        type="button"
        className="app-settings-menu__trigger"
        aria-expanded={open}
        aria-haspopup="true"
        aria-controls={open ? menuId : undefined}
        onClick={() => {
          if (open) {
            setOpen(false)
            return
          }
          setPos(computePopoverPos())
          setOpen(true)
        }}
        aria-label="Настройки отображения и темы"
      >
        <svg
          className="app-settings-menu__trigger-icon"
          width="22"
          height="22"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          aria-hidden
        >
          <circle cx="12" cy="12" r="3" />
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
        </svg>
      </button>

      {open &&
        typeof document !== 'undefined' &&
        createPortal(
          <div
            ref={popoverRef}
            id={menuId}
            className="app-settings-menu__popover"
            role="dialog"
            aria-label="Настройки отображения"
            style={{ top: pos.top, left: pos.left }}
          >
            <div className="app-settings-menu__row">
              <span className="app-settings-menu__label" id={`${menuId}-theme`}>
                Тёмная тема
              </span>
              <button
                type="button"
                className={`app-settings-menu__switch ${isDark ? 'app-settings-menu__switch--on' : ''}`}
                role="switch"
                aria-checked={isDark}
                aria-labelledby={`${menuId}-theme`}
                onClick={() => toggleThemeUi()}
              >
                <span className="app-settings-menu__switch-thumb" />
              </button>
            </div>
          </div>,
          document.body,
        )}
    </div>
  )
}
