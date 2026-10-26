import { useEffect, useMemo, useRef, useState } from 'preact/hooks'
import { api, UnauthorizedError } from './api.js'
import { Icon, Logo } from './icons.jsx'
import {
  compareIP,
  formatBytes,
  formatClock,
  formatDuration,
  isIPv4,
  maskToPrefix,
  nextFreeIP,
  stripPort,
} from './util.js'

const REFRESH_MS = 10000

function readTheme() {
  try {
    return localStorage.getItem('theme') || 'dark'
  } catch {
    return 'dark'
  }
}

export function App() {
  // null = sedang dicek, false = belum login, true = sudah login
  const [authed, setAuthed] = useState(null)
  const [theme, setTheme] = useState(readTheme)

  useEffect(() => {
    document.documentElement.dataset.theme = theme
    try {
      localStorage.setItem('theme', theme)
    } catch {}
  }, [theme])

  useEffect(() => {
    api.config().then(
      () => setAuthed(true),
      () => setAuthed(false),
    )
  }, [])

  const toggleTheme = () => setTheme(theme === 'dark' ? 'light' : 'dark')

  if (authed === null) return null
  if (!authed) return <Login onLogin={() => setAuthed(true)} theme={theme} onToggleTheme={toggleTheme} />
  return <Dashboard onLogout={() => setAuthed(false)} theme={theme} onToggleTheme={toggleTheme} />
}

function ThemeButton({ theme, onToggle }) {
  const label = theme === 'dark' ? 'Mode terang' : 'Mode gelap'
  return (
    <button class="icon-btn" onClick={onToggle} title={label} aria-label={label}>
      <Icon name={theme === 'dark' ? 'sun' : 'moon'} />
    </button>
  )
}

// ------------------------------------------------------------------ login

function Login({ onLogin, theme, onToggleTheme }) {
  const [key, setKey] = useState('')
  const [show, setShow] = useState(false)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function submit(e) {
    e.preventDefault()
    setBusy(true)
    setError('')
    try {
      await api.login(key)
      onLogin()
    } catch (err) {
      setError(err.message)
      setBusy(false)
    }
  }

  return (
    <main class="login">
      <div class="login-corner">
        <ThemeButton theme={theme} onToggle={onToggleTheme} />
      </div>

      <form class="login-box" onSubmit={submit}>
        <div class="brand">
          <Logo size={22} />
          <strong>degonet</strong>
          <span class="brand-sep">/</span>
          <span>vpn</span>
        </div>

        <div class="login-text">
          <h1>Masuk</h1>
          <p class="muted">Gunakan API key server.</p>
        </div>

        <label class="field">
          <span>API key</span>
          <div class="input-wrap">
            <input
              class="mono"
              type={show ? 'text' : 'password'}
              value={key}
              onInput={(e) => setKey(e.currentTarget.value)}
              autoComplete="current-password"
              spellcheck={false}
              autoFocus
              required
            />
            <button
              type="button"
              class="input-action"
              onClick={() => setShow(!show)}
              aria-label={show ? 'Sembunyikan' : 'Tampilkan'}
              title={show ? 'Sembunyikan' : 'Tampilkan'}
            >
              <Icon name={show ? 'eyeOff' : 'eye'} />
            </button>
          </div>
        </label>

        {error && (
          <p class="form-error">
            <Icon name="alert" size={14} /> {error}
          </p>
        )}

        <button class="primary block" disabled={busy}>
          {busy && <Icon name="loader" class="spin" />}
          {busy ? 'Memverifikasi…' : 'Masuk'}
        </button>

        <p class="footnote">Sesi berlaku 12 jam.</p>
      </form>
    </main>
  )
}

// -------------------------------------------------------------- dashboard

