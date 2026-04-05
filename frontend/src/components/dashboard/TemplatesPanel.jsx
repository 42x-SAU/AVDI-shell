import { useState } from 'react'
import { CreateTaskForm } from '@/components/dashboard/CreateTaskForm'
import { TaskTemplatesList } from '@/components/dashboard/TaskTemplatesList'
import { getManualTaskDefaults, TASK_TEMPLATES } from '@/config/taskTemplates'

/**
 * Вкладка шаблонов: каталог из config + форма создания задачи (шаблон или вручную).
 */
export function TemplatesPanel({ agents, loading, error, onRefresh }) {
  const [selectedTemplateId, setSelectedTemplateId] = useState(null)
  const [formKey, setFormKey] = useState(0)

  const handleSelectFromTemplate = (template) => {
    setSelectedTemplateId(template.id)
    setFormKey((k) => k + 1)
  }

  const handleManual = () => {
    setSelectedTemplateId(null)
    setFormKey((k) => k + 1)
  }

  if (loading && agents.length === 0) {
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

  const defaultValues = buildDefaultValues(selectedTemplateId, agents)

  return (
    <div className="dashboard-tabs__panel dashboard-tabs__panel--data templates-panel">
      <TaskTemplatesList
        selectedId={selectedTemplateId}
        onSelectTemplate={handleSelectFromTemplate}
      />

      <div className="templates-panel__manual">
        <button type="button" className="btn btn--secondary" onClick={handleManual}>
          Ручное создание задачи
        </button>
        <span className="templates-panel__manual-hint">
          Сбрасывает форму к значениям по умолчанию (без шаблона).
        </span>
      </div>

      <CreateTaskForm
        key={`${formKey}-${selectedTemplateId ?? 'manual'}-${agents[0]?.id ?? 'noagent'}`}
        agents={agents}
        defaultValues={defaultValues}
        onCreated={onRefresh}
      />
    </div>
  )
}

function buildDefaultValues(selectedTemplateId, agents) {
  const baseAgent = agents[0]?.id
  if (selectedTemplateId == null) {
    return { ...getManualTaskDefaults(), agent_id: baseAgent }
  }
  const t = TASK_TEMPLATES.find((x) => x.id === selectedTemplateId)
  if (!t) {
    return { ...getManualTaskDefaults(), agent_id: baseAgent }
  }
  return {
    agent_id: baseAgent,
    check_type: t.check_type,
    payload: t.examplePayload,
    max_retries: t.maxRetries ?? 0,
  }
}
