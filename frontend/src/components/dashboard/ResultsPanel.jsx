import { useMemo, useState } from 'react'
import { PanelExportButtons } from '@/components/dashboard/PanelExportButtons'
import { ResultDetailModal } from '@/components/dashboard/ResultDetailModal'
import { formatDateTime } from '@/utils/formatDateTime'

function exitBadgeClass(exitCode) {
  const n = Number(exitCode)
  if (Number.isNaN(n)) return 'dashboard-badge--exit-unknown'
  return n === 0 ? 'dashboard-badge--exit-ok' : 'dashboard-badge--exit-err'
}

/**
 * Список результатов (кратко в таблице) и модальное окно с полным текстом полей.
 */
export function ResultsPanel({ results, loading, error, onRefresh }) {
  const [taskFilter, setTaskFilter] = useState('all')
  const [exitFilter, setExitFilter] = useState('all')
  const [detail, setDetail] = useState(null)

  const taskOptions = useMemo(() => {
    const ids = new Set()
    results.forEach((r) => {
      if (r.task_id != null) ids.add(Number(r.task_id))
    })
    return Array.from(ids).sort((a, b) => a - b)
  }, [results])

  const filtered = useMemo(() => {
    return results.filter((r) => {
      if (taskFilter !== 'all' && Number(r.task_id) !== Number(taskFilter)) return false
      const code = Number(r.exit_code)
      if (exitFilter === 'ok' && code !== 0) return false
      if (exitFilter === 'err' && code === 0) return false
      return true
    })
  }, [results, taskFilter, exitFilter])

  if (loading && (!results || results.length === 0)) {
    return (
      <div className="dashboard-tabs__panel dashboard-tabs__panel--data">
        <p className="dashboard-data-msg">Загрузка…</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="dashboard-tabs__panel dashboard-tabs__panel--data">
        <p className="dashboard-data-msg dashboard-data-msg--error" role="alert">
          {error}
        </p>
        <button type="button" className="btn btn--secondary dashboard-retry-btn" onClick={onRefresh}>
          Повторить запрос
        </button>
      </div>
    )
  }

  return (
    <>
      <div className="dashboard-tabs__panel dashboard-tabs__panel--data">
        <div className="dashboard-panel-toolbar dashboard-panel-toolbar--split dashboard-panel-toolbar--tasks">
          <div className="dashboard-filters">
            <label className="dashboard-filters__field">
              <span className="dashboard-filters__label">Задача (task_id)</span>
              <select
                className="dashboard-filters__select"
                value={taskFilter}
                onChange={(e) => setTaskFilter(e.target.value)}
              >
                <option value="all">Все</option>
                {taskOptions.map((id) => (
                  <option key={id} value={String(id)}>
                    {id}
                  </option>
                ))}
              </select>
            </label>
            <label className="dashboard-filters__field">
              <span className="dashboard-filters__label">Код выхода</span>
              <select
                className="dashboard-filters__select"
                value={exitFilter}
                onChange={(e) => setExitFilter(e.target.value)}
              >
                <option value="all">Все</option>
                <option value="ok">0 (успех)</option>
                <option value="err">Не 0 (ошибка)</option>
              </select>
            </label>
          </div>
          <div className="dashboard-panel-toolbar__right">
            <PanelExportButtons variant="results" rows={results} />
            <button type="button" className="btn btn--secondary" onClick={onRefresh}>
              Обновить
            </button>
          </div>
        </div>

        <div className="dashboard-table-wrap" role="region" aria-label="Таблица результатов">
          <table className="dashboard-table dashboard-table--results">
            <thead>
              <tr>
                <th scope="col">ID</th>
                <th scope="col">task_id</th>
                <th scope="col">exit_code</th>
                <th scope="col">Создан</th>
                <th scope="col">Действия</th>
              </tr>
            </thead>
            <tbody>
              {filtered.length === 0 ? (
                <tr>
                  <td colSpan={5} className="dashboard-table__empty">
                    Нет записей по фильтрам
                  </td>
                </tr>
              ) : (
                filtered.map((r) => {
                  const problem = Number(r.exit_code) !== 0
                  return (
                    <tr
                      key={r.id}
                      className={problem ? 'dashboard-table__row--problem' : undefined}
                    >
                      <td className="dashboard-table__num">{r.id}</td>
                      <td className="dashboard-table__num">{r.task_id}</td>
                      <td>
                        <span
                          className={`dashboard-badge dashboard-badge--sm ${exitBadgeClass(r.exit_code)}`}
                        >
                          {r.exit_code}
                        </span>
                      </td>
                      <td className="dashboard-table__mono">{formatDateTime(r.created_at)}</td>
                      <td>
                        <button
                          type="button"
                          className="btn btn--secondary btn--compact"
                          onClick={() => setDetail(r)}
                        >
                          Подробнее
                        </button>
                      </td>
                    </tr>
                  )
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
      <ResultDetailModal result={detail} onClose={() => setDetail(null)} />
    </>
  )
}
