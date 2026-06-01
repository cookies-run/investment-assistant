import client from './client'

export interface StockGroupItem {
  id: number
  group_id: number
  stock_code: string
  stock_name: string
  sort_order: number
}

export interface StockGroup {
  id: number
  name: string
  sort_order: number
  items: StockGroupItem[]
}

export async function getStockGroups(): Promise<StockGroup[]> {
  const res = await client.get('/stock-groups')
  return res.data
}

export async function createStockGroup(name: string): Promise<StockGroup> {
  const res = await client.post('/stock-groups', { name })
  return res.data
}

export async function updateStockGroup(id: number, name: string): Promise<StockGroup> {
  const res = await client.put(`/stock-groups/${id}`, { name })
  return res.data
}

export async function deleteStockGroup(id: number): Promise<void> {
  await client.delete(`/stock-groups/${id}`)
}

export async function reorderStockGroups(ids: number[]): Promise<void> {
  await client.post('/stock-groups/reorder', { ids })
}

export async function createStockGroupItem(groupId: number, item: { stock_code: string; stock_name: string }): Promise<StockGroupItem> {
  const res = await client.post('/stock-group-items', { group_id: groupId, ...item })
  return res.data
}

export async function deleteStockGroupItem(id: number): Promise<void> {
  await client.delete(`/stock-group-items/${id}`)
}

export async function reorderStockGroupItems(ids: number[]): Promise<void> {
  await client.post('/stock-group-items/reorder', { ids })
}