function Dashboard({ onLogout, theme, onToggleTheme }) {
  const [config, setConfig] = useState({ vpn_subnet: '', default_netmask: '255.255.255.0' })
  const [users, setUsers] = useState(null)
  const [clients, setClients] = useState([])
  const [statusError, setStatusError] = useState('')
  const [lastSync, setLastSync] = useState(null)
  const [syncing, setSyncing] = useState(false)
  const [query, setQuery] = useState('')
  const [filter, setFilter] = useState('all')
  const [selected, setSelected] = useState(new Set())
  const [editing, setEditing] = useState(null) // null | { mode: 'add' } | { mode: 'edit', user }
  const [confirm, setConfirm] = useState(null) // null | { title, message, action, label, icon }
  const [toast, setToast] = useState(null)

  const notify = (message, type = 'ok') => {
    setToast({ message, type, id: Date.now() })
  }

  function handleError(err) {
    if (err instanceof UnauthorizedError) onLogout()
    else notify(err.message, 'error')
  }

  async function refresh() {
    setSyncing(true)
    try {
      setUsers(await api.users())
    } catch (err) {
      setSyncing(false)
      handleError(err)
      return
    }
    try {
      setClients(await api.status())
      setStatusError('')
    } catch (err) {
      if (err instanceof UnauthorizedError) return onLogout()
      setClients([])
      setStatusError(err.message)
    }
    setLastSync(new Date())
    setSyncing(false)
  }

  useEffect(() => {
    api.config().then(setConfig, handleError)
    refresh()
    const timer = setInterval(refresh, REFRESH_MS)
    return () => clearInterval(timer)
  }, [])

  useEffect(() => {
    if (!toast) return
    const timer = setTimeout(() => setToast(null), 4000)
    return () => clearTimeout(timer)
  }, [toast])

  const online = useMemo(() => new Map(clients.map((c) => [c.username, c])), [clients])

  const rows = useMemo(() => {
    const q = query.trim().toLowerCase()
    return (users || [])
      .map((u) => ({ ...u, client: online.get(u.username) }))
      .filter((u) => !q || u.username.toLowerCase().includes(q) || u.ip.includes(q))
      .filter((u) => filter === 'all' || (filter === 'online') === Boolean(u.client))
      .sort((a, b) => compareIP(a.ip, b.ip))
  }, [users, online, query, filter])

  const visibleSelected = rows.filter((u) => selected.has(u.username))
  const allSelected = rows.length > 0 && visibleSelected.length === rows.length

  function toggle(username) {
    const next = new Set(selected)
    next.has(username) ? next.delete(username) : next.add(username)
    setSelected(next)
  }

  function toggleAll() {
    setSelected(allSelected ? new Set() : new Set(rows.map((u) => u.username)))
  }

  function askDelete(usernames) {
    setConfirm({
      title: usernames.length === 1 ? 'Hapus user' : `Hapus ${usernames.length} user`,
      message:
        usernames.length === 1
          ? `User "${usernames[0]}" akan dihapus permanen beserta file CCD-nya.`
          : `${usernames.join(', ')} akan dihapus permanen beserta file CCD-nya.`,
      label: 'Hapus',
      action: async () => {
        const failed = []
        for (const name of usernames) {
          try {
            await api.deleteUser(name)
          } catch (err) {
            if (err instanceof UnauthorizedError) return onLogout()
            failed.push(name)
          }
        }
        setSelected(new Set([...selected].filter((n) => !usernames.includes(n) || failed.includes(n))))
        if (failed.length) notify(`Gagal menghapus: ${failed.join(', ')}`, 'error')
        else notify(usernames.length === 1 ? `User ${usernames[0]} dihapus.` : `${usernames.length} user dihapus.`)
        refresh()
      },
    })
  }

  function askKick(username) {
    setConfirm({
      title: 'Putuskan koneksi',
      message: `Koneksi "${username}" akan diputus. Client biasanya akan langsung mencoba konek ulang.`,
      label: 'Putuskan',
      action: async () => {
        try {
          await api.kick(username)
          notify(`Koneksi ${username} diputus.`)
          refresh()
        } catch (err) {
          handleError(err)
        }
      },
    })
  }

  async function logout() {
    try {
      await api.logout()
    } finally {
      onLogout()
    }
  }

  const totalUsers = users ? users.length : 0
  const totalRx = clients.reduce((n, c) => n + c.bytes_received, 0)
  const totalTx = clients.reduce((n, c) => n + c.bytes_sent, 0)
  const onlinePct = totalUsers ? Math.round((clients.length / totalUsers) * 100) : 0
  const noStatus = Boolean(statusError)

  return (
    <div class="shell">
      <header class="topbar">
        <div class="brand">
          <Logo size={20} />
          <strong>degonet</strong>
          <span class="brand-sep">/</span>
          <span>vpn</span>
        </div>

        <div class="topbar-right">
          <span class={`conn ${noStatus ? 'conn-warn' : ''}`} title={statusError || 'Management interface terhubung'}>
            <span class="dot" />
            <span class="hide-sm">{noStatus ? 'management offline' : 'management'}</span>
          </span>
          <span class="sync hide-sm">sync {formatClock(lastSync)}</span>
          <button class="icon-btn" onClick={refresh} title="Muat ulang" aria-label="Muat ulang">
            <Icon name="refresh" class={syncing ? 'spin' : ''} />
          </button>
          <ThemeButton theme={theme} onToggle={onToggleTheme} />
          <button class="icon-btn" onClick={logout} title="Keluar" aria-label="Keluar">
            <Icon name="logout" />
          </button>
        </div>
      </header>

      <main class="content">
        <div class="page-head">
          <h1>Client</h1>
          {config.vpn_subnet && <span class="subnet mono">{config.vpn_subnet}</span>}
        </div>

        <section class="stats">
          <Stat label="User" value={users ? totalUsers : '–'} />
          <Stat
            label="Online"
            value={noStatus ? '–' : clients.length}
            unit={noStatus || !totalUsers ? '' : `/ ${totalUsers}`}
            bar={noStatus ? null : onlinePct}
          />
          <Stat label="Masuk" value={noStatus ? '–' : formatBytes(totalRx)} hint="dari client, sesi aktif" />
          <Stat label="Keluar" value={noStatus ? '–' : formatBytes(totalTx)} hint="ke client, sesi aktif" />
        </section>

        {statusError && (
          <p class="banner">
            <Icon name="alert" /> Status online tidak tersedia: {statusError}
          </p>
        )}

        <section class="panel">
          <div class="toolbar">
            <div class="input-wrap search">
              <Icon name="search" />
              <input
                type="search"
                placeholder="Cari username atau IP…"
                value={query}
                onInput={(e) => setQuery(e.currentTarget.value)}
              />
            </div>
            <div class="segmented" role="group" aria-label="Filter status">
              {[
                ['all', 'Semua', totalUsers],
                ['online', 'Online', clients.length],
                ['offline', 'Offline', Math.max(totalUsers - clients.length, 0)],
              ].map(([value, label, count]) => (
                <button key={value} class={filter === value ? 'active' : ''} onClick={() => setFilter(value)}>
                  {label}
                  <span class="count">{count}</span>
                </button>
              ))}
            </div>
            <div class="spacer" />
            {visibleSelected.length > 0 && (
              <button class="danger" onClick={() => askDelete(visibleSelected.map((u) => u.username))}>
                <Icon name="trash" /> Hapus {visibleSelected.length}
              </button>
            )}
            <button class="primary" onClick={() => setEditing({ mode: 'add' })}>
              <Icon name="plus" /> Tambah user
            </button>
          </div>

          <div class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th class="check">
                    <input type="checkbox" checked={allSelected} onChange={toggleAll} aria-label="Pilih semua" />
                  </th>
                  <th>Client</th>
                  <th class="hide-sm">Status</th>
                  <th class="hide-sm">IP VPN</th>
                  <th class="hide-md">IP publik</th>
                  <th class="hide-md">Uptime</th>
                  <th class="hide-md num">Trafik</th>
                  <th class="actions-head" />
                </tr>
              </thead>
              <tbody>
                {users === null &&
                  [0, 1, 2].map((i) => (
                    <tr key={i} class="skeleton-row">
                      <td colSpan={8}>
                        <span class="skeleton" />
                      </td>
                    </tr>
                  ))}
                {users !== null && rows.length === 0 && (
                  <tr>
                    <td colSpan={8}>
                      <div class="empty">
                        <Icon name="inbox" size={28} />
                        <strong>{users.length === 0 ? 'Belum ada user' : 'Tidak ada yang cocok'}</strong>
                        <span>
                          {users.length === 0
                            ? 'Tambahkan user pertama untuk mulai membuat tunnel.'
                            : 'Coba kata kunci atau filter lain.'}
                        </span>
                      </div>
                    </td>
                  </tr>
                )}
                {rows.map((u) => (
                  <tr key={u.username} class={selected.has(u.username) ? 'selected' : ''}>
                    <td class="check">
                      <input
                        type="checkbox"
                        checked={selected.has(u.username)}
                        onChange={() => toggle(u.username)}
                        aria-label={`Pilih ${u.username}`}
                      />
                    </td>
                    <td>
                      <div class="client">
                        <span class={`dot ${u.client ? 'on' : ''}`} title={u.client ? 'Online' : 'Offline'} />
                        <div>
                          <strong>{u.username}</strong>
                          <span class="sub mono show-sm">
                            {u.ip}/{maskToPrefix(u.netmask)}
                          </span>
                        </div>
                      </div>
                    </td>
                    <td class={`hide-sm status ${u.client ? 'on' : ''}`}>{u.client ? 'online' : 'offline'}</td>
                    <td class="hide-sm mono">
                      {u.ip}
                      <span class="faint">/{maskToPrefix(u.netmask)}</span>
                    </td>
                    <td class="hide-md mono">{u.client ? stripPort(u.client.real_address) : <span class="faint">—</span>}</td>
                    <td class="hide-md mono">
                      {u.client ? formatDuration(u.client.connected_since) : <span class="faint">—</span>}
                    </td>
                    <td class="hide-md num mono">
                      {u.client ? (
                        <span class="traffic">
                          <span>
                            <Icon name="arrowDown" size={12} />
                            {formatBytes(u.client.bytes_received)}
                          </span>
                          <span>
                            <Icon name="arrowUp" size={12} />
                            {formatBytes(u.client.bytes_sent)}
                          </span>
                        </span>
                      ) : (
                        <span class="faint">—</span>
                      )}
                    </td>
                    <td class="actions">
                      <button
                        class="icon-btn sm"
                        onClick={() => setEditing({ mode: 'edit', user: u })}
                        title="Edit"
                        aria-label={`Edit ${u.username}`}
                      >
                        <Icon name="pencil" />
                      </button>
                      <button
                        class="icon-btn sm"
                        onClick={() => askKick(u.username)}
                        disabled={!u.client}
                        title={u.client ? 'Putuskan koneksi' : 'Tidak sedang online'}
                        aria-label={`Putuskan ${u.username}`}
                      >
                        <Icon name="unplug" />
                      </button>
                      <button
                        class="icon-btn sm danger-hover"
                        onClick={() => askDelete([u.username])}
                        title="Hapus"
                        aria-label={`Hapus ${u.username}`}
                      >
                        <Icon name="trash" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {users !== null && users.length > 0 && (
            <div class="table-foot">
              <span>
                {rows.length} dari {users.length} user
              </span>
              <span class="hide-sm">refresh otomatis tiap {REFRESH_MS / 1000} detik</span>
            </div>
          )}
        </section>
      </main>

      {editing && (
        <UserForm
          mode={editing.mode}
          user={editing.user}
          online={editing.user && online.has(editing.user.username)}
          suggestedIP={nextFreeIP(config.vpn_subnet, (users || []).map((u) => u.ip))}
          defaultNetmask={config.default_netmask}
          onClose={() => setEditing(null)}
          onSaved={(message) => {
            setEditing(null)
            notify(message)
            refresh()
          }}
          onUnauthorized={onLogout}
        />
      )}

      {confirm && <Confirm {...confirm} onClose={() => setConfirm(null)} />}

      {toast && (
        <div key={toast.id} class={`toast toast-${toast.type}`} role="status">
          <Icon name={toast.type === 'error' ? 'alert' : 'check'} />
          {toast.message}
        </div>
      )}
    </div>
  )
}

function Stat({ label, value, unit, hint, bar }) {
  return (
    <div class="stat" title={hint}>
      <span class="stat-label">{label}</span>
      <span class="stat-value">
        {value}
        {unit && <span class="stat-unit">{unit}</span>}
      </span>
      {bar != null && (
        <span class="bar">
          <span style={{ width: `${bar}%` }} />
        </span>
      )}
    </div>
  )
}

// ------------------------------------------------------------------ modal

function Modal({ title, subtitle, onClose, children }) {
  const ref = useRef()
  useEffect(() => {
    ref.current.showModal()
  }, [])
  return (
    <dialog
      ref={ref}
      class="modal"
      onClose={onClose}
      onClick={(e) => e.target === ref.current && ref.current.close()}
    >
      <div class="modal-head">
        <div>
          <h2>{title}</h2>
          {subtitle && <p>{subtitle}</p>}
        </div>
        <button class="icon-btn sm modal-close" onClick={() => ref.current.close()} aria-label="Tutup">
          <Icon name="x" />
        </button>
      </div>
      <div class="modal-body">{children}</div>
    </dialog>
  )
}

function UserForm({ mode, user, online, suggestedIP, defaultNetmask, onClose, onSaved, onUnauthorized }) {
  const isEdit = mode === 'edit'
  const [form, setForm] = useState({
    username: user?.username || '',
    ip: user?.ip || suggestedIP,
    netmask: user?.netmask || defaultNetmask,
    password: '',
  })
  const [showPass, setShowPass] = useState(false)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const set = (field) => (e) => setForm({ ...form, [field]: e.currentTarget.value })

  async function submit(e) {
    e.preventDefault()
    if (!isIPv4(form.ip)) return setError('Format IP tidak valid.')
    if (!isIPv4(form.netmask)) return setError('Format netmask tidak valid.')

    setBusy(true)
    setError('')
    try {
      if (isEdit) {
        const changes = {}
        if (form.ip !== user.ip) changes.ip = form.ip
        if (form.netmask !== user.netmask) changes.netmask = form.netmask
        if (form.password) changes.password = form.password
        await api.updateUser(user.username, changes)
        onSaved(`User ${user.username} diperbarui.`)
      } else {
        await api.addUser(form)
        onSaved(`User ${form.username} ditambahkan.`)
      }
    } catch (err) {
      if (err instanceof UnauthorizedError) return onUnauthorized()
      setError(err.message)
      setBusy(false)
    }
  }

  const ipChanged = isEdit && (form.ip !== user.ip || form.netmask !== user.netmask)

  return (
    <Modal
      title={isEdit ? `Edit ${user.username}` : 'Tambah user'}
      subtitle={isEdit ? 'Ubah alokasi IP atau reset password.' : 'Buat kredensial dan alokasi IP tunnel baru.'}
      onClose={onClose}
    >
      <form class="form" onSubmit={submit}>
        {!isEdit && (
          <label class="field">
            <span>Username</span>
            <div class="input-wrap">
              <input class="mono" value={form.username} onInput={set('username')} spellcheck={false} required autoFocus />
            </div>
          </label>
        )}
        <div class="row">
          <label class="field">
            <span>IP VPN</span>
            <div class="input-wrap">
              <input class="mono" value={form.ip} onInput={set('ip')} spellcheck={false} required />
            </div>
          </label>
          <label class="field">
            <span>Netmask</span>
            <div class="input-wrap">
              <span class="input-prefix mono">/{maskToPrefix(form.netmask) || '?'}</span>
              <input class="mono" value={form.netmask} onInput={set('netmask')} spellcheck={false} required />
            </div>
          </label>
        </div>
        {!isEdit && suggestedIP && form.ip === suggestedIP && (
          <p class="hint">IP kosong berikutnya di subnet.</p>
        )}
        <label class="field">
          <span>Password</span>
          <div class="input-wrap">
            <input
              type={showPass ? 'text' : 'password'}
              value={form.password}
              onInput={set('password')}
              placeholder={isEdit ? 'Kosongkan jika tidak diubah' : 'Kosongkan = sama dengan username'}
              autoComplete="new-password"
            />
            <button
              type="button"
              class="input-action"
              onClick={() => setShowPass(!showPass)}
              aria-label={showPass ? 'Sembunyikan' : 'Tampilkan'}
            >
              <Icon name={showPass ? 'eyeOff' : 'eye'} />
            </button>
          </div>
        </label>
        {ipChanged && online && (
          <p class="hint warn">
            <Icon name="alert" size={13} /> User sedang online. IP baru berlaku setelah client konek ulang.
          </p>
        )}
        {error && (
          <p class="form-error">
            <Icon name="alert" size={14} /> {error}
          </p>
        )}
        <div class="buttons">
          <button type="button" class="ghost" onClick={onClose}>
            Batal
          </button>
          <button class="primary" disabled={busy}>
            {busy && <Icon name="loader" class="spin" />}
            {busy ? 'Menyimpan…' : 'Simpan'}
          </button>
        </div>
      </form>
    </Modal>
  )
}

function Confirm({ title, message, label, action, onClose }) {
  const [busy, setBusy] = useState(false)

  async function run() {
    setBusy(true)
    await action()
    onClose()
  }

  return (
    <Modal title={title} onClose={onClose}>
      <p class="confirm-message">{message}</p>
      <div class="buttons">
        <button class="ghost" onClick={onClose}>
          Batal
        </button>
        <button class="danger" onClick={run} disabled={busy}>
          {busy && <Icon name="loader" class="spin" />}
          {busy ? 'Memproses…' : label}
        </button>
      </div>
    </Modal>
  )
}
