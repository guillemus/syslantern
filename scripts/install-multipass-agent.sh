#!/usr/bin/env sh
set -eu

SERVICE_NAME=syslantern-agent
BIN_PATH=/usr/local/bin/syslantern-agent
UNIT_PATH=/etc/systemd/system/${SERVICE_NAME}.service

install -m 0755 /tmp/syslantern-agent "$BIN_PATH"

cat >"$UNIT_PATH" <<'EOF'
[Unit]
Description=SysLantern Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/syslantern-agent
Restart=always
RestartSec=2
User=root

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable "$SERVICE_NAME.service"

if systemctl is-active --quiet "$SERVICE_NAME.service"; then
	systemctl restart "$SERVICE_NAME.service"
else
	systemctl start "$SERVICE_NAME.service"
fi

systemctl status --no-pager "$SERVICE_NAME.service"
