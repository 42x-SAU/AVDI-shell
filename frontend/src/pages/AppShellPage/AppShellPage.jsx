import { useMemo } from 'react'
import { DashboardSummary, DashboardTabs, DashboardTopBar } from '@/components/dashboard'
import { useMonitoringSnapshot } from '@/hooks/useMonitoringSnapshot'
import { getMonitoringSummarySignals } from '@/utils/monitoringSummarySignals'
import '@/components/dashboard/dashboard.css'
import '@/pages/AppShellPage/AppShellPage.css'

const hasApiBaseUrl = Boolean(import.meta.env.VITE_API_BASE_URL)

/**
 * Рабочая страница /app: агенты, задачи и результаты с сервера диагностики.
 */
export function AppShellPage() {
  const {
    agents,
    tasks,
    results,
    loading,
    error,
    refreshError,
    lastFetchedAt,
    refetch,
    runningCount,
  } = useMonitoringSnapshot()

  const summarySignals = useMemo(
    () => getMonitoringSummarySignals({ agents, tasks, results }),
    [agents, tasks, results],
  )

  return (
    <div className="app-shell-page" aria-label="Рабочая область приложения">
      <DashboardTopBar
        hasBaseUrl={hasApiBaseUrl}
        loading={loading}
        error={error}
        refreshError={refreshError}
        lastFetchedAt={lastFetchedAt}
      />
      <DashboardSummary
        agentCount={agents.length}
        taskCount={tasks.length}
        runningCount={runningCount}
        resultCount={results.length}
        loading={loading}
        signals={summarySignals}
      />
      <DashboardTabs
        agents={agents}
        tasks={tasks}
        results={results}
        loading={loading}
        error={error}
        onRefresh={refetch}
      />
    </div>
  )
}
