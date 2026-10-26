export class UnauthorizedError extends Error {}

const MESSAGES = {
  'username already exists': 'Username sudah dipakai.',
  'ip already in use': 'IP sudah dipakai user lain.',
  'username and ip are required': 'Username dan IP wajib diisi.',
  'invalid ip address': 'Format IP tidak valid.',
  'invalid netmask': 'Format netmask tidak valid.',
  'user not found': 'User tidak ditemukan.',
  'client not connected': 'Client sudah tidak terkoneksi.',
  'management interface unavailable': 'Management interface OpenVPN tidak bisa diakses.',
  'invalid API key': 'API key salah.',
}

async function request(method, path, body) {
  const res = await fetch(path, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  })

  let data = {}
  try {
    data = await res.json()
  } catch {}

  if (res.status === 401 || res.status === 403) {
    throw new UnauthorizedError(MESSAGES[data.error] || 'Sesi berakhir, silakan login lagi.')
  }
  if (!res.ok || data.success === false) {
    throw new Error(MESSAGES[data.error] || data.error || `Request gagal (${res.status})`)
  }
  return data
}

export const api = {
  login: (apiKey) => request('POST', '/api/login', { api_key: apiKey }),
  logout: () => request('POST', '/api/logout'),
  config: () => request('GET', '/api/config'),
  status: () => request('GET', '/api/status').then((d) => d.clients),
  users: () => request('GET', '/api/users').then((d) => d.users),
  addUser: (user) => request('POST', '/api/users', user),
  updateUser: (username, changes) =>
    request('PUT', `/api/users/${encodeURIComponent(username)}`, changes),
  deleteUser: (username) => request('DELETE', `/api/users/${encodeURIComponent(username)}`),
  kick: (username) => request('POST', `/api/users/${encodeURIComponent(username)}/kick`),
}
