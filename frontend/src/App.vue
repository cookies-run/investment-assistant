<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import NotificationSettings from './components/NotificationSettings.vue'
import { initClientBaseURL } from './api/client'

const route = useRoute()
const router = useRouter()
const settingsRef = ref<InstanceType<typeof NotificationSettings> | null>(null)

onMounted(() => {
  initClientBaseURL()
})

const tabs = [
  { name: 'market', label: '全球指数', icon: 'TrendCharts' },
  { name: 'stocks', label: '股票持仓', icon: 'Collection' },
  { name: 'funds', label: '基金持仓', icon: 'Box' },
  { name: 'alerts', label: '预警记录', icon: 'Bell' },
]

function goTab(path: string) {
  router.push(`/${path}`)
}

function openSettings() {
  settingsRef.value?.open()
}
</script>

<template>
  <el-container style="min-height: 100vh;">
    <el-header class="app-header">
      <div class="header-left">
        <div class="logo-box">
          <el-icon :size="22" color="#2563eb"><TrendCharts /></el-icon>
        </div>
        <span class="app-title">投资助手</span>
      </div>
      <el-menu
        :default-active="route.path.slice(1) || 'market'"
        mode="horizontal"
        class="header-menu"
        @select="goTab"
      >
        <el-menu-item v-for="tab in tabs" :key="tab.name" :index="tab.name">
          <el-icon><component :is="tab.icon" /></el-icon>
          <span>{{ tab.label }}</span>
        </el-menu-item>
      </el-menu>
      <div class="header-actions">
        <el-button circle text @click="openSettings" title="通知设置">
          <el-icon :size="18"><Setting /></el-icon>
        </el-button>
      </div>
    </el-header>

    <NotificationSettings ref="settingsRef" />

    <el-main class="app-main">
      <router-view />
    </el-main>
  </el-container>
</template>

<style scoped>
.app-header {
  background: #ffffff;
  border-bottom: 1px solid #e5e7eb;
  padding: 0 24px;
  display: flex;
  align-items: center;
  height: 64px !important;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-right: 32px;
}

.logo-box {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.app-title {
  font-size: 20px;
  font-weight: 800;
  color: #111827;
  letter-spacing: -0.5px;
}

.header-menu {
  flex: 1;
  border-bottom: none;
  background: transparent;
}

.header-menu :deep(.el-menu-item) {
  font-size: 15px;
  font-weight: 500;
  color: #6b7280;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.header-menu :deep(.el-menu-item.is-active) {
  color: #2563eb;
  font-weight: 600;
  border-bottom: 2px solid #2563eb;
}

.header-menu :deep(.el-menu-item:hover) {
  color: #2563eb;
  background: #eff6ff;
  border-radius: 6px 6px 0 0;
}

.header-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 8px;
}

.app-main {
  padding: 0;
  background: #f0f2f5;
}
</style>
