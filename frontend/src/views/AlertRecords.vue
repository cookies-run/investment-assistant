<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listAlerts, type AlertRecord } from '../api/alert'
import { formatPercent, formatNumber } from '../utils/format'

const alerts = ref<AlertRecord[]>([])
const loading = ref(false)

async function fetchData() {
  loading.value = true
  try {
    alerts.value = await listAlerts({ limit: 200 })
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function alertTypeLabel(type: string) {
  const map: Record<string, string> = {
    single_day_profit: '单日止盈',
    single_day_loss: '单日止损',
    cumulative_profit: '累计止盈',
    cumulative_loss: '累计止损',
  }
  return map[type] || type
}

function alertTypeTag(type: string) {
  if (type.includes('profit')) return 'success'
  if (type.includes('loss')) return 'danger'
  return 'info'
}

onMounted(fetchData)
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <div>
        <h2>预警记录</h2>
        <p class="section-subtitle">查看所有触发的止盈止损预警</p>
      </div>
      <el-button circle size="small" @click="fetchData">
        <el-icon><Refresh /></el-icon>
      </el-button>
    </div>

    <div class="table-card" v-loading="loading">
      <el-table :data="alerts" style="width: 100%;" :header-cell-style="{ background: '#f9fafb', fontWeight: 600, color: '#374151' }">
        <el-table-column prop="target_code" label="标的代码" width="120">
          <template #default="{ row }">
            <span class="code-text">{{ row.target_code }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="target_type" label="类型" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="row.target_type === 'stock' ? 'primary' : 'warning'" effect="light">
              {{ row.target_type === 'stock' ? '个股' : '基金' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="alert_type" label="预警类型" width="120">
          <template #default="{ row }">
            <el-tag :type="alertTypeTag(row.alert_type)" size="small" effect="light" class="alert-tag">
              {{ alertTypeLabel(row.alert_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="trigger_value" label="触发值" width="120">
          <template #default="{ row }">
            <span :class="alertTypeTag(row.alert_type) === 'success' ? 'text-success' : 'text-danger'">
              {{ formatPercent(row.trigger_value) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="threshold_value" label="阈值" width="120">
          <template #default="{ row }">
            <span class="threshold-text">{{ formatPercent(row.threshold_value) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="current_price" label="当前价" width="120">
          <template #default="{ row }">
            <span class="price-text">{{ row.current_price ? formatNumber(row.current_price, 3) : '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="triggered_at" label="触发时间" min-width="160">
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon :size="14" color="#9ca3af"><Clock /></el-icon>
              <span>{{ row.triggered_at }}</span>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && alerts.length === 0" description="暂无预警记录" />
    </div>
  </div>
</template>

<style scoped>
.table-card {
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.02);
  border: 1px solid rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

.table-card :deep(.el-table__header-wrapper) {
  border-radius: 16px 16px 0 0;
}

.table-card :deep(.el-table__inner-wrapper::before) {
  display: none;
}

.code-text {
  font-family: 'SF Mono', monospace;
  font-weight: 600;
  color: #374151;
}

.threshold-text {
  color: #6b7280;
  font-weight: 500;
}

.price-text {
  font-weight: 600;
  color: #111827;
}

.time-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #6b7280;
  font-size: 13px;
}

.alert-tag {
  font-weight: 600;
}
</style>
