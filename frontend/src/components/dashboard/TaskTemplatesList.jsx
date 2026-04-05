import { TASK_TEMPLATES } from '@/config/taskTemplates'

/**
 * Карточки шаблонов: только отображение и выбор (данные из config).
 */
export function TaskTemplatesList({ selectedId, onSelectTemplate }) {
  return (
    <div className="task-templates-list">
      <h3 className="task-templates-list__title">Готовые шаблоны</h3>
      <p className="task-templates-list__lead">
        Выберите шаблон — поля формы ниже заполнятся; при необходимости отредактируйте payload перед
        отправкой.
      </p>
      <ul className="task-templates-list__grid">
        {TASK_TEMPLATES.map((t) => {
          const active = selectedId === t.id
          return (
            <li key={t.id}>
              <article
                className={`task-templates-card${active ? ' task-templates-card--active' : ''}`}
              >
                <h4 className="task-templates-card__name">{t.name}</h4>
                <p className="task-templates-card__desc">{t.description}</p>
                <dl className="task-templates-card__meta">
                  <div>
                    <dt>check_type</dt>
                    <dd>
                      <code className="task-templates-card__code">{t.check_type}</code>
                    </dd>
                  </div>
                  <div>
                    <dt>Пример payload</dt>
                    <dd>
                      <pre className="task-templates-card__payload">{t.examplePayload || '—'}</pre>
                    </dd>
                  </div>
                </dl>
                <button
                  type="button"
                  className="btn btn--secondary btn--compact task-templates-card__btn"
                  onClick={() => onSelectTemplate(t)}
                >
                  Заполнить из шаблона
                </button>
              </article>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
