# Proposal-freeze review

- Date: 2026-07-31
- Scope: proposal, design, operator-console delta, engine-safety delta
- Voices: Manager self-review + adversarial Eng self-review
- User premise confirmation: the user explicitly requested menu-controlled LIVE
  order/engine execution and boot deployment in this thread.

## Manager review

The product gap is narrow and real: LIVE policy, gate, and manual engine controls
already exist, but lifecycle approval is not persistent. A separate default-OFF
autostart key is preferable to silently redefining the existing gate because the
machine already has the gate ON. It also gives the user a visible answer to
“what will happen after reboot?”

Accepted scope: config key, one settings form, immediate reuse of the existing
start seam, one startup decision, audit, Compose data mapping, tests and deployment.
Deferred: crash-loop supervisor, VPN installation, public/LAN exposure, and any new
order API.

## Adversarial Eng review

| Finding | Severity | Decision |
|---|---|---|
| Reusing `automation_gate.enabled` for boot would start this currently armed machine during deployment | critical | Rejected. Separate `engine.autostart`, default/missing false. |
| A second process runner or systemd engine unit could diverge from the console binary/config/journal | high | Rejected. Reuse `startEngine` inside the console process/container. |
| ON save could bypass startup interlock by invoking a lower-level engine API | critical | Prevented by seam type: it calls only the existing `StartEngine`, which spawns `engine run`. |
| OFF might be misread as an emergency stop | high | OFF changes only next-start behavior; UI points to the existing graceful stop button. |
| Startup config read error could accidentally default ON | critical | Fail closed: no start call, visible initial engine note. |
| Docker XDG data mapping currently nests `tossos/tossos` and would detach the console/engine from the existing journal | critical | Add explicit `TOSSOS_DATA_DIR=/var/lib/tossos/data` before deployment. |
| Remote deployment has no active VPN interface today | high | Deploy loopback HTTPS only; do not substitute LAN or wildcard binding. |

## Decision audit

| Decision | Classification | Rationale |
|---|---|---|
| Separate autostart from automation gate | safety/mechanical | preserves current runtime state and makes lifecycle approval explicit |
| Start immediately after ON save | user-requested behavior | the authenticated CSRF POST is the human act; existing interlock remains final |
| Do not stop on OFF | lifecycle separation | prevents a settings checkbox from becoming an undocumented kill path |
| Keep console available when autostart fails | availability | engine refusal must be inspectable from the console |
| Use the console container restart policy for boot | architecture | no Docker socket, host PID namespace, or competing journal writer |

Status: approved for RED tests. A post-implementation independent code/security
review remains required before deployment.

## Post-implementation code/security review

- Date: 2026-07-31
- Method: deterministic Go quality review, manual authority/data-flow review,
  race tests, `go vet`, hardened-container smoke deployment

| Finding | Severity | Resolution |
|---|---|---|
| Compose secret short syntax mounted files below `/run/secrets`, while the process read `/run/tossos` | high | Resolved with explicit absolute secret targets; smoke image loaded the real certificate/token/session paths. |
| Existing repository `.env` held API credentials on a Windows mount where mode 0600 could not be enforced | high | Resolved by moving the original file intact to Linux storage at `~/.local/share/tossos-deploy/secrets/legacy-api.env` with mode 0600; repository `.env` now contains no credentials. |
| Docker's default 10-second stop grace is shorter than `engineStopTimeout` (60 seconds) | high | Resolved with a 75-second Compose stop grace and SIGTERM handling in the console. |
| `runConsole` remains a large assembly function | medium | Accepted existing architectural debt for this change; new lifecycle policy is isolated in `engineautostart.go` and closed seams. No high-risk branch was left untested. |

Result: no unresolved critical/high finding. Deployment proceeded with
`engine.autostart=false`; no LIVE order mutation or engine process was invoked.
