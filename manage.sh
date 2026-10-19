#!/usr/bin/env bash
# Pengelola DegoNet OpenVPN (Docker): set API secret, compose, user VPN, dll.
# Jalankan tanpa argumen untuk menu interaktif, atau: ./manage.sh <aksi>
set -uo pipefail

PROJECT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${PROJECT_DIR}/.env"
CONTAINER="degonet-openvpn"
NETWORK="degonet-network"
NETWORK_SUBNET="10.10.0.0/24"
AUTH_BIN="/etc/openvpn/auth"

cd "${PROJECT_DIR}" || exit 1

if [[ -t 1 ]]; then
  C_RESET=$'\033[0m' C_BOLD=$'\033[1m' C_DIM=$'\033[2m'
  C_RED=$'\033[31m' C_GREEN=$'\033[32m' C_YELLOW=$'\033[33m' C_CYAN=$'\033[36m'
else
  C_RESET="" C_BOLD="" C_DIM="" C_RED="" C_GREEN="" C_YELLOW="" C_CYAN=""
fi

info()    { printf '%s›%s %s\n' "${C_CYAN}" "${C_RESET}" "$*"; }
success() { printf '%s✓ %s%s\n' "${C_GREEN}" "$*" "${C_RESET}"; }
warn()    { printf '%s⚠ %s%s\n' "${C_YELLOW}" "$*" "${C_RESET}"; }
error()   { printf '%s✗ %s%s\n' "${C_RED}" "$*" "${C_RESET}" >&2; }
header()  { printf '\n%s%s%s\n' "${C_BOLD}" "$*" "${C_RESET}"; }

# confirm "Pertanyaan" [y|n] -> default jawaban kalau user langsung Enter.
confirm() {
  local prompt="$1" default="${2:-n}" hint answer
  [[ "${default}" == "y" ]] && hint="[Y/n]" || hint="[y/N]"
  read -r -p "${prompt} ${hint} " answer || return 1
  answer="${answer:-${default}}"
  [[ "${answer,,}" == "y" || "${answer,,}" == "ya" || "${answer,,}" == "yes" ]]
}

pause() {
  [[ -n "${MENU_MODE:-}" ]] || return 0
  echo
  read -r -p "${C_DIM}Tekan Enter untuk kembali ke menu...${C_RESET}" _ || true
}

# ---------------------------------------------------------------- docker ---

COMPOSE=()

