import { AGENT_PRESENCE, getAgentPresence } from '@/utils/agentStatus'

/**
 * Эвристические сигналы для сводки dashboard (без «AI», только простые пороги).
 * @returns {{ level: 'warning' | 'negative', message: string }[]}
 */
export function getMonitoringSummarySignals({ agents, tasks, results }, nowMs = Date.now()) {
  const out = []

  const notOnlineAgents = agents.filter(
    (a) => getAgentPresence(a.last_heartbeat, nowMs) !== AGENT_PRESENCE.ONLINE,
  )
  if (notOnlineAgents.length > 0) {
    out.push({
      level: 'warning',
      message: `Агенты без свежего heartbeat: ${notOnlineAgents.length} (недоступны или слабая связь).`,
    })
  }

  const failedCount = tasks.filter((t) => t.status === 'failed').length
  const taskTotal = tasks.length
  /* «Заметная» доля failed или несколько штук */
  const failedNoticeable =
    failedCount >= 3 || (taskTotal > 0 && failedCount / taskTotal >= 0.25 && failedCount >= 2)
  if (failedNoticeable) {
    out.push({
      level: 'negative',
      message: `Много задач со статусом failed: ${failedCount} из ${taskTotal}.`,
    })
  } else if (failedCount >= 1) {
    out.push({
      level: 'warning',
      message: `Есть задачи со статусом failed: ${failedCount}.`,
    })
  }

  const nonzeroExit = results.filter((r) => Number(r.exit_code) !== 0)
  if (nonzeroExit.length > 0) {
    out.push({
      level: 'warning',
      message: `Результаты с ненулевым exit_code: ${nonzeroExit.length}.`,
    })
  }

  const withStderr = results.filter((r) => r.stderr && String(r.stderr).trim() !== '')
  if (withStderr.length > 0) {
    out.push({
      level: 'warning',
      message: `Результаты с непустым stderr: ${withStderr.length}.`,
    })
  }

  return out
}
