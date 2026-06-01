<script setup lang="ts">
import { formatMoney, formatPercent } from '../utils/format'

defineProps<{
  title: string
  value: number | string
  suffix?: string
  isPercent?: boolean
  gradient?: string
}>()
</script>

<template>
  <div class="stat-card" :class="gradient">
    <div class="stat-icon">
      <slot name="icon">
        <el-icon :size="20"><TrendCharts /></el-icon>
      </slot>
    </div>
    <div class="stat-label">{{ title }}</div>
    <div class="stat-value">
      {{ isPercent ? formatPercent(value) : formatMoney(value) }}
      <span v-if="suffix" class="stat-suffix">{{ suffix }}</span>
    </div>
  </div>
</template>

<style scoped>
.stat-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.02);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid rgba(0, 0, 0, 0.04);
  position: relative;
  overflow: hidden;
}

.stat-card::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 120px;
  height: 120px;
  background: radial-gradient(circle, rgba(59, 130, 246, 0.06) 0%, transparent 70%);
  border-radius: 50%;
  transform: translate(30%, -30%);
}

.stat-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.08);
}

.stat-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #2563eb;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
}

.stat-label {
  font-size: 13px;
  color: #9ca3af;
  font-weight: 500;
  margin-bottom: 6px;
}

.stat-value {
  font-size: 26px;
  font-weight: 800;
  color: #111827;
  line-height: 1.2;
}

.stat-suffix {
  font-size: 14px;
  font-weight: 500;
  color: #6b7280;
  margin-left: 4px;
}

/* Gradient variants */
.stat-card.gradient-green .stat-icon {
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
  color: #10b981;
}
.stat-card.gradient-green::after {
  background: radial-gradient(circle, rgba(16, 185, 129, 0.08) 0%, transparent 70%);
}

.stat-card.gradient-red .stat-icon {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  color: #ef4444;
}
.stat-card.gradient-red::after {
  background: radial-gradient(circle, rgba(239, 68, 68, 0.08) 0%, transparent 70%);
}

.stat-card.gradient-purple .stat-icon {
  background: linear-gradient(135deg, #f5f3ff 0%, #ede9fe 100%);
  color: #8b5cf6;
}
.stat-card.gradient-purple::after {
  background: radial-gradient(circle, rgba(139, 92, 246, 0.08) 0%, transparent 70%);
}

.stat-card.gradient-orange .stat-icon {
  background: linear-gradient(135deg, #fff7ed 0%, #ffedd5 100%);
  color: #f97316;
}
.stat-card.gradient-orange::after {
  background: radial-gradient(circle, rgba(249, 115, 22, 0.08) 0%, transparent 70%);
}

.stat-card.gradient-blue .stat-icon {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #3b82f6;
}
</style>
