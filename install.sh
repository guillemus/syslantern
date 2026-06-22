#!/usr/bin/env bash
set -euo pipefail

log() {
  echo "==> $*"
}

# The hub URL is intentionally filled by the SysLantern hub when it serves
# /install.sh. Self-hosted hubs do not have a static canonical URL, so the hub
# replaces this placeholder with the URL from the current request.
hub_url="__SYSLANTERN_HUB_URL__"

# Optional development override for the agent archive URL. Production installs
# use the latest GitHub release by default. Local dev can point this at the hub's
# /public/syslantern-agent.tar.gz build without changing the script.
syslantern_agent_url="${SYSLANTERN_AGENT_URL:-}"

if [ "$#" -ne 1 ]; then
  echo "Usage: sudo ./install.sh <agent-api-key>" >&2
  exit 1
fi

if [ "$(id -u)" != "0" ]; then
  echo "This script must be run as root." >&2
  exit 1
fi

agent_api_key="$1"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
esac

default_agent_url="https://github.com/guillemus/syslantern/releases/latest/download/syslantern_linux_${arch}.tar.gz"
agent_url="${syslantern_agent_url:-$default_agent_url}"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

log "Installing SysLantern agent"
log "Downloading agent for linux/${arch}"
curl -fsSL "$agent_url" -o "$tmp_dir/syslantern.tar.gz"

log "Unpacking agent"
tar --warning=no-timestamp -xzf "$tmp_dir/syslantern.tar.gz" -C "$tmp_dir"

log "Installing binary to /usr/local/bin/syslantern"
install -m 0755 "$tmp_dir/syslantern" /usr/local/bin/syslantern

if ! id -u syslantern >/dev/null 2>&1; then
  log "Creating syslantern system user"
  useradd --system --home-dir /nonexistent --shell /usr/sbin/nologin syslantern
fi

log "Writing /etc/syslantern-agent/config.json"
install -d -o syslantern -g syslantern -m 0700 /etc/syslantern-agent
cat > /etc/syslantern-agent/config.json <<EOF
{
  "hub_url": "${hub_url}",
  "agent_api_key": "${agent_api_key}"
}
EOF
chown syslantern:syslantern /etc/syslantern-agent/config.json
chmod 0600 /etc/syslantern-agent/config.json

log "Writing systemd service"
cat > /etc/systemd/system/syslantern-agent.service <<EOF
[Unit]
Description=SysLantern Agent
After=network-online.target
Wants=network-online.target

[Service]
User=syslantern
Group=syslantern
ExecStart=/usr/local/bin/syslantern
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

log "Starting syslantern-agent.service"
systemctl daemon-reload
systemctl enable syslantern-agent
systemctl restart syslantern-agent

log "SysLantern agent installed"
log "Open your SysLantern hub to see this agent: ${hub_url}"
systemctl --no-pager --full status syslantern-agent || true
