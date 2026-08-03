# Function Logic Map: `Detector.collect`

Source: `internal/filldetect/detect.go`
Function: `Detector.collect`
Signature: `Detector.collect(params=4, results=2)`
Source SHA-256: `5441296826821097f82da79215934616d295c31644f24c8c4126d5778594fb2b`
Revision: `current`

## Inputs and invariants

- Inputs are `Detector.collect(params=4, results=2)` parameters and receiver or package state.
- Canonical identity is account, market, trading day, symbol, side, and opaque broker order id; ownership must be confirmed, unique, and strictly earlier than evidence.
- External, partial, later-owner, or multi-intent evidence remains fail closed and cannot alter projection, P&L, provenance, reservations, or reconcile recovery.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/detect.go:365 | Preserve the condition, error propagation, and fail-closed behavior. |
| B2 | if | internal/filldetect/detect.go:371 | Preserve the condition, error propagation, and fail-closed behavior. |
| B3 | if | internal/filldetect/detect.go:375 | Preserve the condition, error propagation, and fail-closed behavior. |
| B4 | range | internal/filldetect/detect.go:382 | Preserve the condition, error propagation, and fail-closed behavior. |
| B5 | if | internal/filldetect/detect.go:384 | Preserve the condition, error propagation, and fail-closed behavior. |
| B6 | if | internal/filldetect/detect.go:387 | Preserve the condition, error propagation, and fail-closed behavior. |
| B7 | range | internal/filldetect/detect.go:399 | Preserve the condition, error propagation, and fail-closed behavior. |
| B8 | if | internal/filldetect/detect.go:401 | Preserve the condition, error propagation, and fail-closed behavior. |
| B9 | if | internal/filldetect/detect.go:409 | Preserve the condition, error propagation, and fail-closed behavior. |
| B10 | else | internal/filldetect/detect.go:414 | Preserve the condition, error propagation, and fail-closed behavior. |
| B11 | if | internal/filldetect/detect.go:410 | Preserve the condition, error propagation, and fail-closed behavior. |
| B12 | if | internal/filldetect/detect.go:414 | Preserve the condition, error propagation, and fail-closed behavior. |
| B13 | range | internal/filldetect/detect.go:423 | Preserve the condition, error propagation, and fail-closed behavior. |
| B14 | if | internal/filldetect/detect.go:424 | Preserve the condition, error propagation, and fail-closed behavior. |
| B15 | if | internal/filldetect/detect.go:429 | Preserve the condition, error propagation, and fail-closed behavior. |
| B16 | if | internal/filldetect/detect.go:434 | Preserve the condition, error propagation, and fail-closed behavior. |
| B17 | if | internal/filldetect/detect.go:440 | Preserve the condition, error propagation, and fail-closed behavior. |
| B18 | if | internal/filldetect/detect.go:445 | Preserve the condition, error propagation, and fail-closed behavior. |
| B19 | if | internal/filldetect/detect.go:449 | Preserve the condition, error propagation, and fail-closed behavior. |

## Calls and live bindings

- `execgw.ScanOrders`: errors and state follow mapped branches.
- `fmt.Errorf`: errors and state follow mapped branches.
- `len`: errors and state follow mapped branches.
- `d.Tracked.TrackedOrders`: errors and state follow mapped branches.
- `strings.TrimSpace`: errors and state follow mapped branches.
- `d.Tracked.SelectedAccountRef`: errors and state follow mapped branches.
- `errors.New`: errors and state follow mapped branches.
- `make`: errors and state follow mapped branches.
- `trackedOrderKey`: errors and state follow mapped branches.
- `append`: errors and state follow mapped branches.
- `d.brokerVisibleFallback`: errors and state follow mapped branches.
- `parseSnapshot`: errors and state follow mapped branches.
- `snapshotOrderKey`: errors and state follow mapped branches.
- `brokerstate.Derive`: errors and state follow mapped branches.
- `viewOf`: errors and state follow mapped branches.
- `d.Order.OrderRaw`: errors and state follow mapped branches.
- No live broker mutation, HTTP mutation, or operating-toggle authority is added.

## State mutations and fallbacks

- AST assignment points: 33; return points: 12; deferred operations: 0.
- Schema-v19 binding is additive, confirmed, temporal, and unique-intent; runtime readers use exact durable scope.
- Errors propagate without adoption, projection, reservation release, or recovery-success claims.

## Safety conclusion

The current source hash and every AST branch are bound to focused and full/race verification. Canonical temporal ownership prevents reused identifiers or external observations from contaminating local state.
