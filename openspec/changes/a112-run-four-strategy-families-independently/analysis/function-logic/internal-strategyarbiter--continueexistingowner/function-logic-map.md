# Function Logic Map: `continueExistingOwner`

- Source: `internal/strategyarbiter/arbiter.go` (221-237)
- Function: `continueExistingOwner` in package `strategyarbiter`
- Signature: `continueExistingOwner(params=2, results=1)`
- File SHA-256: `1788b0503479c4c4e8b4b17d7e2e6e2fd189f414ebde7877a7d5043157be6d03`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 3.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

봉인된 자격 집합이 활성 소유자를 하나 들고 있을 때 불린다. 결정도 제안도 정확히 하나여야 하고,
그 제안의 계보가 소유자의 `(Horizon, LaneID, LaneVersion, CampaignID)` 와 정확히 같아야 한다.
점수는 소유자를 바꾸지 못한다 — 더 높은 점수의 다른 가족이 있어도 이 함수는 소유자만 이어 간다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- arbiter untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- arbiter tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter ./internal/strategyarbiter/`
- engine tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- engine untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyarbiter,./internal/app/engine ./internal/app/engine/`
- Per-test attribution set: every `Test*` function in `internal/strategyarbiter`, each run alone under the seam tag. This is the whole package, not a sample.
- Measured entry: executed only from `Arbitrate` in the arbiter tagged suite; the engine fixtures do not build an owned snapshot.

Exact AST return positions: 223:3, 229:3, 233:3, 235:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 222:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAnActiveWeeklyOwnerIsNotReplacedByAHigherScore` |
| B2 | if | 227:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAProposalFromAnotherCampaignDoesNotInheritTheOwner` |
| B3 | if | 232:2 | arm not entered (arbiter untagged suite); arm entered 1x (arbiter tagged suite); arm not entered (engine tagged suite); arm not entered (engine untagged suite); entered by `TestAnActiveOwnerOnALaneWithNoScoreRowIsRefused` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 222:5 |
| `len` | 222:35 |
| `refuse` | 223:10 |
| `refuse` | 229:10 |
| `familyScore` | 231:25 |
| `refuse` | 233:10 |

## State mutations and fallbacks

- AST assignments: 3. Defers: 0. Goroutine statements: 0.

## Safety conclusion

소유자 교체 경로가 없다. 어긋나면 닫고, 맞으면 이미 있던 캠페인을 이어 갈 뿐 새 캠페인을 만들지 않는다.
