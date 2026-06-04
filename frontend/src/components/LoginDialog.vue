<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'

const visible = ref(false)
const email = ref('')
const code = ref('')
const loading = ref(false)
const sending = ref(false)
const countdown = ref(0)
const error = ref('')

const auth = useAuthStore()

function open() {
  visible.value = true
  email.value = ''
  code.value = ''
  error.value = ''
  countdown.value = 0
}

async function sendCode() {
  if (!email.value.trim()) {
    error.value = '请输入邮箱'
    return
  }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value.trim())) {
    error.value = '请输入有效的邮箱地址'
    return
  }
  sending.value = true
  error.value = ''
  try {
    await auth.sendCode(email.value.trim())
    countdown.value = 60
    const timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (e: any) {
    error.value = e.response?.data?.error || '发送验证码失败'
  } finally {
    sending.value = false
  }
}

async function submit() {
  if (!email.value.trim() || !code.value.trim()) {
    error.value = '请输入邮箱和验证码'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await auth.loginByEmail(email.value.trim(), code.value.trim())
    visible.value = false
  } catch (e: any) {
    error.value = e.response?.data?.error || '登录失败'
  } finally {
    loading.value = false
  }
}

defineExpose({ open })
</script>

<template>
  <el-dialog v-model="visible" title="欢迎使用投资助手" width="400px" :close-on-click-modal="false" :show-close="false">
    <div class="login-body">
      <p class="login-desc">请输入邮箱地址并获取验证码登录</p>
      <el-input v-model="email" placeholder="请输入邮箱" style="margin-bottom: 12px;" @keyup.enter="sendCode">
        <template #append>
          <el-button @click="sendCode" :loading="sending" :disabled="countdown > 0">
            {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
          </el-button>
        </template>
      </el-input>
      <el-input v-model="code" placeholder="请输入验证码" maxlength="6" @keyup.enter="submit" />
      <p v-if="error" class="login-error">{{ error }}</p>
    </div>
    <template #footer>
      <el-button type="primary" @click="submit" :loading="loading" :disabled="!email.trim() || !code.trim()">
        登录 / 注册
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.login-body {
  padding: 10px 0;
}

.login-desc {
  color: #6b7280;
  font-size: 14px;
  margin-bottom: 16px;
}

.login-error {
  color: #ef4444;
  font-size: 13px;
  margin-top: 8px;
}
</style>