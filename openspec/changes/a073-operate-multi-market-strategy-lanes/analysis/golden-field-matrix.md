# Frozen strategy runtime projection matrix

The shared server-owned projection is keyed by exact market order `KR`, `US`. Console, REST and SSE consume
the same Go model and field registry. No adapter computes defaults, readiness or effective state.

| Group | Stable fields | Dormant truth | Unavailable truth |
|---|---|---|---|
| envelope | schemaVersion, generatedAt, markets | both markets present | both envelopes retained |
| market | market, status, error | `UNKNOWN` + `NOT_CONFIGURED` | failed market `UNKNOWN` + typed error only |
| lane | id, version, desired, effective | null/null, `OFF/OFF` | identities unavailable, `OFF/OFF` |
| evidence | id, digest, freshness | null/null/`UNOBSERVED` | null/null/`UNKNOWN` |
| campaign | id, legId | null/null | null/null |
| horizon risk | bucket, policyVersion, status | null/null/`UNKNOWN` | null/null/`UNKNOWN` |
| scheduler/calendar | desired, effective, source, version, freshness | `OFF/OFF`, null/null/`UNOBSERVED` | `OFF/OFF`, null/null/`UNKNOWN` |
| activation | desired, effective, digest, status | `OFF/OFF`, null/`NOT_CONFIGURED` | `OFF/OFF`, null/`UNKNOWN` |
| protection | readiness, refusal | exact `UNWIRED`, typed reason | exact `UNWIRED`, `RUNTIME_UNAVAILABLE` |
| reconciliation | status, refusal | `UNKNOWN`, typed reason | `UNKNOWN`, `RUNTIME_UNAVAILABLE` |
| decision | firstRefusal, observedAt | `NOT_CONFIGURED`, null | typed runtime error, null |

Frozen branch scenarios: simultaneous paired current snapshots; KR current/US unavailable; US current/KR
unavailable; both dormant; missing/duplicate/cross-market key rejection; reconnect with a full snapshot replacing
prior partial state; console/REST/SSE JSON-field parity; GET/HEAD only with no query/body; strict Unix unknown
field/schema/size/auth/method rejection; no preview/apply/order/activation/protection mutation capability.

Performance/accounting, Compose replacement/preimage and rollback branches remain pending in tasks 2.5-2.6,
3.5, 4.2, 4.4-4.5 and 5-6.
