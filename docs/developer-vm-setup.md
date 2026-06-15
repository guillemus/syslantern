# Developer VM setup

This project can run the server on macOS and the CLI inside a Linux VM.

Why this setup:

- Server stays easy to run/debug on Mac.
- CLI runs on real Linux userland and reads Linux system stats.
- CLI sends HTTP requests back to the Mac server.

VM tool: Multipass. It creates a small Ubuntu VM named `linuxbox`.

## Basic VM commands

```bash
multipass list
multipass info linuxbox
multipass shell linuxbox
multipass exec linuxbox -- lscpu
multipass stop linuxbox
multipass start linuxbox
```

## Host domain

Inside the VM, use `host.multipass` to call the Mac host.

Setup:

```bash
multipass exec linuxbox -- sh -lc 'grep -q " host.multipass$" /etc/hosts && sudo sed -i "s/^.* host\\.multipass$/$(ip route | awk '\''/default/ {print $3}'\'') host.multipass/" /etc/hosts || echo "$(ip route | awk '\''/default/ {print $3}'\'') host.multipass" | sudo tee -a /etc/hosts'
```

Verify:

```bash
multipass exec linuxbox -- getent hosts host.multipass
```

Run server on macOS bound to all interfaces:

```bash
go run ./cmd/server --addr 0.0.0.0:3000
```

Call it from the VM:

```bash
multipass exec linuxbox -- curl http://host.multipass:3000/hello
```

This survives VM restarts. Rerun setup if the VM is deleted/recreated or Multipass gateway changes.
