/**
 * Сводные карточки и краткие эвристические сигналы по снимку мониторинга.
 */
export function DashboardSummary({
  agentCount,
  taskCount,
  runningCount,
  resultCount,
  loading,
  signals = [],
}) {
  const showPlaceholder = loading && agentCount === 0 && taskCount === 0 && resultCount === 0
  const fmt = (n) => (showPlaceholder ? '…' : String(n))

  return (
    <section className="dashboard-summary" aria-labelledby="dashboard-summary-heading">
      <h2 id="dashboard-summary-heading" className="sr-only">
        Сводка
      </h2>
      <div className="container">
        <div className="dashboard-summary__grid">
          <article className="dashboard-summary-card">
            <h3 className="dashboard-summary-card__label">Агенты</h3>
            <p className="dashboard-summary-card__value">{fmt(agentCount)}</p>
            <p className="dashboard-summary-card__hint">Всего зарегистрировано</p>
          </article>
          <article className="dashboard-summary-card">
            <h3 className="dashboard-summary-card__label">Задачи</h3>
            <p className="dashboard-summary-card__value">{fmt(taskCount)}</p>
            <p className="dashboard-summary-card__hint">Всего в системе</p>
          </article>
          <article className="dashboard-summary-card">
            <h3 className="dashboard-summary-card__label">Активные</h3>
            <p className="dashboard-summary-card__value">{fmt(runningCount)}</p>
            <p className="dashboard-summary-card__hint">Статус running</p>
          </article>
          <article className="dashboard-summary-card">
            <h3 className="dashboard-summary-card__label">Результаты</h3>
            <p className="dashboard-summary-card__value">{fmt(resultCount)}</p>
            <p className="dashboard-summary-card__hint">Записей result</p>
          </article>
        </div>

        {!showPlaceholder && signals.length > 0 ? (
          <ul className="dashboard-summary-signals" aria-label="Краткие заметки по данным">
            {signals.map((s, i) => (
              <li
                key={i}
                className={`dashboard-summary-signals__item dashboard-summary-signals__item--${s.level}`}
              >
                {s.message}
              </li>
            ))}
          </ul>
        ) : null}
      </div>
    </section>
  )
}
