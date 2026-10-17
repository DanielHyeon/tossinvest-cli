# Function Logic Map: `productionRouteFixture.body`

- Source: `internal/strategyrouter/production_test.go` (253-288)
- Function: `productionRouteFixture.body` in package `strategyrouter`
- Signature: `productionRouteFixture.body(params=1, results=1)`
- File SHA-256: `21d69cb2a4f5da5ae8d5aded5ea32e5e8fe4aa4d4b5ae0f64038a2bde6cb3d9a`
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

Exact AST return positions: 260:3, 280:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 265:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B2 | else | 272:9 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane` | 267:4 |
| `lane` | 268:4 |
| `lane` | 269:4 |
| `lane` | 270:4 |
| `lane` | 274:4 |
| `lane` | 275:4 |
| `lane` | 276:4 |
| `lane` | 277:4 |
| `string` | 282:56 |
| `Format` | 282:93 |
| `fixture.now.Add` | 282:93 |
| `string` | 283:48 |
| `string` | 283:101 |
| `string` | 284:65 |
| `string` | 285:95 |
| `Format` | 286:40 |
| `fixture.now.Add` | 286:40 |
| `Format` | 287:15 |
| `fixture.now.Add` | 287:15 |

## State mutations and fallbacks

- AST assignments: 7. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Test-only. It writes files under `t.TempDir()` and signs them with a per-test key; it reaches no live endpoint and mints no runtime authority.

a112 게이트 준비(2026-10-04): a127 82080177(전략 권한 적재기가 현재 원장을 읽음)이 같은 파일을 편집해 이 번들이 낡았다 — 분기 · return · 호출 구조 동일
