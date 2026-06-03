# Admin NodePort allowlist summary

## Purpose

This checkpoint reduces the public admin surface while keeping stable lab URLs.

The user/demo entry remains available through:

- `ecommerce-dashboard:30517`

Admin tools remain reachable through fixed Tailscale URLs, but only from the allowlisted admin device.

## Protected admin NodePorts

- ArgoCD HTTP: `30080`
- ArgoCD HTTPS: `30443`
- Grafana: `31000`
- Chaos Dashboard: `31333`
- Chaos metrics: `30930`
- Kafdrop: `30090`

## Allowlist

- Allowed source: `DoneGitOps / 100.126.121.19/32`
- Target node: `vm1-gateway / 100.65.255.2`
- Interface: `tailscale0`
- iptables table: `raw`
- chain: `DACN_ADMIN_NODEPORTS`

## Runtime behavior

- `30517` remains reachable for demo access.
- Admin ports are reachable from `DoneGitOps`.
- Admin ports are blocked from non-allowlisted `vm3-gitops`.
- Old direct `web-gateway:32193` remains closed.

## Persistence

The rule is persisted on `vm1-gateway` by:

- `/usr/local/sbin/dacn-admin-nodeports-allowlist.sh`
- `/etc/systemd/system/dacn-admin-nodeports-allowlist.service`

## Evidence

- `docs/security/runs/admin-nodeport-allowlist-20260603191356/runtime-state.txt`
- `docs/security/runs/admin-nodeport-allowlist-20260603191356/vm1-firewall-rules.txt`
- `docs/security/runs/admin-nodeport-allowlist-20260603191356/allowlisted-donegitops-access.txt`
- `docs/security/runs/admin-nodeport-allowlist-20260603191356/non-allowlisted-vm3-negative.txt`
