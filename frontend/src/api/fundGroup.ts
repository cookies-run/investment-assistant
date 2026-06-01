import client from './client'

export interface FundGroupItem {
  id: number
  group_id: number
  fund_code: string
  fund_name: string
  sort_order: number
}

export interface FundGroup {
  id: number
  name: string
  sort_order: number
  items: FundGroupItem[]
}

export async function getFundGroups(): Promise<FundGroup[]> {
  const res = await client.get('/fund-groups')
  return res.data
}

export async function createFundGroup(name: string): Promise<FundGroup> {
  const res = await client.post('/fund-groups', { name })
  return res.data
}

export async function updateFundGroup(id: number, name: string): Promise<FundGroup> {
  const res = await client.put(`/fund-groups/${id}`, { name })
  return res.data
}

export async function deleteFundGroup(id: number): Promise<void> {
  await client.delete(`/fund-groups/${id}`)
}

export async function reorderFundGroups(ids: number[]): Promise<void> {
  await client.post('/fund-groups/reorder', { ids })
}

export async function createFundGroupItem(groupId: number, item: { fund_code: string; fund_name: string }): Promise<FundGroupItem> {
  const res = await client.post('/fund-group-items', { group_id: groupId, ...item })
  return res.data
}

export async function deleteFundGroupItem(id: number): Promise<void> {
  await client.delete(`/fund-group-items/${id}`)
}

export async function reorderFundGroupItems(ids: number[]): Promise<void> {
  await client.post('/fund-group-items/reorder', { ids })
}
