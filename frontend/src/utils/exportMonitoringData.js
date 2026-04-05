import { csvLine } from '@/utils/csvFormat'
import { downloadTextFile, exportFilename } from '@/utils/downloadFile'

/** Обрезка длинных полей только для CSV (полный текст остаётся в JSON). */
const CSV_MAX_CELL = 800

function truncateForCsv(s) {
  if (s == null) return ''
  const t = String(s)
  return t.length > CSV_MAX_CELL ? `${t.slice(0, CSV_MAX_CELL)}…` : t
}

/**
 * Экспорт агентов (текущий снимок во frontend).
 */
export function downloadAgentsJson(agents) {
  const name = exportFilename('avdi-agents', 'json')
  downloadTextFile(name, JSON.stringify(agents, null, 2), 'application/json;charset=utf-8')
}

export function downloadAgentsCsv(agents) {
  const header = csvLine(['id', 'name', 'last_heartbeat', 'created_at'])
  const lines = [header]
  for (const a of agents) {
    lines.push(
      csvLine([a.id, a.name ?? '', a.last_heartbeat ?? '', a.created_at ?? '']),
    )
  }
  const name = exportFilename('avdi-agents', 'csv')
  downloadTextFile(name, lines.join('\n'), 'text/csv;charset=utf-8')
}

/**
 * Экспорт задач.
 */
export function downloadTasksJson(tasks) {
  const name = exportFilename('avdi-tasks', 'json')
  downloadTextFile(name, JSON.stringify(tasks, null, 2), 'application/json;charset=utf-8')
}

export function downloadTasksCsv(tasks) {
  const header = csvLine([
    'id',
    'agent_id',
    'check_type',
    'payload',
    'status',
    'retry_count',
    'max_retries',
    'created_at',
    'started_at',
    'finished_at',
  ])
  const lines = [header]
  for (const t of tasks) {
    lines.push(
      csvLine([
        t.id,
        t.agent_id,
        t.check_type ?? '',
        truncateForCsv(t.payload),
        t.status ?? '',
        t.retry_count ?? '',
        t.max_retries ?? '',
        t.created_at ?? '',
        t.started_at ?? '',
        t.finished_at ?? '',
      ]),
    )
  }
  const name = exportFilename('avdi-tasks', 'csv')
  downloadTextFile(name, lines.join('\n'), 'text/csv;charset=utf-8')
}

/**
 * Экспорт результатов.
 */
export function downloadResultsJson(results) {
  const name = exportFilename('avdi-results', 'json')
  downloadTextFile(name, JSON.stringify(results, null, 2), 'application/json;charset=utf-8')
}

export function downloadResultsCsv(results) {
  const header = csvLine([
    'id',
    'task_id',
    'exit_code',
    'created_at',
    'result_json',
    'stdout',
    'stderr',
    'logs',
  ])
  const lines = [header]
  for (const r of results) {
    lines.push(
      csvLine([
        r.id,
        r.task_id,
        r.exit_code ?? '',
        r.created_at ?? '',
        truncateForCsv(r.result_json),
        truncateForCsv(r.stdout),
        truncateForCsv(r.stderr),
        truncateForCsv(r.logs),
      ]),
    )
  }
  const name = exportFilename('avdi-results', 'csv')
  downloadTextFile(name, lines.join('\n'), 'text/csv;charset=utf-8')
}
