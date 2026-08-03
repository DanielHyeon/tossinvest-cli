# Function Logic Map: `Detector.collect`

Source: `internal/filldetect/detect.go`  
Function: `Detector.collect`  
Signature: `Detector.collect(params=4, results=2)`  
Source SHA-256: `5441296826821097f82da79215934616d295c31644f24c8c4126d5778594fb2b`

## Inputs and invariants

- Inputs are the parameters in `Detector.collect(params=4, results=2)` and receiver state.
- Canonical order identity is account/market/trading-day/symbol/side/opaque-order-id; no layer may collapse it back to order-id alone.
- External broker evidence cannot become engine ownership, and ambiguity remains durable and fail closed.

## Branches and early returns

| ID | Kind | Location | Contract |
| --- | --- | --- | --- |
| B1 | if | internal/filldetect/detect.go:365 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B2 | if | internal/filldetect/detect.go:371 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B3 | if | internal/filldetect/detect.go:375 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B4 | range | internal/filldetect/detect.go:382 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B5 | if | internal/filldetect/detect.go:384 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B6 | if | internal/filldetect/detect.go:387 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B7 | range | internal/filldetect/detect.go:399 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B8 | if | internal/filldetect/detect.go:401 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B9 | if | internal/filldetect/detect.go:409 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B10 | else | internal/filldetect/detect.go:414 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B11 | if | internal/filldetect/detect.go:410 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B12 | if | internal/filldetect/detect.go:414 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B13 | range | internal/filldetect/detect.go:423 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B14 | if | internal/filldetect/detect.go:424 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B15 | if | internal/filldetect/detect.go:429 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B16 | if | internal/filldetect/detect.go:434 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B17 | if | internal/filldetect/detect.go:440 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B18 | if | internal/filldetect/detect.go:445 | Preserve the explicit condition, early return, and fail-closed error behavior. |
| B19 | if | internal/filldetect/detect.go:449 | Preserve the explicit condition, early return, and fail-closed error behavior. |

## Calls and live bindings

- `execgw.ScanOrders`: returned errors and state follow the mapped branches.
- `fmt.Errorf`: returned errors and state follow the mapped branches.
- `len`: returned errors and state follow the mapped branches.
- `d.Tracked.TrackedOrders`: returned errors and state follow the mapped branches.
- `strings.TrimSpace`: returned errors and state follow the mapped branches.
- `d.Tracked.SelectedAccountRef`: returned errors and state follow the mapped branches.
- `errors.New`: returned errors and state follow the mapped branches.
- `make`: returned errors and state follow the mapped branches.
- `trackedOrderKey`: returned errors and state follow the mapped branches.
- `append`: returned errors and state follow the mapped branches.
- `d.brokerVisibleFallback`: returned errors and state follow the mapped branches.
- `parseSnapshot`: returned errors and state follow the mapped branches.
- `snapshotOrderKey`: returned errors and state follow the mapped branches.
- `brokerstate.Derive`: returned errors and state follow the mapped branches.
- `viewOf`: returned errors and state follow the mapped branches.
- `d.Order.OrderRaw`: returned errors and state follow the mapped branches.
- Official reads and runtime config stay caller-bound; no live broker mutation or operating-toggle authority is added.

## State mutations and fallbacks

- The AST contains 33 assignment point(s); durable writes precede release visibility.
- Scoped v17 evidence is authoritative. Legacy empty-scope evidence is accepted only when uniquely attributable and never as a reuse wildcard.
- Errors propagate without projection, adoption, reservation release, or recovery success claims.

## Safety conclusion

Every AST branch is bound to this source hash and mapped to focused plus full/race verification. Composite snapshot identity, canonical detector matching, and bidirectional reconciliation matching prevent a reused opaque identifier from producing a false-clean recovery.
