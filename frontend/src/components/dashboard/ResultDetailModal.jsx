import { useEffect, useId, useRef } from 'react'
import { createPortal } from 'react-dom'
import { formatJsonForDisplay } from '@/utils/formatJsonForDisplay'
import { formatDateTime } from '@/utils/formatDateTime'

/**
 * Модальное окно с полным содержимым результата (длинные поля не в таблице).
 */
export function ResultDetailModal({ result, onClose }) {
  const titleId = useId()
  const panelRef = useRef(null)

  useEffect(() => {
    if (!result) return undefined
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
  }, [result, onClose])

  useEffect(() => {
    if (result && panelRef.current) {
      const closeBtn = panelRef.current.querySelector('[data-result-modal-close]')
      if (closeBtn) closeBtn.focus()
    }
  }, [result])

  if (!result) return null

  const jsonFormatted = formatJsonForDisplay(result.result_json)
  const exitOk = Number(result.exit_code) === 0

  const block = (label, text) => (
    <div className="result-detail-modal__block">
      <h3 className="result-detail-modal__block-title">{label}</h3>
      <pre className="result-detail-modal__pre" tabIndex={0}>
        {text == null || text === '' ? '—' : String(text)}
      </pre>
    </div>
  )

  const modal = (
    <div className="result-detail-modal" aria-hidden={false}>
      <button
        type="button"
        className="result-detail-modal__backdrop"
        aria-label="Закрыть"
        onClick={onClose}
      />
      <div
        ref={panelRef}
        className="result-detail-modal__panel"
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
      >
        <header className="result-detail-modal__head">
          <div className="result-detail-modal__head-main">
            <h2 id={titleId} className="result-detail-modal__title">
              Результат #{result.id}
            </h2>
            <p className="result-detail-modal__meta">
              Задача {result.task_id} · {formatDateTime(result.created_at)}
            </p>
          </div>
          <div className="result-detail-modal__head-actions">
            <span
              className={`dashboard-badge dashboard-badge--sm ${exitOk ? 'dashboard-badge--exit-ok' : 'dashboard-badge--exit-err'}`}
            >
              exit {result.exit_code}
            </span>
            <button
              type="button"
              data-result-modal-close
              className="result-detail-modal__close"
              onClick={onClose}
              aria-label="Закрыть окно"
            >
              ×
            </button>
          </div>
        </header>
        <div className="result-detail-modal__body">
          {block('result_json', jsonFormatted)}
          {block('stdout', result.stdout)}
          {block('stderr', result.stderr)}
          {block('logs', result.logs)}
        </div>
      </div>
    </div>
  )

  return typeof document !== 'undefined' ? createPortal(modal, document.body) : null
}
