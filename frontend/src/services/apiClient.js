import axios from 'axios'

/**
 * Базовый HTTP-клиент для будущей интеграции с backend.
 * Endpoints не задаём — их добавят при появлении контракта API.
 */
const baseURL = import.meta.env.VITE_API_BASE_URL || ''

export const apiClient = axios.create({
  baseURL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: false,
})

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    // Точка расширения: единая обработка ошибок, refresh-токены и т.д.
    return Promise.reject(error)
  },
)

export default apiClient
