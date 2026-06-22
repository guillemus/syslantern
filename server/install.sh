#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "Usage: curl -fsSL __SYSLANTERN_HUB_URL__/install.sh | bash -s -- <agent-api-key>" >&2
  exit 1
fi

agent_api_key="$1"
hub_url="__SYSLANTERN_HUB_URL__"

if [ "$(id -u)" != "0" ]; then
  if command -v sudo >/dev/null 2>&1; then
    exec sudo "$0" "$agent_api_key"
  fi
  echo "This script must be run as root." >&2
  exit 1
fi

if ! id -u syslantern >/dev/null 2>&1; then
  useradd --system --home-dir /nonexistent --shell /usr/sbin/nologin syslantern
fi

install -d -o syslantern -g syslantern -m 0700 /etc/syslantern-agent
cat > /etc/syslantern-agent/config.json <<EOF
{
  "hub_url": "${hub_url}",
  "agent_api_key": "${agent_api_key}"
}
EOF
chown syslantern:syslantern /etc/syslantern-agent/config.json
chmod 0600 /etc/syslantern-agent/config.json

systemctl restart syslantern-agent
