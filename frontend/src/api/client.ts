import axios from 'axios'

const client = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

let tauriBaseURLResolved = false

client.interceptors.request.use(async (config) => {
  if (!tauriBaseURLResolved) {
    try {
      const { invoke } = await import('@tauri-apps/api/core')
      const port: string = await invoke('get_backend_port')
      client.defaults.baseURL = `http://127.0.0.1:${port}/api`
      console.log('[client] tauri baseURL set to', client.defaults.baseURL)
    } catch (e) {
      // Not in Tauri environment, keep default /api
      console.log('[client] not in tauri, using default baseURL')
    }
    tauriBaseURLResolved = true
  }
  // Ensure each request uses the resolved baseURL
  if (!config.baseURL || config.baseURL === '/api') {
    config.baseURL = client.defaults.baseURL
  }
  return config
})

export async function initClientBaseURL() {
  // No-op: interceptor handles it automatically on first request
}

export default client
