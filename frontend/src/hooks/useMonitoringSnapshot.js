import { useCallback, useEffect, useRef, useState } from 'react'
import { MONITORING_POLL_INTERVAL_MS } from '@/constants/monitoringPolling'
import { subscribeMonitoringRealtime } from '@/monitoring/realtime'
import {
  fetchMonitoringSnapshot,
  parseMonitoringFetchError,
} from '@/services/monitoringSnapshot'
import { usePolling } from '@/hooks/usePolling'

/**
 * Снимок мониторинга: начальная загрузка, фоновый polling, ручной refetch.
 * Тихие обновления не включают loading — таблицы и фильтры не мерцают.
 */
export function useMonitoringSnapshot() {
  const [agents, setAgents] = useState([])
  const [tasks, setTasks] = useState([])
  const [results, setResults] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  /** Ошибка последнего тихого обновления (данные не очищаем). */
  const [refreshError, setRefreshError] = useState(null)
  const [lastFetchedAt, setLastFetchedAt] = useState(null)

  const load = useCallback(async ({ silent = false } = {}) => {
    if (!silent) {
      setError(null)
      setRefreshError(null)
      setLoading(true)
    }
    try {
      const data = await fetchMonitoringSnapshot()
      setAgents(data.agents)
      setTasks(data.tasks)
      setResults(data.results)
      setLastFetchedAt(new Date())
      setRefreshError(null)
      if (!silent) {
        setError(null)
      }
    } catch (e) {
      const msg = parseMonitoringFetchError(e)
      if (!silent) {
        setError(msg)
        setAgents([])
        setTasks([])
        setResults([])
      } else {
        setRefreshError(msg)
      }
    } finally {
      if (!silent) {
        setLoading(false)
      }
    }
  }, [])

  const loadRef = useRef(load)
  loadRef.current = load

  useEffect(() => {
    void loadRef.current({ silent: false })
  }, [])

  const pollTick = useCallback(() => {
    loadRef.current({ silent: true })
  }, [])

  usePolling(pollTick, MONITORING_POLL_INTERVAL_MS, {
    enabled: !loading,
    pauseWhenHidden: true,
  })

  useEffect(() => {
    return subscribeMonitoringRealtime(() => {
      loadRef.current({ silent: true })
    })
  }, [])

  const refetch = useCallback((options) => load(options ?? { silent: true }), [load])

  const runningCount = tasks.filter((x) => x.status === 'running').length

  return {
    agents,
    tasks,
    results,
    loading,
    error,
    refreshError,
    lastFetchedAt,
    refetch,
    runningCount,
  }
}
