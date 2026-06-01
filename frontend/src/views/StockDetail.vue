<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getStockDetail, updateStock, type StockDetail as StockDetailType } from '../api/stock'
import { formatNumber, formatPercent } from '../utils/format'

const route = useRoute()
const router = useRouter()
const code = route.params.code as string

const detail = ref<StockDetailType | null>(null)
const loading = ref(false)
const saving = ref(false)

const settingsForm = ref({
  daily_profit_line: 3,
  daily_loss_line: 3,
  cumulative_profit_line: 8,
  cumulative_loss_line: 8,
  cumulative_days: 5,
  monitor_interval: 60,
})

const priceClass = computed(() => {
  const pct = detail.value?.change_percent
  if (pct === undefined || pct === null) return 'text-muted'
  return pct >= 0 ? 'text-success' : 'text-danger'
})

const orderSummary = computed(() => {
  const buy = detail.value?.buy_levels || []
  const sell = detail.value?.sell_levels || []
  const totalBuy = buy.reduce((s, l) => s + l.volume, 0)
  const totalSell = sell.reduce((s, l) => s + l.volume, 0)
  const total = totalBuy + totalSell || 1
  return {
    totalBuy,
    totalSell,
    diff: totalBuy - totalSell,
    buyRatio: (totalBuy / total) * 100,
    sellRatio: (totalSell / total) * 100,
    buyCumulatives: buy.reduce<number[]>((acc, l) => {
      acc.push((acc[acc.length - 1] || 0) + l.volume)
      return acc
    }, []),
    sellCumulatives: sell.reduce<number[]>((acc, l) => {
      acc.push((acc[acc.length - 1] || 0) + l.volume)
      return acc
    }, [])
  }
})

