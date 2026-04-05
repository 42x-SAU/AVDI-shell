/**
 * Верхняя панель рабочей зоны и индикатор связи с API.
 * refreshError — только для тихого обновления (polling); основные данные не сбрасываются.
 */
export function DashboardTopBar({ hasBaseUrl, loading, error, refreshError, lastFetchedAt }) {
  const badge =
    !hasBaseUrl
      ? { text: 'Задайте VITE_API_BASE_URL', mod: 'dashboard-badge--warn' }
      : error
        ? { text: 'Ошибка загрузки', mod: 'dashboard-badge--danger' }
        : loading
          ? { text: 'Загрузка…', mod: '' }
          : { text: 'Данные с API', mod: 'dashboard-badge--ok' }

  const syncTime =
    lastFetchedAt instanceof Date && !Number.isNaN(lastFetchedAt.getTime())
      ? new Intl.DateTimeFormat('ru-RU', {
          timeStyle: 'medium',
          dateStyle: 'short',
        }).format(lastFetchedAt)
      : null

  return (
    <header className="dashboard-top-bar">
      <div className="container dashboard-top-bar__inner">
        <div>
          <h1 className="dashboard-top-bar__title">Панель мониторинга</h1>
          <p className="dashboard-top-bar__lead">
            Агенты и задачи загружаются с сервера диагностики. Статус агента считается по времени
            последнего heartbeat. Данные периодически обновляются в фоне.
          </p>
        </div>
        <div className="dashboard-top-bar__meta" aria-live="polite">
          <span className={['dashboard-badge', badge.mod].filter(Boolean).join(' ')}>{badge.text}</span>
          {syncTime ? (
            <span className="dashboard-top-bar__sync" title="Время последнего успешного обновления">
              Обновлено: {syncTime}
            </span>
          ) : null}
          {refreshError && !error ? (
            <span className="dashboard-badge dashboard-badge--warn" title={refreshError}>
              Фоновое обновление не удалось
            </span>
          ) : null}
        </div>
      </div>
    </header>
  )
}