require_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    error "Perintah 'docker' tidak ditemukan."
    return 1
  fi
  if ! docker info >/dev/null 2>&1; then
    error "Docker daemon tidak bisa diakses (belum jalan, atau user ini butuh sudo / grup docker)."
    return 1
  fi
  if [[ ${#COMPOSE[@]} -eq 0 ]]; then
    if docker compose version >/dev/null 2>&1; then
      COMPOSE=(docker compose)
    elif command -v docker-compose >/dev/null 2>&1; then
      COMPOSE=(docker-compose)
    else
      error "Docker Compose tidak ditemukan (butuh 'docker compose' atau 'docker-compose')."
      return 1
    fi
  fi
}

compose() { "${COMPOSE[@]}" "$@"; }

container_state() {
  docker inspect -f '{{.State.Status}}' "${CONTAINER}" 2>/dev/null || echo "tidak ada"
}

container_running() { [[ "$(container_state)" == "running" ]]; }

require_running() {
  if ! container_running; then
    error "Container ${CONTAINER} tidak berjalan. Jalankan dulu (menu Start / ./manage.sh up)."
    return 1
  fi
}

ensure_network() {
  if docker network inspect "${NETWORK}" >/dev/null 2>&1; then
    return 0
  fi
  warn "Network '${NETWORK}' belum ada (dibutuhkan docker-compose.yml sebagai external network)."
  if confirm "Buat network '${NETWORK}' dengan subnet ${NETWORK_SUBNET}?" y; then
    docker network create --subnet "${NETWORK_SUBNET}" "${NETWORK}" >/dev/null &&
      success "Network '${NETWORK}' dibuat."
  else
    error "Dibatalkan: compose tidak bisa jalan tanpa network '${NETWORK}'."
    return 1
  fi
}

# ------------------------------------------------------------------ .env ---

env_get() {
  [[ -f "${ENV_FILE}" ]] || return 0
  grep -E "^${1}=" "${ENV_FILE}" | tail -n 1 | cut -d= -f2-
}

env_set() {
  local key="$1" value="$2" tmp
  tmp="$(mktemp)"
  if [[ -f "${ENV_FILE}" ]]; then
    grep -vE "^${key}=" "${ENV_FILE}" >"${tmp}" || true
  fi
  printf '%s=%s\n' "${key}" "${value}" >>"${tmp}"
  mv "${tmp}" "${ENV_FILE}"
  chmod 600 "${ENV_FILE}"
}

mask() {
  local v="$1"
  if (( ${#v} <= 8 )); then
    printf '%s' "****"
  else
    printf '%s…%s' "${v:0:4}" "${v: -4}"
  fi
}

generate_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
  else
    od -An -N32 -tx1 /dev/urandom | tr -d ' \n'
  fi
}

api_key_status() {
  local key
  key="$(env_get API_KEY)"
  if [[ -z "${key}" ]]; then
    printf '%sBELUM DISET%s (container akan memakai "changeme")' "${C_RED}" "${C_RESET}"
  elif [[ "${key}" == "changeme" ]]; then
    printf '%schangeme%s (ganti!)' "${C_RED}" "${C_RESET}"
  else
    printf '%sterset%s (%s)' "${C_GREEN}" "${C_RESET}" "$(mask "${key}")"
  fi
}

# --------------------------------------------------------------- actions ---

cmd_status() {
  header "Status"
  printf '  API secret : %s\n' "$(api_key_status)"
  require_docker || return 1
  printf '  Container  : %s\n' "$(container_state)"
  if docker network inspect "${NETWORK}" >/dev/null 2>&1; then
    printf '  Network    : %s ada\n' "${NETWORK}"
  else
    printf '  Network    : %s%s belum ada%s\n' "${C_YELLOW}" "${NETWORK}" "${C_RESET}"
  fi
  echo
  compose ps
}

# Container harus dibuat ulang (bukan sekadar restart) agar env baru terbaca.
apply_env_change() {
  require_docker || return 1
  if [[ "$(container_state)" == "tidak ada" ]]; then
    info "Container belum dibuat; secret akan dipakai saat Start/Rebuild."
    return 0
  fi
  if confirm "Terapkan sekarang (container dibuat ulang)?" y; then
    compose up -d && success "Container dibuat ulang dengan secret baru."
  else
    warn "Belum diterapkan. Jalankan './manage.sh up' nanti (restart biasa tidak cukup)."
  fi
}

cmd_secret_set() {
  header "Set API secret"
  printf '  Sekarang: %s\n\n' "$(api_key_status)"
  echo "  1) Generate acak (disarankan)"
  echo "  2) Masukkan manual"
  echo "  0) Batal"
  local choice key key2
  read -r -p "Pilih: " choice || return 1
  case "${choice}" in
    1)
      key="$(generate_secret)"
      ;;
    2)
      read -r -s -p "API secret: " key || return 1
      echo
      read -r -s -p "Ulangi API secret: " key2 || return 1
      echo
      if [[ "${key}" != "${key2}" ]]; then
        error "Secret tidak sama."
        return 1
      fi
      if [[ -z "${key}" || "${key}" =~ [[:space:]] ]]; then
        error "Secret tidak boleh kosong atau berisi spasi."
        return 1
      fi
      if (( ${#key} < 16 )); then
        warn "Secret pendek (${#key} karakter); disarankan minimal 32."
        confirm "Tetap pakai?" n || return 1
      fi
      ;;
    *)
      info "Dibatalkan."
      return 0
      ;;
  esac

  env_set API_KEY "${key}"
  success "API_KEY disimpan ke ${ENV_FILE} (chmod 600)."
  if [[ "${choice}" == "1" ]]; then
    echo
    printf '  %sAPI secret baru:%s %s\n' "${C_BOLD}" "${C_RESET}" "${key}"
    printf '  %sSimpan di client yang memanggil API (header X-API-Key).%s\n\n' "${C_DIM}" "${C_RESET}"
  fi
  apply_env_change
}

cmd_secret_show() {
  local key
  key="$(env_get API_KEY)"
  if [[ -z "${key}" ]]; then
    warn "API_KEY belum diset di ${ENV_FILE}."
    return 1
  fi
  printf '  API secret: %s\n' "$(mask "${key}")"
  if [[ -t 0 ]] && confirm "Tampilkan lengkap?" n; then
    printf '  %s\n' "${key}"
  fi
}

check_secret_before_start() {
  local key
  key="$(env_get API_KEY)"
  if [[ -z "${key}" || "${key}" == "changeme" ]] && [[ -z "${API_KEY:-}" ]]; then
    warn "API secret belum diset; API akan memakai 'changeme'."
    if [[ -t 0 ]] && confirm "Generate secret acak sekarang?" y; then
      key="$(generate_secret)"
      env_set API_KEY "${key}"
      success "API_KEY disimpan ke ${ENV_FILE}."
      printf '  %sAPI secret baru:%s %s\n\n' "${C_BOLD}" "${C_RESET}" "${key}"
    fi
  fi
}

cmd_up() {
  require_docker || return 1
  ensure_network || return 1
  check_secret_before_start
  info "Menjalankan container..."
  compose up -d && success "Container berjalan."
}

cmd_rebuild() {
  require_docker || return 1
  ensure_network || return 1
  check_secret_before_start
  local args=(build)
  if [[ "${1:-}" == "--no-cache" ]] || { [[ -n "${MENU_MODE:-}" ]] && confirm "Build tanpa cache (lebih lama)?" n; }; then
    args+=(--no-cache)
  fi
  info "Build image..."
  compose "${args[@]}" || return 1
  compose up -d && success "Image di-build ulang dan container dijalankan."
}

cmd_restart() {
  require_docker || return 1
  info "Restart container..."
  compose restart && success "Container di-restart."
}

cmd_stop() {
  require_docker || return 1
  compose stop && success "Container dihentikan."
}

cmd_down() {
  require_docker || return 1
  info "Data user & CCD tersimpan di volume dan TIDAK ikut terhapus."
  if [[ -t 0 ]] && ! confirm "Hapus container (down)?" n; then
    info "Dibatalkan."
    return 0
  fi
  compose down && success "Container dihapus."
}

cmd_logs() {
  require_docker || return 1
  info "Menampilkan log (Ctrl+C untuk berhenti)..."
  # Ctrl+C cukup menghentikan log, bukan script ini.
  trap ':' INT
  compose logs -f --tail=100
  trap - INT
}

cmd_health() {
  require_docker || return 1
  require_running || return 1
  local body
  # Pakai API_KEY dari env container, jadi ini juga mengecek secret yang benar-benar aktif.
  body="$(docker exec "${CONTAINER}" sh -c \
    'curl -s -w "\nHTTP %{http_code}" -H "X-API-Key: ${API_KEY}" "http://127.0.0.1:${API_PORT:-80}/api/health"' 2>&1)"
  echo "${body}"
  if [[ "${body}" == *"HTTP 200"* ]]; then
    success "API sehat dan secret di container cocok."
  else
    error "Health check gagal."
    return 1
  fi
}

cmd_shell() {
  require_docker || return 1
  require_running || return 1
  docker exec -it "${CONTAINER}" bash
}

auth_exec() {
  require_docker || return 1
  require_running || return 1
  local flags=(-i)
  [[ -t 0 && -t 1 ]] && flags=(-it)
  docker exec "${flags[@]}" "${CONTAINER}" "${AUTH_BIN}" "$@"
}

cmd_user_list() { auth_exec list; }

cmd_user_add() {
  if [[ $# -gt 0 ]]; then
    auth_exec add "$@"
    return
  fi
  header "Tambah user VPN"
  local username ip password netmask
  read -r -p "Username: " username || return 1
  read -r -p "IP (contoh 10.8.0.10): " ip || return 1
  read -r -s -p "Password (kosong = sama dengan username): " password || return 1
  echo
  read -r -p "Netmask (kosong = default): " netmask || return 1
  if [[ -z "${username}" || -z "${ip}" ]]; then
    error "Username dan IP wajib diisi."
    return 1
  fi
  local args=(add "${username}" "${ip}")
  if [[ -n "${password}" || -n "${netmask}" ]]; then
    args+=("${password:-${username}}")
  fi
  [[ -n "${netmask}" ]] && args+=("${netmask}")
  auth_exec "${args[@]}"
}

# Tanpa argumen: checklist interaktif untuk memilih banyak user.
cmd_user_delete() { auth_exec delete "$@"; }

# ------------------------------------------------------------------- ui ---

show_help() {
  cat <<EOF
Pemakaian: ./manage.sh [aksi]

Tanpa aksi akan membuka menu interaktif.

Container:
  status              Status container, network, dan API secret
  up                  Jalankan container (buat network kalau belum ada)
  rebuild [--no-cache]  Build ulang image lalu jalankan
  restart             Restart container
  stop                Hentikan container
  down                Hapus container (volume data tetap aman)
  logs                Ikuti log container
  shell               Masuk shell di dalam container

Konfigurasi:
  secret              Set API secret (generate / manual) ke .env
  secret:show         Lihat API secret (disamarkan)
  health              Cek GET /api/health dengan secret aktif

User VPN:
  users               Daftar user
  user:add [username ip [password] [netmask]]
  user:delete [username...] [-y]   Tanpa username: pilih banyak secara interaktif
EOF
}

menu() {
  MENU_MODE=1
  local choice state
  while true; do
    state="$( { docker inspect -f '{{.State.Status}}' "${CONTAINER}" 2>/dev/null || echo "tidak ada"; } )"
    clear 2>/dev/null || true
    printf '%sDegoNet OpenVPN Manager%s\n' "${C_BOLD}" "${C_RESET}"
    printf '%sContainer:%s %s   %sAPI secret:%s %s\n' "${C_DIM}" "${C_RESET}" "${state}" "${C_DIM}" "${C_RESET}" "$(api_key_status)"
    cat <<EOF

 ${C_CYAN}Container${C_RESET}
   1) Status            5) Stop
   2) Start / up        6) Down (hapus container)
   3) Restart           7) Logs
   4) Rebuild & start   8) Shell ke container

 ${C_CYAN}Konfigurasi${C_RESET}
   9) Set API secret   10) Lihat API secret   11) Cek health API

 ${C_CYAN}User VPN${C_RESET}
  12) Daftar user      13) Tambah user        14) Hapus user (bisa banyak)

   0) Keluar
