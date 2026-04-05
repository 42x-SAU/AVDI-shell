/**
 * Пороги «свежести» heartbeat относительно текущего времени (мс).
 * online — агент недавно отчитывался; stale — связь слабая; offline — давно без сигнала.
 */
export const AGENT_HEARTBEAT_ONLINE_MS = 2 * 60 * 1000
export const AGENT_HEARTBEAT_STALE_MS = 15 * 60 * 1000

/** Вычисляемое присутствие агента (в API поля status нет). */
export const AGENT_PRESENCE = {
  ONLINE: 'online',
  STALE: 'stale',
  OFFLINE: 'offline',
}

/**
 * По времени последнего heartbeat возвращает категорию присутствия.
 * @param {string | null | undefined} lastHeartbeatIso — ISO-строка с сервера
 * @param {number} [nowMs=Date.now()]
 * @returns {keyof typeof AGENT_PRESENCE}
 */
export function getAgentPresence(lastHeartbeatIso, nowMs = Date.now()) {
  if (lastHeartbeatIso == null || lastHeartbeatIso === '') {
    return AGENT_PRESENCE.OFFLINE
  }
  const t = new Date(lastHeartbeatIso).getTime()
  if (Number.isNaN(t)) {
    return AGENT_PRESENCE.OFFLINE
  }
  const age = nowMs - t
  if (age < AGENT_HEARTBEAT_ONLINE_MS) return AGENT_PRESENCE.ONLINE
  if (age < AGENT_HEARTBEAT_STALE_MS) return AGENT_PRESENCE.STALE
  return AGENT_PRESENCE.OFFLINE
}

/** Короткая подпись для UI. */
export function getAgentPresenceLabel(presence) {
  switch (presence) {
    case AGENT_PRESENCE.ONLINE:
      return 'Онлайн'
    case AGENT_PRESENCE.STALE:
      return 'Неактивен'
    case AGENT_PRESENCE.OFFLINE:
      return 'Офлайн'
    default:
      return '—'
  }
}

/** Пояснение для подсказки / aria-label. */
export function getAgentPresenceHint(presence) {
  switch (presence) {
    case AGENT_PRESENCE.ONLINE:
      return 'Heartbeat получен недавно'
    case AGENT_PRESENCE.STALE:
      return 'Heartbeat устарел (от 2 до 15 минут назад)'
    case AGENT_PRESENCE.OFFLINE:
      return 'Нет недавнего heartbeat или время неизвестно'
    default:
      return ''
  }
}
