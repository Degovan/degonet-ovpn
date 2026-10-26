const ipToInt = (ip) => ip.split('.').reduce((n, part) => (n << 8) + Number(part), 0) >>> 0
const intToIp = (n) => [24, 16, 8, 0].map((s) => (n >>> s) & 255).join('.')

export const isIPv4 = (s) =>
  /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/.test(s) && s.split('.').every((p) => Number(p) <= 255)

// IP host terkecil yang belum dipakai di subnet VPN (.1 dipakai server OpenVPN).
export function nextFreeIP(subnet, usedIPs) {
  const [base, bits] = (subnet || '').split('/')
  if (!isIPv4(base || '') || !bits) return ''
  const size = 2 ** (32 - Number(bits))
  const network = ipToInt(base) & ~(size - 1)
  const used = new Set(usedIPs)
  for (let i = 2; i < size - 1; i++) {
    const ip = intToIp(network + i)
    if (!used.has(ip)) return ip
  }
  return ''
}

export const compareIP = (a, b) => ipToInt(a) - ipToInt(b)

export function formatBytes(n) {
  if (!n) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.min(Math.floor(Math.log(n) / Math.log(1024)), units.length - 1)
  return `${(n / 1024 ** i).toFixed(i ? 1 : 0)} ${units[i]}`
}

export function formatDuration(sinceUnix) {
  const s = Math.max(0, Math.floor(Date.now() / 1000 - sinceUnix))
  const d = Math.floor(s / 86400)
  const h = Math.floor((s % 86400) / 3600)
  const m = Math.floor((s % 3600) / 60)
  if (d) return `${d} hari ${h} jam`
  if (h) return `${h} jam ${m} mnt`
  if (m) return `${m} mnt`
  return 'baru saja'
}

// "1.2.3.4:51234" -> "1.2.3.4"
export const stripPort = (addr) => (addr || '').replace(/:\d+$/, '')

// "255.255.255.0" -> 24
export const maskToPrefix = (mask) =>
  isIPv4(mask || '') ? ipToInt(mask).toString(2).replace(/0+$/, '').length : ''

export const formatClock = (date) =>
  date ? date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' }) : '--:--:--'
