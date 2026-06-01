<script setup lang="ts">
import { computed } from 'vue'
import type { Stock } from '../api/stock'
import { formatNumber, formatPercent, formatMoney, formatDate } from '../utils/format'

const props = defineProps<{ stock: Stock }>()
const emit = defineEmits<{
  (e: 'detail', stock: Stock): void
  (e: 'settings', stock: Stock): void
  (e: 'edit', stock: Stock): void
  (e: 'delete', code: string): void
}>()

function priceClass(n?: number | string) {
  if (n === undefined || n === null) return 'text-muted'
  const num = typeof n === 'string' ? parseFloat(n) : n
  if (num > 0) return 'text-up'
  if (num < 0) return 'text-down'
  return 'text-muted'
}

const dailyPnl = computed(() => {
  const mv = typeof props.stock.market_value === 'number' ? props.stock.market_value : parseFloat(props.stock.market_value as any)
  const cp = typeof props.stock.change_percent === 'number' ? props.stock.change_percent : parseFloat(props.stock.change_percent as any)
  if (isNaN(mv) || isNaN(cp)) return undefined
  return mv * cp / 100
})
</script>

<template>
  <div class="stock-card" @click="emit('detail', stock)">
    <div class="card-header">
      <div class="stock-left">
        <div class="stock-meta">
          <span class="code-tag">{{ stock.stock_code }}</span>
          <span class="status-dot" :class="stock.is_active ? 'active' : 'inactive'">监控中</span>
        </div>
        <div class="stock-name">{{ stock.stock_name }}</div>
      </div>
      <div class="price-section">
        <div class="price-label">当前价</div>
        <div class="price-main">
          <span class="price-value">{{ formatNumber(stock.current_price) }}</span>
          <span class="price-change" :class="priceClass(stock.change_percent)">{{ formatPercent(stock.change_percent) }}</span>
        </div>
      </div>
    </div>

    <div class="profit-grid">
      <div class="profit-item">
        <div class="profit-label">当日盈亏</div>
        <div class="profit-value" :class="priceClass(dailyPnl)">{{ formatMoney(dailyPnl) }}</div>
      </div>
      <div class="profit-item">
        <div class="profit-label">累计盈亏</div>
        <div class="profit-value" :class="priceClass(stock.total_pnl)">{{ formatMoney(stock.total_pnl) }}</div>
      </div>
    </div>

    <div class="card-footer">
      <div class="footer-actions">
        <el-button size="small" plain @click.stop="emit('settings', stock)">设置</el-button>
        <el-button size="small" plain @click.stop="emit('edit', stock)">编辑</el-button>
        <el-button size="small" text type="danger" @click.stop="emit('delete', stock.stock_code)">删除</el-button>
      </div>
      <div class="update-time">更新于: {{ formatDate(stock.updated_at) }}</div>
    </div>
  </div>
</template>

<style scoped>
.stock-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.02);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid rgba(0, 0, 0, 0.04);
  position: relative;
  overflow: hidden;
  cursor: pointer;
}

.stock-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.stock-left {
  flex: 1;
  min-width: 0;
}

.stock-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.stock-name {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.code-tag {
  display: inline-block;
  background: #eff6ff;
  color: #2563eb;
  font-size: 12px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 6px;
  font-family: 'SF Mono', monospace;
}

.status-dot {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #6b7280;
  font-weight: 500;
}

.status-dot::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #d1d5db;
}

.status-dot.active::before {
  background: #22c55e;
}

.price-section {
  text-align: right;
}

.price-label {
  font-size: 11px;
  color: #9ca3af;
  font-weight: 500;
  margin-bottom: 2px;
}

.price-main {
  display: flex;
  align-items: baseline;
  gap: 8px;
  justify-content: flex-end;
}

.price-value {
  font-size: 24px;
  font-weight: 800;
  color: #111827;
  line-height: 1;
}

.price-change {
  font-size: 13px;
  font-weight: 700;
}

.profit-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}

.profit-label {
  font-size: 12px;
  color: #9ca3af;
  font-weight: 500;
  margin-bottom: 4px;
}

.profit-value {
  font-size: 20px;
  font-weight: 700;
  line-height: 1.2;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 14px;
  border-top: 1px solid #f3f4f6;
}

.footer-actions {
  display: flex;
  gap: 8px;
}

.update-time {
  font-size: 12px;
  color: #9ca3af;
}

.text-up {
  color: #dc2626;
}

.text-down {
  color: #16a34a;
}

.text-muted {
  color: #9ca3af;
}
</style>
