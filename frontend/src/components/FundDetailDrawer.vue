<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getFundDetail, syncHoldings, type FundDetail } from '../api/holding'
import { updateFund } from '../api/fund'
import { createLot, deleteLot, type FundLot } from '../api/fundLot'
import { formatNumber, formatPercent, formatDateTime, formatDate } from '../utils/format'
import { useFundAdvisor, type IndexMAStatus } from '../composables/useFundAdvisor'

const props = defineProps<{
  visible: boolean
  fundCode: string
  fundName: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
}>()

const detail = ref<FundDetail | null>(null)
const loading = ref(false)
const syncing = ref(false)
const saving = ref(false)
let autoRefreshTimer: number | null = null
const REFRESH_INTERVAL = 30000

function startAutoRefresh() {
  stopAutoRefresh()
  autoRefreshTimer = window.setInterval(() => {
    if (props.visible && props.fundCode) {
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

// 设计说明：资金规模预设采用三档制（SMALL/MEDIUM/LARGE），
// 每一档联动修改单日阈值、累计止损线及均线过滤开关，
// 实现 PRD 要求的"资金联动 -> 风控前置"流水线第一步。
type PresetType = 'CUSTOM' | 'SMALL' | 'MEDIUM' | 'LARGE'

const settingsForm = ref({
  daily_profit_line: 2.5,
  daily_loss_line: 2.5,
  cumulative_profit_line: 8,
  cumulative_loss_line: 5,
  cumulative_days: 5,
  long_term_profit_line: 20,
  long_term_loss_line: 15,
  capital_scale_preset: 'MEDIUM' as PresetType,
})

const preset = ref<PresetType>('MEDIUM')
const maFilterEnabled = ref(false)

const { loading: advising, result: advice, actionClass, actionLabel, getAdvice, clearAdvice, calcLotDays } = useFundAdvisor()
const totalCapital = ref<number | undefined>(undefined)
const indexMAStatus = ref<IndexMAStatus>('unknown')

const capitalStorageKey = 'fund_advisor_total_capital'
const maFilterStorageKey = 'fund_advisor_ma_filter_enabled'

// Lot management
const lotFormVisible = ref(false)
const lotForm = ref({ quantity: 0, cost: 0, purchased_at: '' })
const lotLoading = ref(false)

function loadCapitalConfig() {
  const saved = localStorage.getItem(capitalStorageKey)
  if (saved) {
    const n = parseFloat(saved)
    if (!isNaN(n) && n > 0) {
      totalCapital.value = n
    }
  }
  const maSaved = localStorage.getItem(maFilterStorageKey)
  if (maSaved) {
    maFilterEnabled.value = maSaved === 'true'
  }
}

function saveCapitalConfig() {
  if (totalCapital.value && totalCapital.value > 0) {
    localStorage.setItem(capitalStorageKey, String(totalCapital.value))
  } else {
    localStorage.removeItem(capitalStorageKey)
  }
  localStorage.setItem(maFilterStorageKey, String(maFilterEnabled.value))
}

// 设计说明：applyPreset 是资金规模联动的核心实现。
// SMALL(1-5万)：高频小波段，阈值收紧至1.8%，关闭MA过滤；
// MEDIUM(5-30万)：标准网格，阈值2.5%，标准风控；
// LARGE(30万+)：大资金长线，阈值放宽至3.5%，强制开启MA过滤防范系统性风险。
function applyPreset(p: PresetType) {
  settingsForm.value.capital_scale_preset = p
  if (p === 'SMALL') {
    settingsForm.value.daily_profit_line = 1.8
    settingsForm.value.daily_loss_line = 1.8
    settingsForm.value.cumulative_loss_line = 4.0
    maFilterEnabled.value = false
  } else if (p === 'MEDIUM') {
    settingsForm.value.daily_profit_line = 2.5
    settingsForm.value.daily_loss_line = 2.5
    settingsForm.value.cumulative_loss_line = 5.0
    // MA 保持用户原有选择
  } else if (p === 'LARGE') {
    settingsForm.value.daily_profit_line = 3.5
    settingsForm.value.daily_loss_line = 3.5
    settingsForm.value.cumulative_loss_line = 6.0
    maFilterEnabled.value = true
  }
}

watch(preset, applyPreset)

function onGetAdvice() {
  if (!detail.value) return
  saveCapitalConfig()
  // Bug 修复：将用户当前在设置表单中修改的阈值合并到 detail，
  // 确保策略建议使用最新值而非仅后端已保存值。
  const mergedDetail = {
    ...detail.value,
    daily_profit_line: settingsForm.value.daily_profit_line,
    daily_loss_line: settingsForm.value.daily_loss_line,
    cumulative_profit_line: settingsForm.value.cumulative_profit_line,
    cumulative_loss_line: settingsForm.value.cumulative_loss_line,
    cumulative_days: settingsForm.value.cumulative_days,
    long_term_profit_line: settingsForm.value.long_term_profit_line,
    long_term_loss_line: settingsForm.value.long_term_loss_line,
    capital_scale_preset: settingsForm.value.capital_scale_preset,
  }
  getAdvice({
    detail: mergedDetail,
    totalCapital: totalCapital.value,
    indexMAStatus: indexMAStatus.value,
    lotSummary: detail.value.lot_summary,
    maFilterEnabled: maFilterEnabled.value,
  })
}

function openLotForm() {
  lotForm.value = { quantity: 0, cost: 0, purchased_at: new Date().toISOString().split('T')[0] }
  lotFormVisible.value = true
}

async function onAddLot() {
  if (!props.fundCode || lotForm.value.quantity <= 0) return
  lotLoading.value = true
  try {
    await createLot(props.fundCode, { ...lotForm.value })
    ElMessage.success('添加批次成功')
    lotFormVisible.value = false
    await fetchDetail()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '添加失败')
  } finally {
    lotLoading.value = false
  }
}

async function onDeleteLot(lot: FundLot) {
  try {
    await deleteLot(props.fundCode, lot.id)
    ElMessage.success('删除成功')
    await fetchDetail()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '删除失败')
  }
}

function feeBadgeClass(days: number): string {
  if (days >= 30) return 'badge-free'
  if (days >= 7) return 'badge-reduced'
  return 'badge-penalty'
}

function feeBadgeText(days: number): string {
  if (days >= 30) return '0%'
  if (days >= 7) return '0.5%'
  return '1.5%'
}

async function fetchDetail() {
  if (!props.fundCode) return
  loading.value = true
  try {
    detail.value = await getFundDetail(props.fundCode)
    if (detail.value) {
      const p = detail.value.capital_scale_preset || 'MEDIUM'
      const validPreset = (p === 'SMALL' || p === 'MEDIUM' || p === 'LARGE' || p === 'CUSTOM') ? p as PresetType : 'MEDIUM'
      preset.value = validPreset
      settingsForm.value = {
        daily_profit_line: detail.value.daily_profit_line ?? 2.5,
        daily_loss_line: detail.value.daily_loss_line ?? 2.5,
        cumulative_profit_line: detail.value.cumulative_profit_line ?? 8,
        cumulative_loss_line: detail.value.cumulative_loss_line ?? 5,
        cumulative_days: detail.value.cumulative_days ?? 5,
        long_term_profit_line: detail.value.long_term_profit_line ?? 20,
        long_term_loss_line: detail.value.long_term_loss_line ?? 15,
        capital_scale_preset: validPreset,
      }
    }
  } catch (e) {
    detail.value = null
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  saving.value = true
  try {
    await updateFund(props.fundCode, settingsForm.value)
    ElMessage.success('保存成功')
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

async function onSync() {
  if (!props.fundCode) return
  syncing.value = true
  try {
    await syncHoldings(props.fundCode)
    ElMessage.success('同步成功')
    await fetchDetail()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '同步失败')
  } finally {
    syncing.value = false
  }
}

function priceClass(n?: number) {
  if (n === undefined || n === null) return 'text-muted'
  if (n > 0) return 'text-up'
  if (n < 0) return 'text-down'
  return 'text-muted'
}

watch(() => props.visible, (v) => {
  if (v) {
    fetchDetail()
    loadCapitalConfig()
    clearAdvice()
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
    :title="`${fundName} (${fundCode})`"
    size="600px"
  >
    <div v-loading="loading" class="drawer-content">
      <div class="detail-header" v-if="detail">
        <div class="est-change-box">
          <div class="est-label">实际净值涨跌</div>
          <div class="est-value" :class="priceClass(detail.actual_change)">
            {{ formatPercent(detail.actual_change) }}
          </div>
        </div>
        <div class="est-change-box">
          <div class="est-label">{{ detail.fund_type === 'INDEX' ? '指数跟踪误差' : '持仓时效偏差' }}</div>
          <div class="est-value" :class="priceClass(detail.fund_type === 'INDEX' ? detail.tracking_error_index : detail.deviation)">
            {{ formatPercent(detail.fund_type === 'INDEX' ? detail.tracking_error_index : detail.deviation) }}
          </div>
        </div>
        <div class="est-change-box">
          <div class="est-label">持仓总收益</div>
          <div class="est-value" :class="priceClass(detail.total_return_est)">
            {{ formatPercent(detail.total_return_est) }}
          </div>
        </div>
        <div class="est-change-box">
          <div class="est-label">最近买入天数</div>
          <div class="est-value">
            <span :class="(detail.min_hold_days || 0) < 7 ? 'text-danger' : 'text-success'">{{ detail.min_hold_days ?? 0 }} 天</span>
          </div>
        </div>
        <el-button v-if="detail?.fund_type !== 'INDEX'" type="primary" :loading="syncing" @click="onSync">
          <el-icon><Refresh /></el-icon> 同步持仓
        </el-button>
      </div>

      <!-- Strategy Advisor -->
      <div class="advisor-section" v-if="detail">
        <div class="advisor-title">
          <el-icon :size="16"><TrendCharts /></el-icon>
          <span>策略建议 (14:50 决策辅助)</span>
        </div>
        <div class="advisor-config">
          <div class="advisor-config-item">
            <span class="advisor-config-label">总资金</span>
            <el-input-number v-model="totalCapital" :min="0" :precision="2" placeholder="可选，用于计算仓位" style="width: 160px;" />
          </div>
          <div class="advisor-config-item">
            <span class="advisor-config-label">均线过滤</span>
            <el-switch v-model="maFilterEnabled" active-text="开启" inactive-text="关闭" />
          </div>
          <div class="advisor-config-item" :class="{ disabled: !maFilterEnabled }">
            <span class="advisor-config-label">大盘20日均线</span>
            <el-select v-model="indexMAStatus" style="width: 120px;" :disabled="!maFilterEnabled">
              <el-option label="未知" value="unknown" />
              <el-option label="上方" value="above" />
              <el-option label="下方" value="below" />
            </el-select>
          </div>
          <el-button type="primary" plain :loading="advising" @click="onGetAdvice">
            <el-icon><MagicStick /></el-icon> 获取决策建议
          </el-button>
        </div>

        <!-- 指数基金汇率微缩提示 -->
        <div v-if="detail.fund_type === 'INDEX' && detail.base_currency && detail.base_currency !== 'CNY'" class="fx-hint">
          <el-icon :size="12" color="#6b7280"><InfoFilled /></el-icon>
          <span>
            {{ detail.base_currency === 'USD' ? '离岸人民币(CNH)' : '在岸人民币(CNY)' }}日内: {{ formatPercent(detail.fx_daily_change, 4) }}
            <span v-if="(detail.fx_daily_change || 0) > 0" class="fx-hint-press">（压制净值）</span>
            <span v-else-if="(detail.fx_daily_change || 0) < 0" class="fx-hint-support">（支撑净值）</span>
            <span v-else>（汇率中性）</span>
          </span>
        </div>

        <div class="advisor-cumulative" v-if="detail">
          <template v-if="detail.fund_type === 'INDEX'">
            <div class="advisor-cumulative-item">
              <span class="advisor-cumulative-label">关联指数涨跌</span>
              <span class="advisor-cumulative-value" :class="priceClass(detail.related_index_return)">
                {{ formatPercent(detail.related_index_return) }}
              </span>
            </div>
            <div class="advisor-cumulative-item">
              <span class="advisor-cumulative-label">汇率影响</span>
              <span class="advisor-cumulative-value" :class="priceClass(detail.fx_daily_change)">
                {{ formatPercent(detail.fx_daily_change, 4) }}
              </span>
            </div>
            <div class="advisor-cumulative-item">
              <span class="advisor-cumulative-label">折算后涨跌</span>
              <span class="advisor-cumulative-value" :class="priceClass(detail.est_daily_change)">
                {{ formatPercent(detail.est_daily_change) }}
              </span>
            </div>
            <div class="advisor-cumulative-item">
              <span class="advisor-cumulative-label">跟踪误差</span>
              <span class="advisor-cumulative-value" :class="priceClass(detail.tracking_error_index)">
                {{ formatPercent(detail.tracking_error_index) }}
              </span>
            </div>
          </template>
          <template v-else>
            <div class="advisor-cumulative-item">
              <span class="advisor-cumulative-label">单日阈值 (轨道A)</span>
              <span class="advisor-cumulative-value">±{{ detail.daily_profit_line ?? 2.5 }}%</span>
            </div>
            <div class="advisor-cumulative-item">
              <span class="advisor-cumulative-label">滚动复合 (轨道B)</span>
              <span class="advisor-cumulative-value" :class="priceClass(detail.rolling_cumulative_return)">
                {{ formatPercent(detail.rolling_cumulative_return) }}
              </span>
            </div>
            <div class="advisor-cumulative-item">
              <span class="advisor-cumulative-label">战略收益</span>
              <span class="advisor-cumulative-value" :class="priceClass(detail.total_return_est)">
                {{ formatPercent(detail.total_return_est) }}
              </span>
            </div>
          </template>
        </div>

        <div v-if="advice" class="advisor-result" :class="actionClass">
          <div class="advisor-result-header">
            <div class="advisor-action-badge">{{ actionLabel }}</div>
            <div class="advisor-confidence">
              <span class="confidence-label">置信度</span>
              <el-progress :percentage="advice.confidence" :color="advice.confidence >= 80 ? '#10b981' : advice.confidence >= 60 ? '#f59e0b' : '#6b7280'" :stroke-width="8" style="width: 120px;" />
            </div>
          </div>
          <div class="advisor-result-title">{{ advice.reasonTitle }}</div>
          <div class="advisor-result-detail">{{ advice.reasonDetail }}</div>
          <div v-if="advice.suggestedAmount" class="advisor-suggested">
            <span class="suggested-label">建议金额</span>
            <span class="suggested-value">{{ advice.suggestedAmount }}</span>
          </div>
          <div v-if="advice.suggestedQuantity" class="advisor-suggested">
            <span class="suggested-label">建议份额</span>
            <span class="suggested-value">{{ formatNumber(advice.suggestedQuantity) }}</span>
          </div>
          <div v-if="advice.feeNote" class="advisor-fee-note">{{ advice.feeNote }}</div>
        </div>
      </div>

      <!-- Fund Lots -->
      <div class="lots-section" v-if="detail">
        <div class="lots-header">
          <div class="lots-title">
            <el-icon :size="16"><Wallet /></el-icon>
            <span>持仓批次</span>
          </div>
          <el-button size="small" plain @click="openLotForm">
            <el-icon><Plus /></el-icon> 新增批次
          </el-button>
        </div>

        <div v-if="detail.lot_summary?.lots?.length" class="lots-summary">
          <div class="lots-stat">
            <span class="lots-stat-label">总份额</span>
            <span class="lots-stat-value">{{ formatNumber(detail.lot_summary.total_quantity) }}</span>
          </div>
          <div class="lots-stat">
            <span class="lots-stat-label">免手续费</span>
            <span class="lots-stat-value text-success">{{ formatNumber(detail.lot_summary.fee_free_quantity) }}</span>
          </div>
          <div class="lots-stat">
            <span class="lots-stat-label">惩罚费率</span>
            <span class="lots-stat-value text-danger">{{ formatNumber(detail.lot_summary.fee_penalty_quantity) }}</span>
          </div>
        </div>

        <el-table v-if="detail.lot_summary?.lots?.length" :data="detail.lot_summary.lots" size="small" style="width: 100%;">
          <el-table-column prop="purchased_at" label="购买日期" width="120">
            <template #default="{ row }">
              <span>{{ formatDate(row.purchased_at) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="quantity" label="份额" width="100">
            <template #default="{ row }">
              <span>{{ formatNumber(row.quantity) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="cost" label="成本" width="100">
            <template #default="{ row }">
              <span>{{ formatNumber(row.cost) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="持有天数" width="90">
            <template #default="{ row }">
              <span>{{ calcLotDays(row.purchased_at) }} 天</span>
            </template>
          </el-table-column>
          <el-table-column label="费率" width="80">
            <template #default="{ row }">
              <span class="fee-badge" :class="feeBadgeClass(calcLotDays(row.purchased_at))">
                {{ feeBadgeText(calcLotDays(row.purchased_at)) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="70">
            <template #default="{ row }">
              <el-button size="small" type="danger" link @click="onDeleteLot(row)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-else description="暂无批次记录，请先新增" />

        <div v-if="lotFormVisible" class="lot-form">
          <div class="lot-form-row">
            <div class="lot-form-item">
              <span class="lot-form-label">购买日期</span>
              <el-date-picker v-model="lotForm.purchased_at" type="date" value-format="YYYY-MM-DD" style="width: 140px;" />
            </div>
            <div class="lot-form-item">
              <span class="lot-form-label">份额</span>
              <el-input-number v-model="lotForm.quantity" :min="0" :step="100" style="width: 120px;" />
            </div>
            <div class="lot-form-item">
              <span class="lot-form-label">成本</span>
              <el-input-number v-model="lotForm.cost" :min="0" :precision="4" style="width: 120px;" />
            </div>
          </div>
          <div class="lot-form-actions">
            <el-button size="small" @click="lotFormVisible = false">取消</el-button>
            <el-button size="small" type="primary" :loading="lotLoading" @click="onAddLot">确认</el-button>
          </div>
        </div>
      </div>

      <div v-if="detail?.fund_type !== 'INDEX'">
        <el-table
          v-if="detail && detail.holdings && detail.holdings.length > 0"
          :data="detail.holdings"
          style="width: 100%;"
          :header-cell-style="{ background: '#f9fafb', fontWeight: 600, color: '#374151' }"
          size="small"
        >
          <el-table-column prop="stock_code" label="股票代码" width="100">
            <template #default="{ row }">
              <span class="code-text">{{ row.stock_code }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="stock_name" label="股票名称" width="110">
            <template #default="{ row }">
              <span class="name-text">{{ row.stock_name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="hold_ratio" label="持仓占比" width="100">
            <template #default="{ row }">
              <div class="ratio-cell">
                <div class="ratio-bar-bg">
                  <div class="ratio-bar" :style="{ width: Math.min(row.hold_ratio, 100) + '%' }"></div>
                </div>
                <span class="ratio-text">{{ formatNumber(row.hold_ratio) }}%</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="current_price" label="当前价" width="90">
            <template #default="{ row }">
              <span class="price-text">{{ formatNumber(row.current_price) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="change_percent" label="涨跌幅" width="90">
            <template #default="{ row }">
              <span :class="priceClass(row.change_percent)">{{ formatPercent(row.change_percent) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="contribution" label="贡献度" width="90">
            <template #default="{ row }">
              <span :class="priceClass(row.contribution)">{{ (row.contribution * 100).toFixed(2) }}%</span>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-else description="暂无持仓数据，请先同步" />

        <div class="sync-time" v-if="detail?.synced_at">
          <el-icon :size="14" color="#9ca3af"><Clock /></el-icon>
          <span>同步时间：{{ formatDateTime(detail.synced_at) }}</span>
        </div>
        <div class="sync-time" v-if="detail?.report_date">
          <el-icon :size="14" color="#9ca3af"><Document /></el-icon>
          <span>持仓报告期：{{ formatDate(detail.report_date) }}</span>
        </div>
      </div>

      <!-- 设计说明：观察指标设置面板是用户策略配置的 UI 入口。
           资金规模预设采用三档制（SMALL/MEDIUM/LARGE），
           切换时通过 applyPreset 联动修改阈值与 MA 过滤开关。
           长线战略止盈/止损线独立于短线阈值，构成"战略红线"。 -->
      <div class="settings-section" v-if="detail">
        <div class="settings-title">观察指标设置</div>
        <div class="settings-row">
          <div class="settings-item">
            <span class="settings-label">资金规模与激进程度预设</span>
            <el-select v-model="preset" style="width: 100%;">
              <el-option label="自定义" value="CUSTOM" />
              <el-option label="小额高频 (1-5万)" value="SMALL" />
              <el-option label="标准网格 (5-30万)" value="MEDIUM" />
              <el-option label="大额长线 (30万+)" value="LARGE" />
            </el-select>
          </div>
        </div>
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
            <span class="settings-label">长线战略止盈线%</span>
            <el-input-number v-model="settingsForm.long_term_profit_line" :precision="2" :min="0" style="width: 100%;" />
          </div>
          <div class="settings-item">
            <span class="settings-label">长线战略止损/熔断线%</span>
            <el-input-number v-model="settingsForm.long_term_loss_line" :precision="2" :min="0" style="width: 100%;" />
          </div>
        </div>
        <div class="settings-row">
          <div class="settings-item">
            <span class="settings-label">累计天数</span>
            <el-input-number v-model="settingsForm.cumulative_days" :min="1" style="width: 100%;" />
          </div>
        </div>
        <div class="settings-actions">
          <el-button type="primary" :loading="saving" @click="saveSettings">保存设置</el-button>
        </div>
      </div>
    </div>
  </el-drawer>
</template>

<style scoped>
.drawer-content {
  padding: 8px 4px;
}

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
  padding: 16px;
  background: linear-gradient(135deg, #f5f3ff 0%, #ede9fe 100%);
  border-radius: 12px;
  flex-wrap: wrap;
}

.est-change-box {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.est-label {
  font-size: 13px;
  color: #6b7280;
}

.est-value {
  font-size: 22px;
  font-weight: 800;
}

.code-text {
  font-family: 'SF Mono', monospace;
  font-weight: 600;
  color: #374151;
}

.name-text {
  font-weight: 600;
  color: #111827;
}

.ratio-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ratio-bar-bg {
  width: 40px;
  height: 5px;
  background: #f3f4f6;
  border-radius: 3px;
  overflow: hidden;
}

.ratio-bar {
  height: 100%;
  background: linear-gradient(90deg, #8b5cf6 0%, #a78bfa 100%);
  border-radius: 3px;
}

.ratio-text {
  font-size: 12px;
  font-weight: 600;
  color: #8b5cf6;
  min-width: 46px;
}

.price-text {
  font-weight: 600;
  color: #374151;
}

.sync-time {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 16px;
  color: #9ca3af;
  font-size: 13px;
}

.settings-section {
  margin-top: 24px;
  padding: 16px;
  background: #f9fafb;
  border-radius: 12px;
}

.settings-title {
  font-size: 15px;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 16px;
}

.settings-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 12px;
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

.advisor-section {
  margin-top: 20px;
  padding: 16px;
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
}

.advisor-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 12px;
}

.advisor-config {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.advisor-config-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.advisor-config-item.disabled {
  opacity: 0.5;
}

.advisor-cumulative {
  display: flex;
  gap: 16px;
  margin-bottom: 12px;
  padding: 10px 12px;
  background: #f9fafb;
  border-radius: 8px;
}

.advisor-cumulative-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.advisor-cumulative-label {
  font-size: 11px;
  color: #9ca3af;
}

.advisor-cumulative-value {
  font-size: 14px;
  font-weight: 700;
  color: #111827;
}

.advisor-config-label {
  font-size: 13px;
  color: #6b7280;
  font-weight: 500;
}

.advisor-result {
  margin-top: 12px;
  padding: 16px;
  border-radius: 10px;
  border-left: 4px solid #9ca3af;
  background: #f9fafb;
}

.advisor-result.advisor-buy {
  border-left-color: #ef4444;
  background: rgba(239, 68, 68, 0.04);
}

.advisor-result.advisor-sell {
  border-left-color: #10b981;
  background: rgba(16, 185, 129, 0.04);
}

.advisor-result.advisor-warning {
  border-left-color: #f59e0b;
  background: rgba(245, 158, 11, 0.04);
}

.advisor-result.advisor-hold {
  border-left-color: #6b7280;
  background: #f9fafb;
}

.advisor-result.advisor-meltdown {
  border-left-color: #dc2626;
  background: rgba(239, 68, 68, 0.08);
}

.advisor-result.advisor-strategic-sell {
  border-left-color: #059669;
  background: rgba(16, 185, 129, 0.08);
}

.advisor-result.advisor-hold-lock {
  border-left-color: #f59e0b;
  background: rgba(245, 158, 11, 0.06);
}

.advisor-result.advisor-hold-max {
  border-left-color: #9ca3af;
  background: #f9fafb;
}

.advisor-result.advisor-hold-break {
  border-left-color: #dc2626;
  background: rgba(239, 68, 68, 0.06);
}

.advisor-result.advisor-hold-fx-lock {
  border-left-color: #f59e0b;
  background: rgba(245, 158, 11, 0.06);
}

.advisor-result.advisor-hold-fx-buffer {
  border-left-color: #9ca3af;
  background: #f9fafb;
}

.fx-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 10px;
  font-size: 12px;
  color: #6b7280;
}

.fx-hint-press {
  color: #ef4444;
}

.fx-hint-support {
  color: #10b981;
}

.advisor-result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.advisor-action-badge {
  padding: 4px 12px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 700;
  color: #ffffff;
  background: #6b7280;
}

.advisor-buy .advisor-action-badge {
  background: #ef4444;
}

.advisor-sell .advisor-action-badge {
  background: #10b981;
}

.advisor-warning .advisor-action-badge {
  background: #f59e0b;
}

.advisor-hold .advisor-action-badge {
  background: #6b7280;
}

.advisor-meltdown .advisor-action-badge {
  background: #dc2626;
}

.advisor-strategic-sell .advisor-action-badge {
  background: #059669;
}

.advisor-hold-lock .advisor-action-badge {
  background: #f59e0b;
}

.advisor-hold-max .advisor-action-badge {
  background: #9ca3af;
}

.advisor-hold-break .advisor-action-badge {
  background: #dc2626;
}

.advisor-hold-fx-lock .advisor-action-badge {
  background: #f59e0b;
}

.advisor-hold-fx-buffer .advisor-action-badge {
  background: #9ca3af;
}

.advisor-confidence {
  display: flex;
  align-items: center;
  gap: 8px;
}

.confidence-label {
  font-size: 12px;
  color: #6b7280;
}

.advisor-result-title {
  font-size: 14px;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 6px;
}

.advisor-result-detail {
  font-size: 13px;
  color: #4b5563;
  line-height: 1.6;
}

.advisor-suggested {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px dashed #e5e7eb;
  display: flex;
  align-items: center;
  gap: 8px;
}

.suggested-label {
  font-size: 12px;
  color: #6b7280;
}

.suggested-value {
  font-size: 16px;
  font-weight: 800;
  color: #111827;
}

.advisor-fee-note {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px dashed #e5e7eb;
  font-size: 12px;
  color: #6b7280;
  line-height: 1.5;
}

.lots-section {
  margin-top: 20px;
  padding: 16px;
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
}

.lots-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.lots-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 700;
  color: #1f2937;
}

.lots-summary {
  display: flex;
  gap: 16px;
  margin-bottom: 12px;
  padding: 12px;
  background: #f9fafb;
  border-radius: 8px;
}

.lots-stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.lots-stat-label {
  font-size: 11px;
  color: #9ca3af;
}

.lots-stat-value {
  font-size: 14px;
  font-weight: 700;
  color: #111827;
}

.fee-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.fee-badge.badge-free {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.fee-badge.badge-reduced {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.fee-badge.badge-penalty {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.lot-form {
  margin-top: 12px;
  padding: 12px;
  background: #f9fafb;
  border-radius: 8px;
}

.lot-form-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.lot-form-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.lot-form-label {
  font-size: 12px;
  color: #6b7280;
}

.lot-form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 12px;
}

@media (max-width: 768px) {
  .settings-row {
    grid-template-columns: 1fr;
  }

  .advisor-config {
    flex-direction: column;
    align-items: flex-start;
  }

  .lot-form-row {
    flex-direction: column;
    align-items: flex-start;
  }

  .lots-summary {
    flex-wrap: wrap;
  }
}
</style>
