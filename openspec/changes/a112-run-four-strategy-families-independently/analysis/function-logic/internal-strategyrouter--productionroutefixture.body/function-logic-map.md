# Function Logic Map: `productionRouteFixture.body`

- Source: `internal/strategyrouter/production_test.go` (248-283)
- Function: `productionRouteFixture.body` in package `strategyrouter`
- Signature: `productionRouteFixture.body(params=1, results=1)`
- File SHA-256: `fccd226dcf67215fe7792bb850fbf5ffbddd93c399f72028be3dc6a55400bd38`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Test helper. Builds the signed manifest body used by every production-route test. Candidates are now written with named struct fields rather than positional literals, and the body carries the approved arbitration score version and calibration digest.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- tagged package suite: `go test -count=1 -tags tossos_testseams -covermode=count ./internal/strategyrouter/`
- Measured entry: the function body was not instrumented — `go test` does not add coverage counters to `_test.go` files; it ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/`.

Exact AST return positions: 255:3, 275:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 260:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B2 | else | 267:9 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane` | 262:4 |
| `lane` | 263:4 |
| `lane` | 264:4 |
| `lane` | 265:4 |
| `lane` | 269:4 |
| `lane` | 270:4 |
| `lane` | 271:4 |
| `lane` | 272:4 |
| `string` | 277:56 |
| `Format` | 277:93 |
| `fixture.now.Add` | 277:93 |
| `string` | 278:48 |
| `string` | 278:101 |
| `string` | 279:65 |
| `string` | 280:95 |
| `Format` | 281:40 |
| `fixture.now.Add` | 281:40 |
| `Format` | 282:15 |
| `fixture.now.Add` | 282:15 |

## State mutations and fallbacks

- AST assignments: 7. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Test-only. It writes files under `t.TempDir()` and signs them with a per-test key; it reaches no live endpoint and mints no runtime authority.
