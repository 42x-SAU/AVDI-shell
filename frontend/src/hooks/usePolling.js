import { useEffect, useRef } from 'react'

/**
 * Периодический вызов callback с паузой, пока вкладка скрыта (экономия запросов).
 * Не привязан к домену мониторинга — общий таймер без дублирования в компонентах.
 */
export function usePolling(callback, intervalMs, { enabled = true, pauseWhenHidden = true } = {}) {
  const cbRef = useRef(callback)

  useEffect(() => {
    cbRef.current = callback
  }, [callback])

  useEffect(() => {
    if (!enabled || intervalMs <= 0) return undefined

    let intervalId = null

    const tick = () => {
      cbRef.current?.()
    }

    const clear = () => {
      if (intervalId != null) {
        clearInterval(intervalId)
        intervalId = null
      }
    }

    const start = () => {
      clear()
      intervalId = setInterval(tick, intervalMs)
    }

    const onVisibility = () => {
      if (pauseWhenHidden && document.hidden) {
        clear()
        return
      }
      tick()
      start()
    }

    if (pauseWhenHidden && document.hidden) {
      /* ждём возврата на вкладку */
    } else {
      start()
    }

    document.addEventListener('visibilitychange', onVisibility)
    return () => {
      clear()
      document.removeEventListener('visibilitychange', onVisibility)
    }
  }, [intervalMs, enabled, pauseWhenHidden])
}
