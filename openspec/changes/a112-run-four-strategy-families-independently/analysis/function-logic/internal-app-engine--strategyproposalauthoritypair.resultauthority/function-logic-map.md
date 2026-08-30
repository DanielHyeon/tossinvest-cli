# Function Logic Map: `strategyProposalAuthorityPair.ResultAuthority`

- Source: `internal/app/engine/strategy_proposal_authority.go` (132-140)
- Function: `strategyProposalAuthorityPair.ResultAuthority` in package `engine`
- Signature: `strategyProposalAuthorityPair.ResultAuthority(params=0, results=1)`
- File SHA-256: `f4f6627f945825c7a199bd38f9152122decd51da410acd8660ea8463093413e0`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 1.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

시장마다 항목이 정확히 하나이고 그 제안이 봉인 그대로일 때만 결과 권한을 내준다.
이 함수는 5.4 에서 바뀌지 않았다. 같은 파일의 다른 함수를 편집해 파일 해시가 달라졌으므로 증거를 다시 만든다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- engine tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/app/engine/`
- engine untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/app/engine/`
- proposal tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/strategyproposal/`
- Per-test attribution set: the seven `Test*` functions that can reach `strategyProposalAuthorityLoader.collectMarket` — the six in `a112_arbitration_test.go` and `strategy_proposal_authority_test.go` plus none elsewhere, because no other engine test constructs a proposal loader or a production assembly. This is the complete reaching set, not a sample.
- Measured entry: executed in the engine tagged suite only.

Exact AST return positions: 135:4, 137:3, 139:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 134:3 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 134:6 |
| `ValidProposal` | 134:34 |
| `value.entries.authority.Proposal` | 134:34 |
| `value.entries.authority.Proposal` | 137:77 |
| `convert` | 139:70 |
| `convert` | 139:110 |

## State mutations and fallbacks

- AST assignments: 1. Defers: 0. Goroutine statements: 0.

## Safety conclusion

`len(entries) != 1` 관문의 세 소비자 중 하나다. 관문 자체는 태스크 5.2 가 바꿀 몫이고 5.4 는 건드리지 않았다.
