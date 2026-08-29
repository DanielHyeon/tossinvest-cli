# Function Logic Map: `ownerSnapshotFault`

- Source: `internal/strategyarbiter/arbiter.go` (194-212)
- Function: `ownerSnapshotFault` in package `strategyarbiter`
- Signature: `ownerSnapshotFault(params=1, results=2)`
- File SHA-256: `1788b0503479c4c4e8b4b17d7e2e6e2fd189f414ebde7877a7d5043157be6d03`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 5.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

자격 집합이 서지 못했을 때 그 이유를 소유자 스냅샷에서 좁힌다.
`RouteSet` 은 묵은 리비전과 소유자 중복을 `RefusalReconstructionMismatch` 하나로 뭉쳐서 돌려주는데,
동결 골든은 그 둘에 서로 다른 코드(`ARBITRATION_STALE_OWNER`, `ARBITRATION_MULTIPLE_OWNER`)를 요구한다.
돌아온 코드를 되짚어 추측하지 않고 스냅샷의 소유자 목록·리비전·신선도 창을 직접 본다.
아무것도 못 찾으면 봉인 문제로 넘긴다 — 모르는 것을 아는 척하지 않는다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- arbiter untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- arbiter tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- engine tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- engine untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- Per-test attribution set: every `Test*` function in `internal/strategyarbiter`, each run alone under the seam tag. This is the whole package, not a sample.
- Measured entry: executed from `Arbitrate` only when the sealed eligible set fails to reconstruct; measured in the arbiter tagged suite.

Exact AST return positions: 202:3, 205:3, 209:3, 211:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 196:2 | arm not entered (arbiter untagged suite); arm entered 3x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAStaleOwnerRevisionIsItsOwnRefusal`, `TestTwoActiveOwnersAreItsOwnRefusal` |
| B2 | if | 197:3 | arm not entered (arbiter untagged suite); arm entered 3x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAStaleOwnerRevisionIsItsOwnRefusal`, `TestTwoActiveOwnersAreItsOwnRefusal` |
| B3 | if | 201:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestTwoActiveOwnersAreItsOwnRefusal` |
| B4 | if | 204:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAStaleOwnerRevisionIsItsOwnRefusal` |
| B5 | if | 207:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAnOwnerSnapshotOutsideItsFreshnessWindowIsAStaleOwner` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `request.EvaluatedAt.IsZero` | 207:6 |
| `request.Snapshot.ObservedAt.IsZero` | 207:39 |
| `request.EvaluatedAt.Before` | 208:4 |
| `request.EvaluatedAt.Before` | 208:64 |

## State mutations and fallbacks

- AST assignments: 2. Defers: 0. Goroutine statements: 0.

## Safety conclusion

판정을 만들지 않고 이유를 좁히기만 한다. 어느 갈래도 제안을 통과시키지 않는다.
