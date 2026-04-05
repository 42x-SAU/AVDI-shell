import { useMemo, useState } from 'react'
import { PanelExportButtons } from '@/components/dashboard/PanelExportButtons'
import { retryTask } from '@/services/monitoringApi'
import { parseMonitoringFetchError } from '@/services/monitoringSnapshot'
import { formatDateTime } from '@/utils/formatDateTime'

const STATUS_OPTIONS = [
  { value: 'all', label: 'Все статусы' },
  { value: 'pending', label: 'pending' },
  { value: 'running', label: 'running' },
  { value: 'done', label: 'done' },
  { value: 'failed', label: 'failed' },
]

function taskStatusBadgeClass(status) {
  switch (status) {
    case 'pending':
      return 'dashboard-badge--task-pending'
    case 'running':
      return 'dashboard-badge--task-running'
    case 'done':
      return 'dashboard-badge--task-done'
    case 'failed':
      return 'dashboard-badge--task-failed'
    default:
      return ''
  }
}

/**
 * Таблица задач, фильтры и retry для done/failed.
 */
export function TasksPanel({ tasks, loading, error, onRefresh }) {
  const [statusFilter, setStatusFilter] = useState('all')
  const [agentFilter, setAgentFilter] = useState('all')
  const [retryingId, setRetryingId] = useState(null)
  const [actionError, setActionError] = useState(null)

  const agentOptions = useMemo(() => {
    const ids = new Set()
    tasks.forEach((t) => {
      if (t.agent_id != null) ids.add(Number(t.agent_id))
    })
    return Array.from(ids).sort((a, b) => a - b)
  }, [tasks])

  const filteredTasks = useMemo(() => {
    return tasks.filter((t) => {
      if (statusFilter !== 'all' && t.status !== statusFilter) return false
      if (agentFilter !== 'all' && Number(t.agent_id) !== Number(agentFilter)) return false
      return true
    })
  }, [tasks, statusFilter, agentFilter])

  const handleRetry = async (taskId) => {
    setActionError(null)
    setRetryingId(taskId)
    try {
      await retryTask(taskId)
      await onRefresh()
    } catch (e) {
      setActionError(parseMonitoringFetchError(e, 'Не удалось выполнить повтор'))
    } finally {
      setRetryingId(null)
    }
  }

  if (loading && (!tasks || tasks.length === 0)) {
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
    <div className="dashboard-tabs__panel dashboard-tabs__panel--data">
      <div className="dashboard-panel-toolbar dashboard-panel-toolbar--split dashboard-panel-toolbar--tasks">
        <div className="dashboard-filters">
          <label className="dashboard-filters__field">
            <span className="dashboard-filters__label">Статус</span>
            <select
              className="dashboard-filters__select"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
            >
              {STATUS_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </select>
          </label>
          <label className="dashboard-filters__field">
            <span className="dashboard-filters__label">Агент</span>
            <select
              className="dashboard-filters__select"
              value={agentFilter}
              onChange={(e) => setAgentFilter(e.target.value)}
            >
              <option value="all">Все агенты</option>
              {agentOptions.map((id) => (
                <option key={id} value={String(id)}>
                  {id}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className="dashboard-panel-toolbar__right">
          <PanelExportButtons variant="tasks" rows={tasks} />
          <button type="button" className="btn btn--secondary" onClick={onRefresh}>
            Обновить
          </button>
        </div>
      </div>
      {actionError ? (
        <p className="dashboard-data-msg dashboard-data-msg--error" role="alert">
          {actionError}
        </p>
      ) : null}
      <div className="dashboard-table-wrap" role="region" aria-label="Таблица задач">
        <table className="dashboard-table">
          <thead>
            <tr>
              <th scope="col">ID</th>
              <th scope="col">agent_id</th>
              <th scope="col">check_type</th>
              <th scope="col">status</th>
              <th scope="col">retry</th>
              <th scope="col">max</th>
              <th scope="col">Создана</th>
              <th scope="col">Старт</th>
              <th scope="col">Завершена</th>
              <th scope="col">Действия</th>
            </tr>
          </thead>
          <tbody>
            {filteredTasks.length === 0 ? (
              <tr>
                <td colSpan={10} className="dashboard-table__empty">
                  Нет задач по выбранным фильтрам
                </td>
              </tr>
            ) : (
              filteredTasks.map((t) => {
                const canRetry = t.status === 'done' || t.status === 'failed'
                return (
                  <tr key={t.id}>
                    <td className="dashboard-table__num">{t.id}</td>
                    <td className="dashboard-table__num">{t.agent_id}</td>
                    <td>{t.check_type ?? '—'}</td>
                    <td>
                      <span
                        className={`dashboard-badge dashboard-badge--sm ${taskStatusBadgeClass(t.status)}`}
                      >
                        {t.status ?? '—'}
                      </span>
                    </td>
                    <td className="dashboard-table__num">{t.retry_count ?? '—'}</td>
                    <td className="dashboard-table__num">{t.max_retries ?? '—'}</td>
                    <td className="dashboard-table__mono">{formatDateTime(t.created_at)}</td>
                    <td className="dashboard-table__mono">{formatDateTime(t.started_at)}</td>
                    <td className="dashboard-table__mono">{formatDateTime(t.finished_at)}</td>
                    <td>
                      {canRetry ? (
                        <button
                          type="button"
                          className="btn btn--primary btn--compact"
                          disabled={retryingId === t.id}
                          onClick={() => handleRetry(t.id)}
                        >
                          {retryingId === t.id ? '…' : 'Повтор'}
                        </button>
                      ) : (
                        <span className="dashboard-table__dash">—</span>
                      )}
                    </td>
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
