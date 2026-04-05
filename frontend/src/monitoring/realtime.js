/**
 * Точка расширения для будущего real-time (WebSocket и т.п.).
 *
 * Когда на сервере появятся события, реализуйте подписку здесь и вызывайте
 * requestSilentRefresh() для тихого обновления снимка или передайте дельту в отдельный merge-слой.
 *
 * @param {() => void} requestSilentRefresh — вызвать для полного тихого refetch (как успешный poll)
 * @returns {() => void} функция отписки (очистка сокета / слушателей)
 */
export function subscribeMonitoringRealtime(requestSilentRefresh) {
  /* Заглушка: параметр зарезервирован для будущей подписки, без предупреждения линтера. */
  void requestSilentRefresh
  return () => {}
}
