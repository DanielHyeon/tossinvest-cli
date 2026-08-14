# Function Logic Map: `Tracker.Observe`

- Source: `internal/reconcile/mismatch.go`
- Post-edit evidence: `ast.json`, `risk-pattern-report.md`
- Frozen base: `3615f793c4a9dbe027fdbe88b3ed01e140b05cc9`

## Inputs and invariants

| Input/state | Valid range | Authority | Failure behavior |
|---|---|---|---|
| `diff` | one authoritative comparison | `Comparer.Compare` | invalid identity still blocks ordinarily but earns no streak |
| `streaks` | current unique exact disputes only | process-local tracker | absent/clean drops evidence; never restored |
| `blocks` | durable plus pending fail-closed projection | journal + memory | write ambiguity retains gate until authority read |
| `adjusted` | symbol credits strictly before comparison | converger | absent/stale credit cannot release |

## Branches and early returns

| Branches | Condition | Mutation/result | Test |
|---|---|---|---|
| B1 | block map absent | initialize | first mismatch |
| B2, B4–B9 | comparison clean; pending authority and exact-cause credited release | fail-closed error, awaiting or proposed release | commit-timeout/read-error and a083 credit tests |
| B3, B10–B13 | blocking; pending earning key disappeared | advance continuity, pre-latch current ordinary blocks/gate, then authority-read | F9 authority-outage test |
| B14–B17 | ordinary current blocks advance/latch exactly once | no duplicate add after successful authority read | ordinary latch tests |
| B18–B19 | deterministic exact key reaches threshold | add pending account permanent with earning evidence | same-key/handoff tests |
| B20–B21 | durable additions confirm | clear pending identity only on account permanent | retry/order tests |
| B22–B23 | authoritative conflict replaces proposal | preserve durable owner | conflict tests |
| B24 | confirmed releases delete exact block | publish release | release suites |
| B25–B28 | credit lifecycle per symbol | spend/refute/preserve bounded credits | a083/a083b suites |
| Return | all paths | recompute permanent, re-sync gate, unlock | full reconcile suite |

## Calls and live bindings

| Callee | Purpose | Contract | Evidence |
|---|---|---|---|
| `promotionKeys`/`advanceStreaks` | exact per-dispute evidence | unique current keys; max compatibility scalar | A110 identity tests |
| `resolvePendingPermanentAuthority` | settle ambiguous write before withdrawal | durable wins; read error fail-closed | commit-timeout tests |
| `pendingBlocksForPersistence`/`persist` | permanent-first deterministic durability | additions pending on error; release needs commit | dual-pending tests |
| `syncGate` | publish conservative projection | before and after journal I/O | gate tests |

## State mutations and fallbacks

- Replaces the former account scalar with a bounded current-dispute map; `failures` is derived maximum only.
- Permanent promotion remains the existing account-wide, durable, operator-only row.
- Ordinary blocks remain immediate and independently retryable even when promotion identity is unclassifiable.
- Pending permanent retries only while its exact earning key is immediately consecutive; authority proves absence before withdrawal.
- A new ordinary mismatch is projected before an authority read that may fail, so later withdrawal of the wider pending proposal cannot reopen that symbol.

## Safety conclusion

- Safe edit boundary: promotion identity/lifecycle plus deterministic persistence order; adjustment release and exit allowance stay unchanged.
- High-risk impact: yes; reconciliation gates automatic adoption and exposure-raising entry.
