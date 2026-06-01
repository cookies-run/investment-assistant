import client from './client'

export interface PortfolioStats {
  total_cost: number
  total_market_value: number
  total_pnl: number
  total_pnl_percent: number
  stock_count: number
  fund_count: number
}

export async function getStats(): Promise<PortfolioStats> {
  const res = await client.get('/stats')
  return res.data
}
