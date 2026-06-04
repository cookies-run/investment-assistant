import client from './client'

export interface User {
  id: number
  nickname: string
  avatar?: string
  created_at: string
}

export async function quickRegister(nickname: string): Promise<{ token: string; user: User }> {
  const res = await client.post('/auth/quick-register', { nickname })
  return res.data
}

export async function getMe(): Promise<User> {
  const res = await client.get('/auth/me')
  return res.data
}
