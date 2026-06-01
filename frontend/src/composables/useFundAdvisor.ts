import { ref, computed } from 'vue'
import type { FundDetail } from '../api/holding'
import type { LotSummary } from '../api/fundLot'

// 设计说明： AdvisorAction 扩展为 9 种精细化动作类型，
// 覆盖 PRD 中全部 5 关决策分支。每种动作对应独立的 UI 样式与文案色彩。
export type AdvisorAction =
  | 'BUY'
  | 'SELL'
  | 'HOLD'
  | 'WARNING'
  | 'STRATEGIC_MELTDOWN'
  | 'SELL_STRATEGIC'
  | 'HOLD_LOCK'
  | 'HOLD_MAX'
  | 'HOLD_BREAK'
  | 'HOLD_FX_LOCK'
  | 'HOLD_FX_BUFFER'

export type IndexMAStatus = 'above' | 'below' | 'unknown'

export interface AdvisorResult {
  action: AdvisorAction
  confidence: number
  reasonTitle: string
  reasonDetail: string
  suggestedAmount?: string
  suggestedQuantity?: number
  feeNote?: string
}

export interface AdvisorInput {
  detail: FundDetail
  totalCapital?: number
  indexMAStatus?: IndexMAStatus
  lotSummary?: LotSummary
  maFilterEnabled?: boolean
}

function toNum(v: any): number {
  if (v === undefined || v === null) return 0
  if (typeof v === 'number') return v
  const n = parseFloat(String(v))
  return isNaN(n) ? 0 : n
}

