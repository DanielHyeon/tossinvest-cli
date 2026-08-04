# Isolated protection lifecycle — Function Logic and Branch Test Map

Date: 2026-08-04
Scope: new dormant `internal/protectionlifecycle` package only

## Hard evidence and edit boundary

CodeGraph evidence at base `d17b90ff819b097908c359943f58665655417b46` identifies the existing
runtime decision and mutation surfaces at:

- `internal/protection/controller.go`: `Controller.Register` (149), `Controller.Replace` (223),
  `Controller.Reconcile` (307), and `Controller.Recover` (385)
- `internal/execgw/protection.go`: `Gateway.checkProtection` (120) and the shipped
  `ProfileProtection=UNWIRED` declaration (68)
- `internal/app/engine/interlock.go`: `protectionReadiness` (273)
- `internal/protectionreadiness/readiness.go`: isolated signed-attestation `Assess` (86)

This wave does not edit, import, or call those functions or `internal/journal`. It adds a sealed, pure lifecycle
model and test-only fake official broker. Therefore there is no pre-edit AST for the new functions and no existing
function logic is changed.

## Planned function logic

### `newState`

1. Validate account and position identity, market, generation, holding quantity, trigger, and other sell claims.
2. Reject control characters, empty identities, unsupported markets, zero generation/quantity/trigger, or claims
   that consume all holdings.
3. Create a market-scoped position with no observed broker protection and seal the complete canonical state.

### `prepareRegister`

1. Validate the state seal and locate the exact account/position/market key.
2. Refuse when the market or position entry latch is closed, protection is active, or another operation is pending.
3. Bound the claim by `holdings - otherSellClaims`.
4. Advance the protection revision exactly once and derive the operation key from length-framed
   `(account, position, market, generation, revision, SUBMIT)`.
5. Durably record desired protection plus `SUBMIT_PENDING` before returning a broker command.

### `applySubmitResult` / `recoverSubmit`

1. Accept only an exact generation/revision/operation-key and full-order observation.
2. Accepted creates one observed ACTIVE order with its exact broker ID and clears pending state.
3. Unknown closes the entry latch and records `SUBMIT_UNKNOWN`; it never emits another mutation.
4. Recovery looks up the exact operation key. ACTIVE converges to the same broker ID.
5. NOT_FOUND permits the same-key retry only when broker idempotency/dedup is attested; otherwise transition to
   `RECONCILE_REQUIRED` with no command. UNKNOWN remains latched with no command.

### `prepareReplace` / `applyReplaceResult` / `recoverReplace`

1. Require an exact ACTIVE observed order and atomic or continuously-covered replace capability.
2. Refuse a lower long-position trigger, excess quantity, or any cancel-then-place shape without changing ACTIVE.
3. Advance revision once; return one exact `REPLACE` command and retain old ACTIVE coverage while pending.
4. Unknown keeps old ACTIVE coverage and closes entry. Exact broker-ID recovery alone may converge the replacement.

### `prepareCancel` / `applyCancelResult` / `recoverCancel`

1. Require exact ACTIVE broker ID and cancel-query capability.
2. Unknown retains ACTIVE coverage and closes entry; cancellation is never inferred.
3. Recovery queries the exact broker ID. ACTIVE retains coverage; CANCELED/FILLED converges terminal state.

### `applyFill`

1. Validate exact broker ID and stable fill ID.
2. Apply a fill ID once. An identical duplicate is a no-op; a conflicting duplicate closes entry and requires
   reconciliation.
3. Refuse quantities exceeding broker remaining claim or holdings.
4. Partial fill immediately decrements holdings and observed/desired quantity and advances lifecycle revision once.
5. Full fill converges terminal protection without delaying common exit/fill processing.

### `discoverOrphan`

1. Compare only durable exact ownership fields; symbol/time heuristics are absent from the model.
2. Exact ownership may be reported for reconciliation.
3. Any unowned or ambiguous broker order is recorded as an orphan, closes only its market entry latch, and returns
   no cancel/adopt/replace command.

## Branch Test Map

| Branch | Expected safety result | Test scenario |
|---|---|---|
| register response lost after broker acceptance | restart exact lookup finds same broker order; submit count stays one | register-response crash matrix |
| unknown submit + NOT_FOUND + idempotency attested | same operation key may be retried | idempotent recovery matrix |
| unknown submit + NOT_FOUND + no idempotency | no resubmit; entry remains latched | no-idempotency matrix |
| cancel unknown | ACTIVE remains protected until exact broker-ID result | cancel-unknown matrix |
| replace unknown | old ACTIVE remains and no retreat occurs | replacement crash matrix |
| lower trigger | refusal; active order unchanged | non-retreat test |
| protection + local sell claims exceed holdings | refusal before mutation | oversubscription test |
| duplicate identical fill | one quantity/revision transition | duplicate-fill test |
| duplicate conflicting fill | reconciliation latch, no guessed mutation | conflicting-fill test |
| unowned orphan | no adopt/cancel/replace | orphan test |
| KR recovery failure | KR only latched; US state byte-equivalent except global reseal | market-isolation test |
| fake transport inspection | no production live hostname/import/toggle/approval | dependency guard |

## Crash-point matrix

The isolated fixture covers process loss before broker acceptance, after acceptance/before response, after response/
before local convergence, cancel response loss, replace response loss, repeated recovery, and repeated fill delivery.
Every restart begins from a previously sealed durable state; recovery accepts only exact operation-key or broker-ID
observations. The fixture never opens a socket and cannot construct production broker authority.

## Post-edit AST and entry truth table

The repository Go AST extractor reports `internal/protectionlifecycle/lifecycle.go` source SHA-256
`de50441bc89c79ec5cfeb8308a837db1cffede7e3ab52c661eccb9515d1688e5` and
`internal/protectionlifecycle/state.go` source SHA-256
`df5c5459c6d2add80bcfdcadb03f4af1dbcb19cce1424687c2d855a40fd66cbe`. The extracted function ranges cover
`prepare/apply/recover` for submit, replace and cancel, `applyFill`, `discoverOrphan`, `validState`, and
`validPositionTruth`.

`validPositionTruth` rejects a sealed state unless phase, broker status and pending operation agree. `EntryOpen`
must be false in UNPROTECTED, every PENDING/UNKNOWN state, RECONCILE_REQUIRED and TERMINAL. It is true only when
phase/status are ACTIVE, no operation is pending, no position or market latch exists, and
`observed broker quantity + other sell claims == holdings`. Cancel recovery that still observes ACTIVE deliberately
keeps a position latch, so it preserves broker coverage without reopening exposure.
