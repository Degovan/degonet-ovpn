#!/usr/bin/env bash
set -euo pipefail

SERVICE_NAME="${SERVICE_NAME:-degonet-ovpn}"
SYSTEMD_DIR="${SYSTEMD_DIR:-/etc/systemd/system}"
ENABLE_ON_BOOT="${ENABLE_ON_BOOT:-0}"
START_NOW="${START_NOW:-0}"

PROJECT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_LOCAL_PATH="${PROJECT_DIR}/${SERVICE_NAME}.service"
SERVICE_SYSTEMD_PATH="${SYSTEMD_DIR}/${SERVICE_NAME}.service"

OPENVPN_BIN="${OPENVPN_BIN:-$(command -v openvpn || true)}"
if [[ -z "${OPENVPN_BIN}" ]]; then
  echo "Error: openvpn binary not found in PATH."
  exit 1
fi

cat > "${SERVICE_LOCAL_PATH}" <<EOF
[Unit]
Description=DegoNet OpenVPN Server
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=${PROJECT_DIR}
ExecStartPre=/usr/bin/mkdir -p ${PROJECT_DIR}/logs
ExecStart=${OPENVPN_BIN} --config ${PROJECT_DIR}/server.conf
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

echo "Generated service file: ${SERVICE_LOCAL_PATH}"

if [[ "${EUID}" -eq 0 ]]; then
  install -m 644 "${SERVICE_LOCAL_PATH}" "${SERVICE_SYSTEMD_PATH}"
  systemctl daemon-reload

  if [[ "${ENABLE_ON_BOOT}" == "1" ]]; then
    systemctl enable "${SERVICE_NAME}"
  fi

  if [[ "${START_NOW}" == "1" ]]; then
    systemctl restart "${SERVICE_NAME}"
  fi
else
  sudo install -m 644 "${SERVICE_LOCAL_PATH}" "${SERVICE_SYSTEMD_PATH}"
  sudo systemctl daemon-reload

  if [[ "${ENABLE_ON_BOOT}" == "1" ]]; then
    sudo systemctl enable "${SERVICE_NAME}"
  fi

  if [[ "${START_NOW}" == "1" ]]; then
    sudo systemctl restart "${SERVICE_NAME}"
  fi
fi

echo "Installed service: ${SERVICE_SYSTEMD_PATH}"
echo "Done. Check status with: systemctl status ${SERVICE_NAME}"
