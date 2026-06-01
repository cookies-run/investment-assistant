<script setup lang="ts">
import { computed } from 'vue'
import type { StockDetail } from '../api/stock'

const props = defineProps<{
  detail: StockDetail
}>()

// 3.1 订单簿不平衡度 OBI
const obi = computed(() => {
  const buy = props.detail.buy_levels?.reduce((s, l) => s + l.volume, 0) ?? 0
  const sell = props.detail.sell_levels?.reduce((s, l) => s + l.volume, 0) ?? 0
  const total = buy + sell
  return total === 0 ? 0 : (buy - sell) / total
})

// 3.2 主动攻击度 AR
const ar = computed(() => {
  const buy = props.detail.active_buy_vol ?? 0
  const sell = props.detail.active_sell_vol ?? 0
  if (sell === 0) return buy > 0 ? 999 : 1
  return buy / sell
})

// 3.3 筹码锁定匹配度 MI
const mi = computed(() => {
  const vola = Number(props.detail.minute_price_volatility ?? 0)
  const rate = Number(props.detail.minute_turnover_rate ?? 0)
  if (rate === 0) return 0
  return vola / rate
})

const obiPercent = computed(() => {
  // OBI in [-1, 1] => map to [0, 100]%
  return ((obi.value + 1) / 2) * 100
})

const obiColor = computed(() => {
  if (obi.value > 0.3) return '#ef4444'
  if (obi.value < -0.3) return '#10b981'
  return '#9ca3af'
})

const arClass = computed(() => {
  if (ar.value >= 1.5) return 'metric-up'
  if (ar.value <= 0.7) return 'metric-down'
  return 'metric-neutral'
})

const miClass = computed(() => {
  if (mi.value > 10) return 'metric-up'
  if (mi.value < 1) return 'metric-down'
  return 'metric-neutral'
})

const hasActiveData = computed(() => {
  return (props.detail.active_buy_vol ?? 0) > 0 || (props.detail.active_sell_vol ?? 0) > 0
})

// 4. 矩阵交叉验证与策略决策树
const decision = computed(() => {
  const o = obi.value
  const a = ar.value
  const m = mi.value
  const vola = props.detail.minute_price_volatility ?? 0
  const change = props.detail.change_percent ?? 0

  // 判定分支 1：虚假繁荣（筑墙诱多陷阱）
  if (o > 0.5 && a < 0.8 && vola <= 0) {
    return { truth: '虚假繁荣，主力筑墙诱多，实际在撤退', action: '坚决不追 / 减仓', color: 'red' }
  }

  // 判定分支 2：故意压盘（边洗边吸蓄势）
  // 注：PRD 原需 large_order_net_inflow > 0（L2 逐笔），当前用高 AR 降级替代
  if (o < -0.5 && a > 1.2) {
    return { truth: '故意压盘，主力边洗边吸，突破意愿强', action: '分批跟进', color: 'green' }
  }

  // 判定分支 3：缩量轻装（筹码高度锁定）
  const totalLevels = (props.detail.buy_levels?.reduce((s, l) => s + l.volume, 0) ?? 0)
                    + (props.detail.sell_levels?.reduce((s, l) => s + l.volume, 0) ?? 0)
  if (totalLevels < 100000 && a > 1.5 && m > 10) {
    return { truth: '筹码高度锁定，空头无力反抗，轻装上阵', action: '持股待涨 / 右侧轻仓试水', color: 'blue' }
  }

  // 判定分支 4：主力真红军（合力合拍进攻）
  // 注：PRD 原需 large_order_net_inflow > 0（L2 逐笔），当前用高 AR 降级替代
  if (o > 0.5 && a > 1.2 && change > 0) {
    return { truth: '真正的主力护盘兼合力进攻', action: '安心持有 / 回踩买入', color: 'purple' }
  }

  return null
})
</script>

