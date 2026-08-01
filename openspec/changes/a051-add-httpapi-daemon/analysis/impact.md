# Impact map: a051-add-httpapi-daemon

## Transport-neutral read graph

| API resource | Existing authority/source | a051 adapter rule |
|---|---|---|
| engine | advisory engine marker and existing console engine-status seam | read-only projection; never start/stop or touch the gate |
| positions | existing holdings reader plus a043 `operatorview.ExitLineView` | reuse the canonical exit-line builder and expose its stable fields |
| orders | existing bounded broker read seam; `exitLine` is reserved/null in a051 | no cancel/amend/place path in the dependency closure; a043 evidence remains on positions |
| candidates | existing persisted signal/candidate readers | no discovery source call and no approval mutation |
| performance | `performance.OpenReadOnly` and `performancejournal.New(*journal.ReadOnly)` | one existing immutable checkpoint; never create/migrate/collect |
| settings | config read services and owner metadata | return current/default/effective state without exposing gate/LIVE writes |
| optimization | a050 category order, registry descriptors, snapshots/history and actor commander | shared descriptors; only explicitly allowlisted non-weakening lifecycle commands may be wired |

## Network and authentication graph

```text
direct TCP peer
  -> shared networkboundary (loopback/VPN CIDR, exact proxy peer, Host/origin)
  -> read-only REST/SSE (no app token)
  -> shared mutation guard supports browser session + CSRF + canonical Origin
     or enrolled mTLS identity; the a051 daemon wiring exposes only the
     Ed25519 one-time 60s capability mode
  -> actor/client + method + resource + canonical-body digest + idempotency key
  -> If-Match CAS -> narrow optimization commander -> append-only audit
```

Forwarded headers are unusable unless startup configuration names the exact TCP
proxy peer. Repeated, comma-chained, partial, opaque, credential-bearing or
non-canonical origin evidence is rejected. No route exists for LIVE, gate,
kill-switch, protection weakening, activation-manifest changes, orders or engine
start/stop.

## Process and deployment graph

- New `tossctl httpapi` process owns its HTTP server, SSE hub and API-owned
  nonce/idempotency state. It does not autostart the trading engine.
- `cmd/tossctl/root.go:newRootCmd` receives one fixed command registration; its
  pre/post hooks and all existing commands stay unchanged.
- `compose.yaml` adds a separately health-checked service published only on the
  configured VPN bind IP. Console/engine lifecycle and rollback remain separate.
- `deploy/container-entrypoint.sh` copies only secrets required by the selected
  command, so the API service is not forced to mount a legacy remote login token.
- Rollback stops only the API service after its API-owned pending command count
  is zero; console and engine stay running.

## High-risk invariants and test branches

1. No-token access reaches GET/HEAD REST and SSE only; mutation is 404/405.
2. Request bodies are capped at 256 KiB before decoding; server header/read
   timeouts are exactly 5 seconds and broader values fail startup validation.
3. SSE uses process epoch plus monotonic sequence, max 32 clients, queue 64,
   15-second heartbeat and slow-client-only disconnect.
4. Journal/performance sources are read-only and immutable; the API dependency
   graph cannot import trading, exec gateway, official writers or console HTML.
5. Optimization category/descriptor JSON and HTML derive from owner registries,
   with no copied default/help constants or arbitrary-input UI.
