function toNum(n: number | string | undefined): number | undefined {
  if (n === undefined || n === null) return undefined
  if (typeof n === 'number') return n
  const parsed = parseFloat(n as string)
  return isNaN(parsed) ? undefined : parsed
}

export function formatNumber(n: number | string | undefined, digits = 2): string {
  const num = toNum(n)
  if (num === undefined) return '-'
  return num.toFixed(digits)
}

export function formatPercent(n: number | string | undefined, digits = 2): string {
  const num = toNum(n)
  if (num === undefined) return '-'
  const sign = num > 0 ? '+' : ''
  return `${sign}${num.toFixed(digits)}%`
}

export function formatMoney(n: number | string | undefined): string {
  const num = toNum(n)
  if (num === undefined) return '-'
  return `¥${num.toFixed(2)}`
}

export function formatDateTime(s: string | undefined): string {
  if (!s) return '-'
  const d = new Date(s)
  if (isNaN(d.getTime())) return s
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function formatDate(s: string | undefined): string {
  if (!s) return '-'
  const d = new Date(s)
  if (isNaN(d.getTime())) return s
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}
