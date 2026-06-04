import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { quickRegister, sendEmailCode, emailLogin, getMe, type User } from '../api/auth'

const TOKEN_KEY = 'auth_token'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const user = ref<User | null>(null)
  const loading = ref(false)

  const isLoggedIn = computed(() => !!token.value && !!user.value)

  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem(TOKEN_KEY, newToken)
  }

  function clearToken() {
    token.value = null
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
  }

  async function register(nickname: string) {
    loading.value = true
    try {
      const data = await quickRegister(nickname)
      setToken(data.token)
      user.value = data.user
      return data
    } finally {
      loading.value = false
    }
  }

  async function loginByEmail(email: string, code: string) {
    loading.value = true
    try {
      const data = await emailLogin(email, code)
      setToken(data.token)
      user.value = data.user
      return data
    } finally {
      loading.value = false
    }
  }

  async function sendCode(email: string) {
    await sendEmailCode(email)
  }

  async function fetchUser() {
    if (!token.value) return
    loading.value = true
    try {
      const data = await getMe()
      user.value = data
    } catch (e) {
      clearToken()
    } finally {
      loading.value = false
    }
  }

  function logout() {
    clearToken()
  }

  return { token, user, loading, isLoggedIn, setToken, clearToken, register, loginByEmail, sendCode, fetchUser, logout }
})
