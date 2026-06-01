<script setup lang="ts">
import { ref, watch, computed, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getStockDetail, type StockDetail as StockDetailType } from '../api/stock'
import { formatNumber, formatPercent } from '../utils/format'
import OrderBookPsychologyCard from './OrderBookPsychologyCard.vue'

const props = defineProps<{
  visible: boolean
  stockCode: string
  stockName: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
}>()

const detail = ref<StockDetailType | null>(null)
const loading = ref(false)
let autoRefreshTimer: number | null = null
const REFRESH_INTERVAL = 30000

function startAutoRefresh() {
  stopAutoRefresh()
  autoRefreshTimer = window.setInterval(() => {
    if (props.visible && props.stockCode) {
      fetchDetail()
    }
  }, REFRESH_INTERVAL)
}

function stopAutoRefresh() {
  if (autoRefreshTimer !== null) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
}

onUnmounted(stopAutoRefresh)

const priceClass = computed(() => {
  const pct = detail.value?.change_percent
  if (pct === undefined || pct === null) return 'text-muted'
  return pct >= 0 ? 'text-danger' : 'text-success'
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
  if (!props.stockCode) return
  loading.value = true
  try {
    detail.value = await getStockDetail(props.stockCode)
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '获取详情失败')
    detail.value = null
  } finally {
    loading.value = false
  }
}

watch(() => props.visible, (v) => {
  if (v) {
    fetchDetail()
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
})
</script>

<template>
  <el-drawer
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    :title="`${stockName} (${stockCode})`"
    size="600px"
  >
    <div v-loading="loading" class="drawer-content">
      <div class="detail-header" v-if="detail">
        <div class="price-box">
          <div class="price-label">当前价</div>
          <div class="price-value">{{ formatNumber(detail.current_price) }}</div>
        </div>
        <div class="price-box">
          <div class="price-label">涨跌幅</div>
          <div class="price-value" :class="priceClass">{{ formatPercent(detail.change_percent) }}</div>
        </div>
        <div class="price-box">
          <div class="price-label">成交量</div>
          <div class="price-value">{{ (detail.volume ?? 0).toLocaleString() }}</div>
        </div>
        <div class="price-box">
          <div class="price-label">买卖差</div>
          <div class="price-value" :class="detail.buy_sell_diff && detail.buy_sell_diff >= 0 ? 'text-danger' : 'text-success'">
            {{ detail.buy_sell_diff && detail.buy_sell_diff > 0 ? '+' : '' }}{{ (detail.buy_sell_diff ?? 0).toLocaleString() }}
          </div>
        </div>
      </div>

      <!-- 盘口博弈验证 -->
      <OrderBookPsychologyCard v-if="detail" :detail="detail" />

      <!-- Order Book -->
      <div class="order-book-card" v-if="detail">
        <div class="section-title">买卖五档</div>

        <div class="order-summary">
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
  </el-drawer>
</template>

<style scoped>
.drawer-content {
  padding: 8px 4px;
  overflow-y: auto;
  max-height: calc(100vh - 80px);
}

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
  padding: 16px;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  border-radius: 12px;
  flex-wrap: wrap;
}

.price-box {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.price-label {
  font-size: 13px;
  color: #6b7280;
}

.price-value {
  font-size: 22px;
  font-weight: 800;
}

.order-book-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.02);
  border: 1px solid rgba(0, 0, 0, 0.04);
}

.section-title {
  font-size: 15px;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 16px;
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

.text-success {
  color: #10b981;
}

.text-danger {
  color: #ef4444;
}

.text-muted {
  color: #9ca3af;
}
</style>
