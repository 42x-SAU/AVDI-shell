import { useEffect, useLayoutEffect, useRef } from 'react'
import { useLocation } from 'react-router-dom'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import '@/styles/nprogress-overrides.css'
/**
 * Централизованная интеграция nprogress: старт при смене маршрута и завершение после отрисовки.
 * Первый цикл пропускает start: полоску уже запустил main.jsx (холодный старт страницы).
 */
export function NavigationProgress() {
  const location = useLocation()
  const timerRef = useRef(null)
  const isFirstCycle = useRef(true)

  useEffect(() => {
    NProgress.configure({ showSpinner: false, trickleSpeed: 175, minimum: 0.12 })
  }, [])

  useLayoutEffect(() => {
    if (isFirstCycle.current) {
      isFirstCycle.current = false
      return () => {
        if (timerRef.current) clearTimeout(timerRef.current)
      }
    }
    NProgress.start()
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [location.pathname, location.search, location.hash])

  useEffect(() => {
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(() => {
      NProgress.done()
      timerRef.current = null
    }, 240)
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [location.pathname, location.search, location.hash])

  return null
}
