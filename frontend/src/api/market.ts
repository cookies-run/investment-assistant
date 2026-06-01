import client from './client'

export interface MarketDataItem {
  symbol: string
  name: string
  price: number
  change: number
  change_percent: number
  trade_date: string
  ma20?: number
}

export interface MarketDataCategory {
  name: string
  items: MarketDataItem[]
}

export interface MarketDashboard {
  update_time: string
  categories: MarketDataCategory[]
}

export interface MarketGroupItem {
  id: number
  group_id: number
  symbol: string
  name: string
  source_type: string
  sort_order: number
}

export interface MarketGroup {
  id: number
  name: string
  sort_order: number
  items: MarketGroupItem[]
}

export interface AvailableIndex {
  symbol: string
  name: string
  source_type: string
  category: string
}

export async function getMarketDashboard(): Promise<MarketDashboard> {
  const res = await client.get('/market-dashboard')
  return res.data
}

export async function getMarketGroups(): Promise<MarketGroup[]> {
  const res = await client.get('/market-groups')
  return res.data
}

export async function createMarketGroup(name: string): Promise<MarketGroup> {
  const res = await client.post('/market-groups', { name })
  return res.data
}

export async function updateMarketGroup(id: number, name: string): Promise<MarketGroup> {
  const res = await client.put(`/market-groups/${id}`, { name })
  return res.data
}

export async function deleteMarketGroup(id: number): Promise<void> {
  await client.delete(`/market-groups/${id}`)
}

export async function reorderMarketGroups(ids: number[]): Promise<void> {
  await client.post('/market-groups/reorder', { ids })
}

export async function createMarketItem(groupId: number, item: { symbol: string; name: string; source_type: string }): Promise<MarketGroupItem> {
  const res = await client.post('/market-items', { group_id: groupId, ...item })
  return res.data
}

export async function deleteMarketItem(id: number): Promise<void> {
  await client.delete(`/market-items/${id}`)
}

export async function reorderMarketItems(ids: number[]): Promise<void> {
  await client.post('/market-items/reorder', { ids })
}

export async function getAvailableIndices(): Promise<AvailableIndex[]> {
  const res = await client.get('/available-indices')
  return res.data
}
