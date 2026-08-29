# Evidence reconciliation: trusted-network revision

| Topic | CodeGraph / current HEAD | Supporting context | Reconciled decision |
|---|---|---|---|
| Read-route application login | Every operational route passes `session0` | Same central gate identified | Bypass `session0` only for local and explicit trusted-network mode |
| State-changing safety | `mutating` separately enforces POST, remote same-origin and CSRF | Origin/CSRF path identified | Do not weaken or reorder `mutating` |
| Remote network boundary | `remote.security` enforces security headers, peer CIDR and Host | Remote middleware identified | Preserve before handler execution |
| CLI configuration | `remoteAccessOptions` currently makes token file mandatory | Compose supplies that token | Add explicit trusted-network choice; reject implicit/conflicting mode |
| LIVE/account authority | Engine/verify seams are downstream of existing interlocks and approvals | No auth-specific authority expansion found | No changes outside console access assembly |

There is no unresolved evidence conflict. The user-approved requirement changes
application identity only; current HEAD remains authoritative for all safety seams.
