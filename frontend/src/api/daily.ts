import client from './client'

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

export async function listDailyRecords(params?: { target_code?: string; target_type?: string; limit?: number }): Promise<DailyRecord[]> {
  const res = await client.get('/daily-records', { params })
  return res.data
}
