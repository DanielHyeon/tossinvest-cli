# Branch Test Map: `EntryGate.Clear`

- Source: `internal/execgw/retry.go` (543-555); current
- 시험 칸은 측정값: `analysis/harness/branch_coverage.py` 가 `github.com/JungHoonGhae/tossinvest-cli/internal/execgw` 의 시험 59 개를 하나씩 돌린 커버 프로필(-covermode=set)에서 그 분기 본문 블록을 실행한 시험. 「합집합」은 패키지 전체 한 판.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 547:2 | `if g.clearEpochs == nil {` | `TestAConditionalBlockAfterAClearChangesNothing` · `TestAuthFailureLatchesEntryImmediately` · `TestEveryClearRequestAdvancesTheReasonsEpoch` (+3) | a124 RED — 편집 전에 없던 분기 | 블록 547.26-549.3 을 시험 6 개가 실행, 전부 PASS |
| B2 | if at 551:2 | `if _, exists := g.latches[reason]; exists {` | `TestAuthFailureLatchesEntryImmediately` · `TestEveryClearRequestAdvancesTheReasonsEpoch` · `TestNothingButThatReasonsClearMovesItsEpoch` (+2) | a124 RED — 편집 전에 없던 분기 | 블록 551.44-554.3 을 시험 5 개가 실행, 전부 PASS |
