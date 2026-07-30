# VPN console threat model

Date: 2026-07-31

## Scope and assets

- Assets: VPN/host access identity, TLS private key, broker session, CSRF,
  configuration toggles, verification approval, audit log.
- In scope: mobile browser → VPN → host/container publish → HTTPS console →
  network-trusted handlers → config/process/verification seams.
- Out of scope: VPN account provisioning, public Internet deployment, multi-user
  identity/RBAC, broker API security.

## Data flow and trust boundaries

```text
[Mobile browser]
       |
       | VPN membership + HTTPS
       v
----- VPN trust boundary ------------------------------------------
[Host VPN IP:port] -> [Docker DNAT (optional)] -> [TLS console]
                                                   |
                           cert/key ----------------+-- config/data/audit mounts
                                                   |
                                                   +-- existing console seams
                                                       (engine/verify/settings)
```

The user explicitly approves loopback or authenticated VPN membership as the sole
application access identity for this single-operator deployment. There is no
application login or session. Network, TLS, Host/Origin, CSRF and action interlocks
remain independent request-safety boundaries.

## Prioritized STRIDE/DREAD findings

| ID | Element | STRIDE | Threat | DREAD (D/R/E/A/D) | Average | Owner and mitigation |
|---|---|---|---|---|---:|---|
| T1 | external/browser | S/E | stolen VPN identity or VPN device gives full console control | 10/9/8/7/8 | 8.4 | user-approved sole identity; VPN admin revokes account/device; host publish remains VPN-IP/loopback only |
| T2 | listener/data flow | I/T | HTTP or bad certificate exposes console actions | 10/9/8/6/8 | 8.2 | console requires TLS 1.3 and cert hostname validation before non-loopback serve |
| T3 | host/container boundary | E/I | wildcard or public port exposes console outside VPN | 10/9/7/6/9 | 8.2 | required explicit bind + allowed CIDR; app peer check; Compose host publish has required VPN IP and no wildcard default |
| T4 | HTTP process | T/E | DNS rebinding, forged Host, cross-origin POST | 9/8/7/6/8 | 7.6 | exact canonical Host; exact Origin/Referer; existing CSRF |
| T5 | secret store/image | I/S | TLS key or broker session enters image/repo/log | 10/8/7/6/8 | 7.8 | 0600/read-only host files, `.dockerignore`, pre-merge secret scan |
| T6 | container process | E/T | compromised console gains root/host capability | 9/7/6/6/7 | 7.0 | non-root UID, read-only root, cap drop ALL, no-new-privileges, no Docker socket/host mounts |
| T7 | request process | D | VPN peer floods console or audit disk | 7/8/7/5/8 | 7.0 | peer-CIDR rejection first, HTTP timeouts/resource limits, bounded existing action endpoints |
| T8 | action/audit | R | operator or attacker denies a remote action | 7/7/5/5/6 | 6.0 | existing setting/action audits remain authoritative; peer identity comes from VPN/host operations |
| T9 | access policy | E | trusted-network is enabled accidentally outside intended deployment | 10/8/7/7/8 | 8.0 | explicit flag, required bind/CIDR/TLS/public URL, mode-conflict rejection, no wildcard host publish |
| T10 | health endpoint | I/E | unauthenticated health leaks state or becomes mutation path | 6/8/6/5/8 | 6.6 | exact GET/HEAD `/healthz`, fixed body only, no dependency calls, static route guard |

## Residual risks

- A device with VPN access can perform every capability the local operator console
  exposes. This is the explicitly approved full-control single-user model, not a
  read-only role.
- Malware on the VPN-connected device can use the console without a second
  application credential. Revoking VPN device/account access is the containment path.
- VPN/firewall correctness is externally operated. Application CIDR checks and
  TLS remain mandatory so a network-policy failure is not the only boundary.

## Verification

- Every DFD element has at least one applicable STRIDE row.
- Every finding with average DREAD ≥7 has a named code/operator/container control.
- RED tests must cover T1–T7, T9 and T10; existing action-audit tests cover T8.
- The senior-security secret scanner must report zero high/critical findings after
  Docker/Compose/env artifacts are added.
