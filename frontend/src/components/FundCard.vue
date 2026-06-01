<script setup lang="ts">
import { computed } from 'vue'
import type { Fund } from '../api/fund'
import { formatNumber, formatPercent, formatMoney, formatDate } from '../utils/format'

const props = defineProps<{ fund: Fund }>()
const emit = defineEmits<{
  (e: 'edit', fund: Fund): void
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
  const mv = typeof props.fund.market_value === 'number' ? props.fund.market_value : parseFloat(props.fund.market_value as any)
  const cp = typeof props.fund.change_percent === 'number' ? props.fund.change_percent : parseFloat(props.fund.change_percent as any)
  if (isNaN(mv) || isNaN(cp)) return undefined
  return mv * cp / 100
})
</script>

<template>
  <div class="fund-card">
    <div class="card-header">
      <div class="fund-left">
        <div class="fund-meta">
          <span class="code-tag">{{ fund.fund_code }}</span>
          <span class="status-dot" :class="fund.is_active ? 'active' : 'inactive'">监控中</span>
        </div>
        <div class="fund-name">{{ fund.fund_name }}</div>
      </div>
      <div class="nav-section">
        <div class="nav-label">单位净值</div>
        <div class="nav-main">
          <span class="nav-value">{{ formatNumber(fund.current_nav) }}</span>
          <span class="nav-change" :class="priceClass(fund.change_percent)">{{ formatPercent(fund.change_percent) }}</span>
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
        <div class="profit-value" :class="priceClass(fund.total_pnl)">{{ formatMoney(fund.total_pnl) }}</div>
      </div>
    </div>

    <div class="card-footer">
      <div class="footer-actions">
        <el-button size="small" plain @click.stop="emit('edit', fund)">编辑</el-button>
        <el-button size="small" text type="danger" @click.stop="emit('delete', fund.fund_code)">删除</el-button>
      </div>
      <div class="update-time">更新于: {{ formatDate(fund.updated_at) }}</div>
    </div>
  </div>
</template>

<style scoped>
.fund-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.02);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid rgba(0, 0, 0, 0.04);
  position: relative;
  overflow: hidden;
}

.fund-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.fund-left {
  flex: 1;
  min-width: 0;
}

.fund-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.fund-name {
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
  background: #eef2ff;
  color: #4f46e5;
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

.nav-section {
  text-align: right;
}

.nav-label {
  font-size: 11px;
  color: #9ca3af;
  font-weight: 500;
  margin-bottom: 2px;
}

.nav-main {
  display: flex;
  align-items: baseline;
  gap: 8px;
  justify-content: flex-end;
}

.nav-value {
  font-size: 24px;
  font-weight: 800;
  color: #111827;
  line-height: 1;
}

.nav-change {
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
