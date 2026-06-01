import client from './client'

export interface TrendPoint {
  date: string
  open: number
  close: number
  high: number
  low: number
  volume: number
}

export interface MinutePoint {
  time: string
  price: number
  volume: number
  turnover: number
}

export interface LevelInfo {
  price: number
  volume: number
}

export interface Stock {
  stock_code: string
  stock_name: string
  buy_price: number
  hold_quantity: number
  daily_profit_line: number
  daily_loss_line: number
  cumulative_profit_line: number
  cumulative_loss_line: number
  cumulative_days: number
  monitor_interval: number
  is_active: boolean
  created_at: string
  updated_at: string
  current_price?: number
  change_percent?: number
  market_value?: number
  total_pnl?: number
  total_pnl_percent?: number
  volume?: number
  buy_volume?: number
  sell_volume?: number
  buy_sell_diff?: number
  trends?: TrendPoint[]
}

export interface StockDetail {
  stock_code: string
  stock_name: string
  buy_price?: number
  hold_quantity?: number
  current_price?: number
  change_percent?: number
  volume?: number
  turnover?: number
  open?: number
  prev_close?: number
  high?: number
  low?: number
  buy_sell_diff?: number
  buy_levels: LevelInfo[]
  sell_levels: LevelInfo[]
  daily_profit_line?: number
  daily_loss_line?: number
  cumulative_profit_line?: number
  cumulative_loss_line?: number
  cumulative_days?: number
  monitor_interval?: number
  is_active?: boolean
  daily_klines: TrendPoint[]
  minute_trends: MinutePoint[]
  // 盘口博弈引擎字段
  active_buy_vol?: number
  active_sell_vol?: number
  large_order_net_inflow?: number
  minute_turnover_rate?: number
  minute_price_volatility?: number
}

export interface SearchResult {
  code: string
  name: string
  current_price?: number
  change_percent?: number
}

export async function listStocks(): Promise<Stock[]> {
  const res = await client.get('/stocks')
  return res.data
}

export async function createStock(data: Partial<Stock>) {
  return client.post('/stocks', data)
}

export async function updateStock(code: string, data: Partial<Stock>) {
  return client.put(`/stocks/${code}`, data)
}

export async function deleteStock(code: string) {
  return client.delete(`/stocks/${code}`)
}

export async function searchStocks(q: string): Promise<SearchResult[]> {
  const res = await client.get('/stocks/search', { params: { q } })
  return res.data
}

export async function getStockDetail(code: string): Promise<StockDetail> {
  const res = await client.get(`/stocks/detail/${code}`)
  return res.data
}
