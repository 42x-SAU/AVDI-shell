import {
  downloadAgentsCsv,
  downloadAgentsJson,
  downloadResultsCsv,
  downloadResultsJson,
  downloadTasksCsv,
  downloadTasksJson,
} from '@/utils/exportMonitoringData'

/**
 * Кнопки экспорта текущего снимка данных панели (JSON / CSV, только frontend).
 * @param {'agents' | 'tasks' | 'results'} variant
 * @param {unknown[]} rows — массив объектов из снимка
 */
export function PanelExportButtons({ variant, rows, disabled }) {
  const isDisabled = Boolean(disabled)
  const list = Array.isArray(rows) ? rows : []

  const onJson = () => {
    if (isDisabled) return
    if (variant === 'agents') downloadAgentsJson(list)
    if (variant === 'tasks') downloadTasksJson(list)
    if (variant === 'results') downloadResultsJson(list)
  }

  const onCsv = () => {
    if (isDisabled) return
    if (variant === 'agents') downloadAgentsCsv(list)
    if (variant === 'tasks') downloadTasksCsv(list)
    if (variant === 'results') downloadResultsCsv(list)
  }

  return (
    <div className="panel-export-buttons" role="group" aria-label="Экспорт данных в файл">
      <span className="panel-export-buttons__label">Экспорт</span>
      <button
        type="button"
        className="btn btn--secondary btn--compact"
        disabled={isDisabled}
        onClick={onJson}
      >
        JSON
      </button>
      <button
        type="button"
        className="btn btn--secondary btn--compact"
        disabled={isDisabled}
        onClick={onCsv}
      >
        CSV
      </button>
    </div>
  )
}
