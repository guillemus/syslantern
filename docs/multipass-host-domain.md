# Multipass host domain

Goal: let the Linux VM call the macOS host with a stable name instead of an IP.

Example URL from inside VM:

```bash
curl http://host.multipass:3000/hello
```

## Setup

Add `host.multipass` to the VM `/etc/hosts`, pointing at the Multipass gateway IP:

```bash
multipass exec linuxbox -- sh -lc 'grep -q " host.multipass$" /etc/hosts && sudo sed -i "s/^.* host\\.multipass$/$(ip route | awk '\''/default/ {print $3}'\'') host.multipass/" /etc/hosts || echo "$(ip route | awk '\''/default/ {print $3}'\'') host.multipass" | sudo tee -a /etc/hosts'
```

Verify:

```bash
multipass exec linuxbox -- getent hosts host.multipass
```

Expected output shape:

```text
192.168.252.1 host.multipass
```

## Use it

Run the server on macOS bound to all interfaces, not only localhost:

```bash
go run ./cmd/server --addr 0.0.0.0:3000
```

Then call it from the VM:

```bash
multipass exec linuxbox -- curl http://host.multipass:3000/hello
```

## Persistence

This survives VM restarts because `/etc/hosts` is stored inside the VM.

It does not survive deleting/recreating the VM.

The IP can become stale if Multipass changes its gateway. If that happens, rerun the setup command.
