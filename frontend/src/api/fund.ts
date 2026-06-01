import client from './client'

export interface Fund {
  fund_code: string
  fund_name: string
  hold_cost: number
  hold_quantity: number
  daily_profit_line: number
  daily_loss_line: number
  cumulative_profit_line: number
  cumulative_loss_line: number
  cumulative_days: number
  long_term_profit_line?: number
  long_term_loss_line?: number
  capital_scale_preset?: string
  fund_type?: string
  related_index_symbol?: string
  base_currency?: string
  is_active: boolean
  created_at: string
  updated_at: string
  current_nav?: number
  change_percent?: number
  market_value?: number
  total_pnl?: number
  total_pnl_percent?: number
}

export async function listFunds(): Promise<Fund[]> {
  const res = await client.get('/funds')
  return res.data
}

export async function createFund(data: Partial<Fund>) {
  return client.post('/funds', data)
}

export async function updateFund(code: string, data: Partial<Fund>) {
  return client.put(`/funds/${code}`, data)
}

export async function deleteFund(code: string) {
  return client.delete(`/funds/${code}`)
}

export interface FundSearchResult {
  code: string
  name: string
}

export async function searchFunds(q: string): Promise<FundSearchResult[]> {
  const res = await client.get('/funds/search', { params: { q } })
  return res.data
}

export interface DailyRecord {
  id: number
  target_code: string
  target_type: string
  trade_date: string
  open_price?: number
  close_price?: number
  change_percent: number
  cumulative_percent: number
  created_at: string
}

export async function getFundDailyRecords(code: string): Promise<DailyRecord[]> {
  const res = await client.get(`/funds/${code}/daily-records`)
  return res.data
}