async function fetchDetail() {
  loading.value = true
  try {
    detail.value = await getStockDetail(code)
    if (detail.value) {
      settingsForm.value = {
        daily_profit_line: detail.value.daily_profit_line ?? 3,
        daily_loss_line: detail.value.daily_loss_line ?? 3,
        cumulative_profit_line: detail.value.cumulative_profit_line ?? 8,
        cumulative_loss_line: detail.value.cumulative_loss_line ?? 8,
        cumulative_days: detail.value.cumulative_days ?? 5,
        monitor_interval: detail.value.monitor_interval ?? 60,
      }
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '获取详情失败')
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  saving.value = true
  try {
    await updateStock(code, settingsForm.value)
    ElMessage.success('保存成功')
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="detail-header">
      <el-button circle size="small" @click="router.back()">
        <el-icon><ArrowLeft /></el-icon>
      </el-button>
      <div class="detail-title" v-if="detail">
        <span class="stock-name">{{ detail.stock_name }}</span>
        <span class="stock-code">{{ detail.stock_code }}</span>
      </div>
      <div class="detail-price" v-if="detail" :class="priceClass">
        <span class="current-price">{{ formatNumber(detail.current_price) }}</span>
        <span class="change-percent">{{ formatPercent(detail.change_percent) }}</span>
      </div>
    </div>

    <div v-if="detail" class="detail-body">
      <!-- Observation Settings -->
      <div class="settings-card">
        <div class="chart-title">观察指标设置</div>
        <div class="settings-form">
          <div class="settings-row">
            <div class="settings-item">
              <span class="settings-label">单日止盈线%</span>
              <el-input-number v-model="settingsForm.daily_profit_line" :precision="2" :min="0" style="width: 100%;" />
            </div>
            <div class="settings-item">
              <span class="settings-label">单日止损线%</span>
              <el-input-number v-model="settingsForm.daily_loss_line" :precision="2" :min="0" style="width: 100%;" />
            </div>
          </div>
          <div class="settings-row">
            <div class="settings-item">
              <span class="settings-label">累计止盈线%</span>
              <el-input-number v-model="settingsForm.cumulative_profit_line" :precision="2" :min="0" style="width: 100%;" />
            </div>
            <div class="settings-item">
              <span class="settings-label">累计止损线%</span>
              <el-input-number v-model="settingsForm.cumulative_loss_line" :precision="2" :min="0" style="width: 100%;" />
            </div>
          </div>
          <div class="settings-row">
            <div class="settings-item">
              <span class="settings-label">累计天数</span>
              <el-input-number v-model="settingsForm.cumulative_days" :min="1" style="width: 100%;" />
            </div>
            <div class="settings-item">
              <span class="settings-label">监控频率(秒)</span>
              <el-input-number v-model="settingsForm.monitor_interval" :min="10" style="width: 100%;" />
            </div>
          </div>
          <div class="settings-actions">
            <el-button type="primary" :loading="saving" @click="saveSettings">保存设置</el-button>
          </div>
        </div>
      </div>

      <!-- Order Book -->
      <div class="order-book-card">
        <div class="chart-title">买卖五档</div>

        <div class="order-summary" v-if="detail">
          <div class="summary-row">
            <div class="summary-item">
              <span class="summary-label">总买盘</span>
              <span class="summary-value text-danger">{{ orderSummary.totalBuy.toLocaleString() }}</span>
            </div>
            <div class="summary-bar">
              <div class="summary-track">
                <div class="summary-fill buy-fill" :style="{ width: orderSummary.buyRatio + '%' }" />
              </div>
            </div>
            <div class="summary-item">
              <span class="summary-label">总卖盘</span>
              <span class="summary-value text-success">{{ orderSummary.totalSell.toLocaleString() }}</span>
            </div>
          </div>
          <div class="summary-diff" :class="orderSummary.diff >= 0 ? 'text-danger' : 'text-success'">
            买卖差：{{ orderSummary.diff > 0 ? '+' : '' }}{{ orderSummary.diff.toLocaleString() }}
          </div>
        </div>

        <div class="order-book">
          <div class="order-book-side">
            <div class="order-book-header">
              <span>卖盘</span>
              <span class="header-cum">累计</span>
            </div>
            <div class="order-level" v-for="(level, idx) in detail.sell_levels?.slice().reverse()" :key="'sell-' + idx">
              <span class="level-label">卖{{ 5 - idx }}</span>
              <span class="level-price text-success">{{ formatNumber(level.price) }}</span>
              <span class="level-volume">{{ level.volume.toLocaleString() }}</span>
              <span class="level-cumulative">{{ orderSummary.sellCumulatives[4 - idx]?.toLocaleString() }}</span>
            </div>
          </div>
          <div class="order-book-side">
            <div class="order-book-header">
              <span>买盘</span>
              <span class="header-cum">累计</span>
            </div>
            <div class="order-level" v-for="(level, idx) in detail.buy_levels" :key="'buy-' + idx">
              <span class="level-label">买{{ idx + 1 }}</span>
              <span class="level-price text-danger">{{ formatNumber(level.price) }}</span>
              <span class="level-volume">{{ level.volume.toLocaleString() }}</span>
              <span class="level-cumulative">{{ orderSummary.buyCumulatives[idx]?.toLocaleString() }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.detail-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
  padding: 16px;
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.02);
  border: 1px solid rgba(0, 0, 0, 0.04);
}

.detail-title {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stock-name {
  font-size: 18px;
  font-weight: 700;
  color: #111827;
}

.stock-code {
  font-size: 13px;
  color: #9ca3af;
}

.detail-price {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.current-price {
  font-size: 22px;
  font-weight: 800;
}

.change-percent {
  font-size: 13px;
  font-weight: 600;
}

.detail-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.order-book-card,
.settings-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.02);
  border: 1px solid rgba(0, 0, 0, 0.04);
}

.settings-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.settings-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.settings-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.settings-label {
  font-size: 13px;
  color: #6b7280;
  font-weight: 500;
}

.settings-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 8px;
}

@media (max-width: 768px) {
  .settings-row {
    grid-template-columns: 1fr;
  }
}

.chart-title {
  font-size: 15px;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 16px;
}

.order-book {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.order-book-side {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.order-book-header {
  font-size: 13px;
  font-weight: 600;
  color: #6b7280;
  padding-bottom: 8px;
  border-bottom: 1px solid #f3f4f6;
  display: flex;
  justify-content: space-between;
}

.order-level {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: #f9fafb;
  border-radius: 8px;
}

.level-label {
  font-size: 12px;
  color: #9ca3af;
  font-weight: 500;
  width: 36px;
}

.level-price {
  font-size: 14px;
  font-weight: 700;
  flex: 1;
  text-align: center;
}

.level-volume {
  font-size: 13px;
  color: #374151;
  font-weight: 600;
  width: 80px;
  text-align: right;
}

.level-cumulative {
  font-size: 12px;
  color: #6b7280;
  font-weight: 500;
  width: 80px;
  text-align: right;
}

.order-summary {
  margin-bottom: 20px;
  padding: 16px;
  background: #f9fafb;
  border-radius: 12px;
}

.summary-row {
  display: flex;
  align-items: center;
  gap: 16px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  min-width: 80px;
}

.summary-label {
  font-size: 12px;
  color: #9ca3af;
  font-weight: 500;
}

.summary-value {
  font-size: 16px;
  font-weight: 700;
}

.summary-bar {
  flex: 1;
  display: flex;
  align-items: center;
}

.summary-track {
  width: 100%;
  height: 10px;
  background: #e5e7eb;
  border-radius: 5px;
  overflow: hidden;
  display: flex;
}

.buy-fill {
  height: 100%;
  background: linear-gradient(90deg, #fca5a5, #ef4444);
  border-radius: 5px;
  transition: width 0.5s ease;
}

.summary-diff {
  text-align: center;
  font-size: 14px;
  font-weight: 700;
  margin-top: 10px;
}

.header-cum {
  font-size: 11px;
  color: #9ca3af;
  font-weight: 500;
}

@media (max-width: 768px) {
  .order-book {
    grid-template-columns: 1fr;
  }
}
</style>
