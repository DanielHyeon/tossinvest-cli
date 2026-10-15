# Function Logic Map: `wtsOperations`

- Source: `internal/ops/wts_operations.go`
- AST evidence: `ast.json` (`e622a677b56a55caaf4eec189086219112c9b48ba3d4a49e22f690d00cdf1fd2`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The immutable WTS registry exposes the existing `watchlist` operation. This amendment changes only its ProbeSpec method, URL/query, empty body, and local collection validation; its ID, handler and all other probes stay unchanged.

## Branches and early returns

The amended watchlist check is B61-B65: status, JSON decode, required `watchlists`, per-group iteration, and required `items`. Empty arrays are non-nil and pass. B1-B60 remain unmodified registry/handler branches and are not claimed as tested.

## Calls and live bindings

The check calls `ExpectStatus`, `json.Unmarshal`, `fmt.Errorf`, and `d.WTS.ListWatchlist` remains the unchanged operation handler. `TestWatchlistProbeDedicatedReadContract` obtains the registry spec and calls `ProbeSpec.Check` locally; it performs no HTTP request.

## State mutations and fallbacks

Registry construction and validation are in-memory. No operation handler is invoked; no live probe, schedule, configuration, token, or mutation changes.

## Safety conclusion

Dedicated GET shape and required collection checks now match the read contract while retaining existing status validation. Scope is one ProbeSpec literal.
