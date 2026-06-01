import client from './client'

export interface NotificationConfig {
  id: number
  feishu_webhook: string
  enable_feishu: boolean
}

export async function getNotificationConfig(): Promise<NotificationConfig> {
  const res = await client.get('/notification-config')
  return res.data
}

export async function saveNotificationConfig(config: Partial<NotificationConfig>): Promise<NotificationConfig> {
  const res = await client.post('/notification-config', config)
  return res.data
}
