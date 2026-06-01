import client from './client'

export interface FundHolding {
  id: number
  fund_code: string
  stock_code: string
  stock_name: string
  hold_ratio: number
  report_date: string
  synced_at: string
}

export interface FundHoldingSpot {
  stock_code: string
  stock_name: string
  hold_ratio: number
  current_price: number
  change_percent: number
  contribution: number
}

import type { LotSummary } from './fundLot'

export interface FundDetail {
  fund_code: string
  fund_name: string
  hold_cost?: number
  hold_quantity?: number
  daily_profit_line?: number
  daily_loss_line?: number
  cumulative_profit_line?: number
  cumulative_loss_line?: number
  cumulative_days?: number
  long_term_profit_line?: number
  long_term_loss_line?: number
  capital_scale_preset?: string
  is_active?: boolean
  estimated_change: number
  actual_change?: number
  deviation?: number
  current_nav?: number
  total_return_est?: number
  min_hold_days?: number
  rolling_cumulative_return?: number
  fund_type?: string
  related_index_symbol?: string
  base_currency?: string
  related_index_return?: number
  fx_daily_change?: number
  tracking_error_index?: number
  est_daily_change?: number
  synced_at: string
  created_at: string
  report_date?: string
  lot_summary?: LotSummary
  holdings: FundHoldingSpot[]
}

export async function getHoldings(code: string): Promise<FundHolding[]> {
  const res = await client.get(`/fund-holdings/${code}`)
  return res.data
}

export async function getFundDetail(code: string): Promise<FundDetail> {
  const res = await client.get(`/fund-holdings/detail/${code}`)
  return res.data
}

export async function syncHoldings(code: string) {
  return client.post(`/fund-holdings/sync/${code}`)
}
