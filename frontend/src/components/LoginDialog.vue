<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'

const visible = ref(false)
const nickname = ref('')
const loading = ref(false)
const error = ref('')

const auth = useAuthStore()

function open() {
  visible.value = true
  nickname.value = ''
  error.value = ''
}

async function submit() {
  if (!nickname.value.trim()) {
    error.value = '请输入昵称'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await auth.register(nickname.value.trim())
    visible.value = false
  } catch (e: any) {
    error.value = e.response?.data?.error || '注册失败'
  } finally {
    loading.value = false
  }
}

defineExpose({ open })
</script>

<template>
  <el-dialog v-model="visible" title="欢迎使用投资助手" width="400px" :close-on-click-modal="false" :show-close="false">
    <div class="login-body">
      <p class="login-desc">首次使用，请输入一个昵称快速注册</p>
      <el-input v-model="nickname" placeholder="请输入昵称" maxlength="20" show-word-limit @keyup.enter="submit" />
      <p v-if="error" class="login-error">{{ error }}</p>
    </div>
    <template #footer>
      <el-button type="primary" @click="submit" :loading="loading" :disabled="!nickname.trim()">开始使用</el-button>
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
