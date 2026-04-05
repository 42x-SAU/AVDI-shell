import { useEffect, useState } from 'react'
import {
  AGENT_PRESENCE,
  getAgentPresence,
  getAgentPresenceHint,
  getAgentPresenceLabel,
} from '@/utils/agentStatus'
import { PanelExportButtons } from '@/components/dashboard/PanelExportButtons'
import { formatDateTime } from '@/utils/formatDateTime'

/**
 * Таблица агентов: статус вычисляется по last_heartbeat на клиенте.
 */
export function AgentsPanel({ agents, loading, error, onRefresh }) {
  /** Обновление «текущего» времени вне рендера — для порогов online / stale / offline. */
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 30000)
    return () => clearInterval(id)
  }, [])
  if (loading && (!agents || agents.length === 0)) {
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
      <div className="dashboard-panel-toolbar dashboard-panel-toolbar--split">
        <PanelExportButtons variant="agents" rows={agents} />
        <button type="button" className="btn btn--secondary" onClick={onRefresh}>
          Обновить
        </button>
      </div>
      <div className="dashboard-table-wrap" role="region" aria-label="Таблица агентов">
        <table className="dashboard-table">
          <thead>
            <tr>
              <th scope="col">ID</th>
              <th scope="col">Имя</th>
              <th scope="col">Статус</th>
              <th scope="col">Последний heartbeat</th>
              <th scope="col">Создан</th>
            </tr>
          </thead>
          <tbody>
            {agents.length === 0 ? (
              <tr>
                <td colSpan={5} className="dashboard-table__empty">
                  Агентов пока нет
                </td>
              </tr>
            ) : (
              agents.map((a) => {
                const presence = getAgentPresence(a.last_heartbeat, now)
                const label = getAgentPresenceLabel(presence)
                const hint = getAgentPresenceHint(presence)
                const badgeClass =
                  presence === AGENT_PRESENCE.ONLINE
                    ? 'dashboard-badge--presence-online'
                    : presence === AGENT_PRESENCE.STALE
                      ? 'dashboard-badge--presence-stale'
                      : 'dashboard-badge--presence-offline'
                return (
                  <tr key={a.id}>
                    <td className="dashboard-table__num">{a.id}</td>
                    <td>{a.name ?? '—'}</td>
                    <td>
                      <span
                        className={`dashboard-badge dashboard-badge--sm ${badgeClass}`}
                        title={hint}
                      >
                        {label}
                      </span>
                    </td>
                    <td className="dashboard-table__mono">{formatDateTime(a.last_heartbeat)}</td>
                    <td className="dashboard-table__mono">{formatDateTime(a.created_at)}</td>
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
