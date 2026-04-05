import { useId, useState } from 'react'
import { AgentsPanel } from '@/components/dashboard/AgentsPanel'
import { ResultsPanel } from '@/components/dashboard/ResultsPanel'
import { TemplatesPanel } from '@/components/dashboard/TemplatesPanel'
import { TasksPanel } from '@/components/dashboard/TasksPanel'

/** Идентификаторы вкладок */
const TAB_IDS = ['agents', 'tasks', 'results', 'templates']

const TAB_LABELS = {
  agents: 'Агенты',
  tasks: 'Задачи',
  results: 'Результаты',
  templates: 'Шаблоны проверок',
}

/**
 * Вкладки мониторинга: агенты, задачи, результаты, шаблоны и создание задач.
 */
export function DashboardTabs({ agents, tasks, results, loading, error, onRefresh }) {
  const [active, setActive] = useState('agents')
  const baseId = useId()

  return (
    <section className="dashboard-tabs" aria-label="Разделы мониторинга">
      <div className="container">
        <ul className="dashboard-tabs__list" role="tablist">
          {TAB_IDS.map((id) => {
            const isActive = active === id
            return (
              <li key={id} className="dashboard-tabs__tab" role="none">
                <button
                  type="button"
                  role="tab"
                  id={`${baseId}-tab-${id}`}
                  aria-selected={isActive}
                  aria-controls={`${baseId}-panel-${id}`}
                  tabIndex={isActive ? 0 : -1}
                  className={`dashboard-tabs__trigger${isActive ? ' dashboard-tabs__trigger--active' : ''}`}
                  onClick={() => setActive(id)}
                >
                  {TAB_LABELS[id]}
                </button>
              </li>
            )
          })}
        </ul>

        {TAB_IDS.map((id) => (
          <div
            key={id}
            id={`${baseId}-panel-${id}`}
            role="tabpanel"
            aria-labelledby={`${baseId}-tab-${id}`}
            hidden={active !== id}
          >
            {active === id ? (
              <TabPanel
                tabId={id}
                agents={agents}
                tasks={tasks}
                results={results}
                loading={loading}
                error={error}
                onRefresh={onRefresh}
              />
            ) : null}
          </div>
        ))}
      </div>
    </section>
  )
}

function TabPanel({ tabId, agents, tasks, results, loading, error, onRefresh }) {
  if (tabId === 'agents') {
    return (
      <AgentsPanel agents={agents} loading={loading} error={error} onRefresh={onRefresh} />
    )
  }
  if (tabId === 'tasks') {
    return <TasksPanel tasks={tasks} loading={loading} error={error} onRefresh={onRefresh} />
  }
  if (tabId === 'results') {
    return (
      <ResultsPanel results={results} loading={loading} error={error} onRefresh={onRefresh} />
    )
  }

  if (tabId === 'templates') {
    return (
      <TemplatesPanel agents={agents} loading={loading} error={error} onRefresh={onRefresh} />
    )
  }

  return null
}