function formatCurrency(n: number): string {
  return '¥' + n.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function calcPositionRatio(detail: FundDetail, totalCapital?: number): number | null {
  if (!totalCapital || totalCapital <= 0) return null
  const nav = toNum(detail.current_nav) || toNum(detail.hold_cost)
  const marketValue = toNum(detail.hold_quantity) * nav
  return (marketValue / totalCapital) * 100
}

// 设计说明：FundStrategyAdvisor 是 14:50 核心决策引擎的纯函数实现。
// 所有输入均来自后端 Detail 接口的一次性计算结果，前端不做任何复杂数学运算，
// 只做纯逻辑分支判断。决策树按 PRD 要求严格分为 5 关流水线，
// 一旦触发某一分支立即返回，确保优先级从上至下严格执行。
export function useFundAdvisor() {
  const loading = ref(false)
  const result = ref<AdvisorResult | null>(null)

  const actionClass = computed(() => {
    const map: Record<AdvisorAction, string> = {
      BUY: 'advisor-buy',
      SELL: 'advisor-sell',
      HOLD: 'advisor-hold',
      WARNING: 'advisor-warning',
      STRATEGIC_MELTDOWN: 'advisor-meltdown',
      SELL_STRATEGIC: 'advisor-strategic-sell',
      HOLD_LOCK: 'advisor-hold-lock',
      HOLD_MAX: 'advisor-hold-max',
      HOLD_BREAK: 'advisor-hold-break',
      HOLD_FX_LOCK: 'advisor-hold-fx-lock',
      HOLD_FX_BUFFER: 'advisor-hold-fx-buffer',
    }
    return result.value ? map[result.value.action] : ''
  })

  const actionLabel = computed(() => {
    const map: Record<AdvisorAction, string> = {
      BUY: '买入',
      SELL: '卖出',
      HOLD: '持有观望',
      WARNING: '风险提示',
      STRATEGIC_MELTDOWN: '战略熔断',
      SELL_STRATEGIC: '战略赎回',
      HOLD_LOCK: '持有锁定',
      HOLD_MAX: '仓位封顶',
      HOLD_BREAK: '趋势破位',
      HOLD_FX_LOCK: '汇率对冲锁定',
      HOLD_FX_BUFFER: '汇率缓冲观望',
    }
    return result.value ? map[result.value.action] : ''
  })

  function analyze(input: AdvisorInput): AdvisorResult {
    const { detail, totalCapital, indexMAStatus = 'unknown', lotSummary, maFilterEnabled } = input

    const actualChange = toNum(detail.actual_change)
    const deviation = toNum(detail.deviation)
    const dailyProfitLine = toNum(detail.daily_profit_line) || 2
    const dailyLossLine = toNum(detail.daily_loss_line) || 2
    const cumulativeProfitLine = toNum(detail.cumulative_profit_line) || 8
    const cumulativeLossLine = toNum(detail.cumulative_loss_line) || 8
    const longTermProfitLine = toNum(detail.long_term_profit_line) || 20
    const longTermLossLine = toNum(detail.long_term_loss_line) || 15
    const rollingCumulativeReturn = toNum(detail.rolling_cumulative_return)
    const minHoldDays = toNum(detail.min_hold_days)
    const totalReturnEst = toNum(detail.total_return_est)
    const capitalScalePreset = detail.capital_scale_preset || 'MEDIUM'

    const holdQty = toNum(detail.hold_quantity)
    const currentNav = toNum(detail.current_nav) || toNum(detail.hold_cost)
    const marketValue = holdQty * currentNav
    const positionRatio = calcPositionRatio(detail, totalCapital)

    const feeFreeQty = toNum(lotSummary?.fee_free_quantity)

    const isCFund = /(\(C\)|C类|C份额)/i.test(detail.fund_name || '')
    const feeNote = isCFund
      ? 'C 类份额销售服务费约 0.4%/年，日均成本仅 ~0.0011%。日内择时产生的价差可轻松覆盖持有成本。'
      : '当前基金为 A 类/QDII 份额，打一折申购费率约 0.1%，持有满 30 天赎回费为 0%。日内择时产生的价差可覆盖交易成本。'

    // Effective MA status: only used when maFilterEnabled is true
    const effectiveMAStatus = maFilterEnabled ? indexMAStatus : 'unknown'

    // 指数基金模式下，使用后端计算的 est_daily_change（指数+汇率调整）替代原始 actual_change 做阈值判定
    const effectiveActualChange = detail.fund_type === 'INDEX' && detail.est_daily_change !== undefined
      ? toNum(detail.est_daily_change)
      : actualChange

    // ═══════════════════════════════════════════════════════
    // 【第 1 关】数据失真强风控
    // ═══════════════════════════════════════════════════════
    // 指数基金使用跟踪误差替代持仓偏差，避免无持仓明细导致的失真
    const effectiveDeviation = detail.fund_type === 'INDEX'
      ? toNum(detail.tracking_error_index)
      : deviation
    const deviationLabel = detail.fund_type === 'INDEX' ? '跟踪误差' : '估值偏差'
    if (Math.abs(effectiveDeviation) > 1.5) {
      const confidence = Math.min(Math.round((Math.abs(effectiveDeviation) / 1.5) * 60), 95)
      return {
        action: 'WARNING',
        confidence,
        reasonTitle: `${deviationLabel}过大`,
        reasonDetail: `基金近期出现大幅度调仓，当前${detail.fund_type === 'INDEX' ? '指数折算估值' : '估值信息'}严重失真（${deviationLabel} ${effectiveDeviation.toFixed(2)}%）。${detail.fund_type === 'INDEX' ? '实际净值与指数折算估值出现明显偏离，可能因成分股分红、QDII额度限制或汇率巨震导致。' : '实际净值与持仓估算出现明显偏离，'}建议放弃尾盘择时，卧倒观望。`,
        feeNote,
      }
    }

    // ═══════════════════════════════════════════════════════
    // 【第 2 关】战略级长线风控（最高优先级的长期红线）
    // ═══════════════════════════════════════════════════════
    if (totalReturnEst <= -longTermLossLine) {
      return {
        action: 'STRATEGIC_MELTDOWN',
        confidence: 95,
        reasonTitle: '【长期战略止损预警】',
        reasonDetail: `当前持仓累计亏损已达 ${totalReturnEst.toFixed(2)}%，触及长线硬止损/熔断线 ${longTermLossLine}%。系统已自动熔断所有短线补仓提示。请人工审查该基金经理表现及行业基本面是否发生根本性恶化，切勿盲目继续加仓。`,
        feeNote,
      }
    }

    if (totalReturnEst >= longTermProfitLine) {
      if (minHoldDays < 7) {
        return {
          action: 'HOLD_LOCK',
          confidence: 88,
          reasonTitle: '战略止盈点已到达，但持仓未满7天',
          reasonDetail: `当前总收益 ${totalReturnEst.toFixed(2)}% 已达到长期战略止盈点 ${longTermProfitLine}%，但检测到 ${minHoldDays} 天内有新补仓资金。为避免 1.5% 惩罚性赎回费，建议延迟至份额满 7 天后再执行大批量赎回。`,
          feeNote,
        }
      }

      const targetSellAmount = marketValue * 0.20
      const targetSellQty = holdQty * 0.20

      if (feeFreeQty > 0 && feeFreeQty < targetSellQty) {
        const adjustedSellQty = feeFreeQty
        const adjustedSellAmount = adjustedSellQty * currentNav
        return {
          action: 'SELL_STRATEGIC',
          confidence: 85,
          reasonTitle: '【长期战略止盈触发】',
          reasonDetail: `总持仓累计收益率已达 ${totalReturnEst.toFixed(2)}%，进入高估值风险区。但仅有 ${feeFreeQty.toFixed(2)} 份已满 30 天免赎回费，不足以覆盖建议卖出量。建议先卖出免手续费部分，其余持仓继续持有至满 30 天以规避赎回费。`,
          suggestedAmount: formatCurrency(adjustedSellAmount),
          suggestedQuantity: adjustedSellQty,
          feeNote,
        }
      }

      return {
        action: 'SELL_STRATEGIC',
        confidence: 90,
        reasonTitle: '【长期战略止盈触发】',
        reasonDetail: `总持仓累计收益率已达 ${totalReturnEst.toFixed(2)}%，进入高估值风险区。系统开启分批落袋模式，建议执行战略赎回，卖出当前总份额的 20%。`,
        suggestedAmount: formatCurrency(targetSellAmount),
        suggestedQuantity: targetSellQty,
        feeNote,
      }
    }

    // ═══════════════════════════════════════════════════════
    // 【第 2.5 关】跨境汇率与偏差异常风控（仅指数基金）
    // ═══════════════════════════════════════════════════════
    if (detail.fund_type === 'INDEX') {
      const relatedIndexReturn = toNum(detail.related_index_return)
      const estDailyChange = toNum(detail.est_daily_change)
      const trackingErrorIndex = toNum(detail.tracking_error_index)

      // Case C: 跟踪偏差过大（数据失真）
      if (Math.abs(trackingErrorIndex) > 1.0) {
        return {
          action: 'WARNING',
          confidence: 85,
          reasonTitle: '【时效偏差过大】',
          reasonDetail: `当前基金估值与指数实际走势出现超过 1% 的严重背离（tracking_error=${trackingErrorIndex.toFixed(2)}%）。可能由于成分股分红、QDII额度限制、盘中溢价或汇率剧烈巨震导致。当前盘面数据失真，建议放弃尾盘择时。`,
          feeNote,
        }
      }

      // Case A: 指数暴涨但被汇率对冲（防止假止盈）
      if (relatedIndexReturn >= dailyProfitLine && estDailyChange < dailyProfitLine) {
        return {
          action: 'HOLD_FX_LOCK',
          confidence: 78,
          reasonTitle: '【汇率对冲预警】',
          reasonDetail: `虽然底层指数盘中暴涨已达止盈线（${relatedIndexReturn.toFixed(2)}%），但由于今日人民币大幅升值，汇率波动稀释了海外资产净值，导致人民币实际净值未达标（${estDailyChange.toFixed(2)}%）。建议取消本次短线止盈，继续观望。`,
          feeNote,
        }
      }

      // Case B: 指数暴跌但被汇率对冲（防止假补仓）
      if (relatedIndexReturn <= -dailyLossLine && estDailyChange > -dailyLossLine) {
        return {
          action: 'HOLD_FX_BUFFER',
          confidence: 75,
          reasonTitle: '【汇率缓冲预警】',
          reasonDetail: `底层指数已跌入恐慌加仓区（${relatedIndexReturn.toFixed(2)}%），但由于今日人民币贬值带来汇兑收益，抵消了部分跌幅，实际人民币净值并未跌透（${estDailyChange.toFixed(2)}%）。当前位置补仓“性价比”不高，建议暂缓加仓。`,
          feeNote,
        }
      }
    }

    // ═══════════════════════════════════════════════════════
    // 【第 3 关】短线止盈判定轨道（单日突变 vs 多日滚动累进）
    // ═══════════════════════════════════════════════════════
    if (effectiveActualChange >= dailyProfitLine || rollingCumulativeReturn >= cumulativeProfitLine) {
      if (minHoldDays < 7) {
        return {
          action: 'HOLD_LOCK',
          confidence: 85,
          reasonTitle: '触发止盈线但持仓未满7天',
          reasonDetail: `触发短线止盈信号（单日涨幅 ${effectiveActualChange.toFixed(2)}% 或滚动累计 ${rollingCumulativeReturn.toFixed(2)}%），但检测到 ${minHoldDays} 天内有补仓行为，此时卖出将触发 1.5% 惩罚性赎回费。本次操作放弃。`,
          feeNote,
        }
      }

      const targetSellAmount = marketValue * 0.10
      const targetSellQty = holdQty * 0.10

      if (feeFreeQty > 0 && feeFreeQty < targetSellQty) {
        const adjustedSellQty = feeFreeQty
        const adjustedSellAmount = adjustedSellQty * currentNav
        return {
          action: 'SELL',
          confidence: 78,
          reasonTitle: '触发止盈线，建议减仓（已调整免手续费份额）',
          reasonDetail: `触发短线止盈信号，但仅有 ${feeFreeQty.toFixed(2)} 份已满 30 天免赎回费，不足以覆盖建议卖出量 ${targetSellQty.toFixed(2)} 份。建议先卖出免手续费部分，其余持仓继续持有至满 30 天以规避 0.5% 赎回费。`,
          suggestedAmount: formatCurrency(adjustedSellAmount),
          suggestedQuantity: adjustedSellQty,
          feeNote,
        }
      }

      return {
        action: 'SELL',
        confidence: 82,
        reasonTitle: '触发短线止盈信号',
        reasonDetail: `触发短线止盈信号（单日暴涨或多日连涨引发阳线衰竭），建议逢高落袋，赎回当前份额的 10%。`,
        suggestedAmount: formatCurrency(targetSellAmount),
        suggestedQuantity: targetSellQty,
        feeNote,
      }
    }

    // ═══════════════════════════════════════════════════════
    // 【第 4 关】短线补仓判定轨道（单日暴跌 vs 多日阴跌震荡）
    // ═══════════════════════════════════════════════════════
    if (effectiveActualChange <= -dailyLossLine || rollingCumulativeReturn <= -cumulativeLossLine) {
      const posRatio = positionRatio ?? 50

      // 仓位熔断检查
      if (posRatio >= 80) {
        return {
          action: 'HOLD_MAX',
          confidence: 70,
          reasonTitle: '触发加仓线但仓位已达上限',
          reasonDetail: `触发短线加仓线（单日跌幅 ${effectiveActualChange.toFixed(2)}% 或滚动累计 ${rollingCumulativeReturn.toFixed(2)}%），但你的总仓位已达 80% 的设定最高上限。为保证现金流安全，强制停止网格加仓。`,
          feeNote,
        }
      }

      // 趋势破位检查：LARGE 预设下，跌破20MA 阻断 BUY
      if (capitalScalePreset === 'LARGE' && effectiveMAStatus !== 'above') {
        return {
          action: 'HOLD_BREAK',
          confidence: 75,
          reasonTitle: '触发加仓线但趋势破位',
          reasonDetail: `触发加仓线，但当前关联指数已跌破 20 日均线，技术面呈右侧破位下行。大资金需防范连续无底阴跌，建议暂缓加仓，等待右侧企稳。`,
          feeNote,
        }
      }

      // 通过全部过滤器，执行补仓
      const buyAmount = totalCapital ? totalCapital * 0.05 : marketValue * 0.05
      return {
        action: 'BUY',
        confidence: 72,
        reasonTitle: '触发网格加仓点',
        reasonDetail: `触发网格加仓点（单日恐慌暴跌或 ${detail.cumulative_days ?? 5} 日持续阴跌达到震荡衰竭临界点），建议分批加仓总资金的 5% 以摊薄持仓成本。`,
        suggestedAmount: formatCurrency(buyAmount),
        feeNote,
      }
    }

    // ═══════════════════════════════════════════════════════
    // 【第 5 关】日常震荡兜底
    // ═══════════════════════════════════════════════════════
    return {
      action: 'HOLD',
      confidence: 50,
      reasonTitle: '日常波动，暂无明确信号',
      reasonDetail: `当前单日涨跌幅 ${effectiveActualChange.toFixed(2)}% 及多日滚动波动 ${rollingCumulativeReturn.toFixed(2)}% 均在设定阈值内，属于无趋势震荡行情。减少无意义的高频申赎，保持观望。`,
      feeNote,
    }
  }

  function getAdvice(input: AdvisorInput) {
    loading.value = true
    try {
      result.value = analyze(input)
    } finally {
      loading.value = false
    }
  }

  function calcLotDays(purchasedAt: string): number {
    const p = new Date(purchasedAt)
    const now = new Date()
    return Math.floor((now.getTime() - p.getTime()) / (1000 * 60 * 60 * 24))
  }

  function clearAdvice() {
    result.value = null
  }

  return {
    loading,
    result,
    actionClass,
    actionLabel,
    analyze,
    getAdvice,
    clearAdvice,
    calcLotDays,
  }
}