EOF
    echo
    read -r -p "Pilih aksi: " choice || { echo; exit 0; }
    echo
    case "${choice}" in
      1) cmd_status ;;
      2) cmd_up ;;
      3) cmd_restart ;;
      4) cmd_rebuild ;;
      5) cmd_stop ;;
      6) cmd_down ;;
      7) cmd_logs ;;
      8) cmd_shell ;;
      9) cmd_secret_set ;;
      10) cmd_secret_show ;;
      11) cmd_health ;;
      12) cmd_user_list ;;
      13) cmd_user_add ;;
      14) cmd_user_delete ;;
      0 | q | exit) exit 0 ;;
      *) warn "Pilihan tidak dikenal: ${choice}" ;;
    esac
    pause
  done
}

main() {
  local action="${1:-}"
  [[ $# -gt 0 ]] && shift
  case "${action}" in
    "") menu ;;
    status) cmd_status ;;
    up | start) cmd_up ;;
    rebuild) cmd_rebuild "$@" ;;
    restart) cmd_restart ;;
    stop) cmd_stop ;;
    down) cmd_down ;;
    logs) cmd_logs ;;
    shell) cmd_shell ;;
    secret) cmd_secret_set ;;
    secret:show) cmd_secret_show ;;
    health) cmd_health ;;
    users | user:list) cmd_user_list ;;
    user:add) cmd_user_add "$@" ;;
    user:delete) cmd_user_delete "$@" ;;
    -h | --help | help) show_help ;;
    *)
      error "Aksi tidak dikenal: ${action}"
      show_help
      exit 1
      ;;
  esac
}

main "$@"
