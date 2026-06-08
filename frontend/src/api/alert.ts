import client from './client'

export interface AlertRecord {
  id: number
  target_code: string
  target_name: string
  target_type: string
  alert_type: string
  trigger_value: number
  threshold_value: number
  current_price?: number
  notify_status: string
  triggered_at: string
}

export async function listAlerts(params?: { target_code?: string; target_type?: string; limit?: number }): Promise<AlertRecord[]> {
  const res = await client.get('/alerts', { params })
  return res.data
}
