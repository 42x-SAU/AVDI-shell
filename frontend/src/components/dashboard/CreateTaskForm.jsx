import { useState } from 'react'
import { useCreateTask } from '@/hooks/useCreateTask'

/**
 * Форма создания задачи (POST /tasks). Значения по умолчанию задаёт родитель (шаблон или ручной режим).
 */
export function CreateTaskForm({ agents, defaultValues, onCreated }) {
  const { submit, submitting, error, clearError } = useCreateTask(onCreated)

  const [agentId, setAgentId] = useState(() =>
    defaultValues?.agent_id != null ? String(defaultValues.agent_id) : '',
  )
  const [checkType, setCheckType] = useState(() => defaultValues?.check_type ?? 'hostname')
  const [payload, setPayload] = useState(() => defaultValues?.payload ?? '')
  const [maxRetries, setMaxRetries] = useState(() =>
    defaultValues?.max_retries != null ? String(defaultValues.max_retries) : '0',
  )

  const handleSubmit = (e) => {
    e.preventDefault()
    clearError()
    const resolvedAgent = agentId || (agents[0] ? String(agents[0].id) : '')
    if (!resolvedAgent) return
    const mr = Math.max(0, parseInt(maxRetries, 10))
    if (Number.isNaN(mr)) return
    submit({
      agent_id: Number(resolvedAgent),
      check_type: checkType.trim(),
      payload,
      max_retries: mr,
    })
  }

  const agentsEmpty = agents.length === 0

  return (
    <form className="create-task-form" onSubmit={handleSubmit} noValidate>
      <h3 className="create-task-form__title">Создание задачи</h3>
      {agentsEmpty ? (
        <p className="dashboard-data-msg" role="status">
          Нет агентов — сначала зарегистрируйте агента на сервере, затем обновите панель.
        </p>
      ) : null}

      <div className="create-task-form__grid">
        <label className="create-task-form__field">
          <span className="create-task-form__label">Агент</span>
          <select
            className="create-task-form__input"
            value={agentId || (agents[0] ? String(agents[0].id) : '')}
            onChange={(e) => setAgentId(e.target.value)}
            required
            disabled={agentsEmpty || submitting}
          >
            <option value="" disabled={agents.length > 0}>
              Выберите агента
            </option>
            {agents.map((a) => (
              <option key={a.id} value={String(a.id)}>
                {a.id} — {a.name || 'без имени'}
              </option>
            ))}
          </select>
        </label>

        <label className="create-task-form__field">
          <span className="create-task-form__label">check_type</span>
          <input
            className="create-task-form__input"
            type="text"
            value={checkType}
            onChange={(e) => setCheckType(e.target.value)}
            placeholder="hostname, ping, diagnostic…"
            required
            disabled={submitting}
            autoComplete="off"
          />
        </label>

        <label className="create-task-form__field create-task-form__field--full">
          <span className="create-task-form__label">payload (строка для агента)</span>
          <textarea
            className="create-task-form__textarea"
            value={payload}
            onChange={(e) => setPayload(e.target.value)}
            rows={5}
            disabled={submitting}
            spellCheck={false}
          />
        </label>

        <label className="create-task-form__field">
          <span className="create-task-form__label">max_retries</span>
          <input
            className="create-task-form__input"
            type="number"
            min={0}
            step={1}
            value={maxRetries}
            onChange={(e) => setMaxRetries(e.target.value)}
            disabled={submitting}
          />
        </label>
      </div>

      {error ? (
        <p className="dashboard-data-msg dashboard-data-msg--error" role="alert">
          {error}
        </p>
      ) : null}

      <div className="create-task-form__actions">
        <button type="submit" className="btn btn--primary" disabled={agentsEmpty || submitting}>
          {submitting ? 'Отправка…' : 'Создать задачу'}
        </button>
      </div>
    </form>
  )
}
