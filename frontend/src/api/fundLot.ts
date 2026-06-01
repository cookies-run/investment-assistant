import client from './client'

export interface FundLot {
  id: number
  fund_code: string
  quantity: number
  cost: number
  purchased_at: string
  created_at: string
}

export interface LotSummary {
  total_quantity: number
  fee_free_quantity: number
  fee_reduced_quantity: number
  fee_penalty_quantity: number
  lots: FundLot[]
}

export async function listLots(code: string): Promise<FundLot[]> {
  const res = await client.get(`/funds/${code}/lots`)
  return res.data
}

export async function createLot(code: string, data: Partial<FundLot>) {
  return client.post(`/funds/${code}/lots`, data)
}

export async function deleteLot(code: string, id: number) {
  return client.delete(`/funds/${code}/lots/${id}`)
}