<template>
  <div class="psychology-card">
    <div class="section-title">盘口博弈验证</div>

    <!-- OBI 仪表盘 -->
    <div class="obi-section">
      <div class="obi-label">订单簿不平衡度 (OBI)</div>
      <div class="gauge-wrapper">
        <div class="gauge-track">
          <div class="gauge-fill" :style="{ width: obiPercent + '%', background: obiColor }" />
        </div>
        <div class="gauge-pointer" :style="{ left: obiPercent + '%' }" />
      </div>
      <div class="gauge-labels">
        <span class="label-sell">卖压大</span>
        <span class="label-balance">平衡</span>
        <span class="label-buy">买盘大</span>
      </div>
      <div class="obi-value" :style="{ color: obiColor }">
        {{ obi > 0 ? '+' : '' }}{{ obi.toFixed(2) }}
      </div>
    </div>

    <!-- 指标网格 -->
    <div class="metrics-grid">
      <div class="metric-item">
        <div class="metric-label">主动攻击度 (AR)</div>
        <div class="metric-value" :class="arClass">{{ ar.toFixed(2) }}</div>
        <div class="metric-desc">
          外盘/内盘 = {{ (detail.active_buy_vol ?? 0).toLocaleString() }} / {{ (detail.active_sell_vol ?? 0).toLocaleString() }}
        </div>
        <div v-if="!hasActiveData" class="metric-warn">外盘/内盘数据暂缺，需 L2 行情源</div>
      </div>
      <div class="metric-item">
        <div class="metric-label">筹码锁定度 (MI)</div>
        <div class="metric-value" :class="miClass">{{ mi.toFixed(2) }}</div>
        <div class="metric-desc">
          震幅/换手 = {{ Number(detail.minute_price_volatility ?? 0).toFixed(4) }}% / {{ Number(detail.minute_turnover_rate ?? 0).toFixed(4) }}%
        </div>
      </div>
    </div>

    <!-- 决策结果 -->
    <div v-if="decision" class="decision-result" :class="`decision-${decision.color}`">
      <div class="decision-tag">博弈真相</div>
      <div class="decision-truth">{{ decision.truth }}</div>
      <div class="decision-tag">交易动作</div>
      <div class="decision-action">{{ decision.action }}</div>
    </div>

    <div v-else class="decision-neutral">
      当前盘口无显著博弈信号，保持观察
    </div>
  </div>
</template>

<style scoped>
.psychology-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.02);
  border: 1px solid rgba(0, 0, 0, 0.04);
  margin-top: 20px;
}

.section-title {
  font-size: 15px;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 16px;
}

.obi-section {
  margin-bottom: 20px;
}

.obi-label {
  font-size: 13px;
  color: #6b7280;
  font-weight: 500;
  margin-bottom: 10px;
}

.gauge-wrapper {
  position: relative;
  height: 14px;
  margin-bottom: 6px;
}

.gauge-track {
  position: absolute;
  top: 4px;
  left: 0;
  right: 0;
  height: 6px;
  background: linear-gradient(90deg, #86efac 0%, #e5e7eb 50%, #fca5a5 100%);
  border-radius: 3px;
}

.gauge-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.5s ease;
}

.gauge-pointer {
  position: absolute;
  top: 0;
  width: 4px;
  height: 14px;
  background: #1f2937;
  border-radius: 2px;
  transform: translateX(-50%);
  transition: left 0.5s ease;
}

.gauge-labels {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #9ca3af;
  margin-bottom: 6px;
}

.label-sell { color: #10b981; }
.label-buy { color: #ef4444; }

.obi-value {
  text-align: center;
  font-size: 20px;
  font-weight: 800;
}

.metrics-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 16px;
}

.metric-item {
  background: #f9fafb;
  border-radius: 12px;
  padding: 14px;
  text-align: center;
}

.metric-label {
  font-size: 12px;
  color: #6b7280;
  font-weight: 500;
  margin-bottom: 6px;
}

.metric-value {
  font-size: 22px;
  font-weight: 800;
  margin-bottom: 4px;
}

.metric-up { color: #ef4444; }
.metric-down { color: #10b981; }
.metric-neutral { color: #6b7280; }

.metric-desc {
  font-size: 11px;
  color: #9ca3af;
}

.metric-warn {
  font-size: 10px;
  color: #f59e0b;
  margin-top: 4px;
}

.decision-result {
  border-radius: 12px;
  padding: 16px;
  border: 2px solid transparent;
  animation: breathe 2s ease-in-out infinite;
}

@keyframes breathe {
  0%, 100% { box-shadow: 0 0 0 0 rgba(0, 0, 0, 0.05); }
  50% { box-shadow: 0 0 12px 4px rgba(0, 0, 0, 0.08); }
}

.decision-red {
  background: #fef2f2;
  border-color: #fca5a5;
}
.decision-red .decision-truth,
.decision-red .decision-action {
  color: #b91c1c;
}

.decision-green {
  background: #f0fdf4;
  border-color: #86efac;
}
.decision-green .decision-truth,
.decision-green .decision-action {
  color: #15803d;
}

.decision-blue {
  background: #eff6ff;
  border-color: #93c5fd;
}
.decision-blue .decision-truth,
.decision-blue .decision-action {
  color: #1d4ed8;
}

.decision-purple {
  background: #faf5ff;
  border-color: #d8b4fe;
}
.decision-purple .decision-truth,
.decision-purple .decision-action {
  color: #7e22ce;
}

.decision-tag {
  font-size: 11px;
  font-weight: 600;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 4px;
}

.decision-truth {
  font-size: 15px;
  font-weight: 700;
  margin-bottom: 10px;
  line-height: 1.4;
}

.decision-action {
  font-size: 14px;
  font-weight: 600;
}

.decision-neutral {
  text-align: center;
  padding: 16px;
  color: #9ca3af;
  font-size: 14px;
  background: #f9fafb;
  border-radius: 12px;
}

@media (max-width: 768px) {
  .metrics-grid {
    grid-template-columns: 1fr;
  }
}
</style>
