# Branch Test Map: TestProposalCollectMarketClosesTheMarketBeforeBuildingEntries (base 이름, 현재 트리에 없음)

- Source: `internal/app/engine/strategy_proposal_ambiguity_test.go` — **base revision** `aeeb209e`,
  base source SHA-256 `3d60a03e9e4c85be223883e1885f496dc48eb7eb1a53b66622f0014b95303794`.
- 이 함수는 삭제됐다. 현재 트리에 실행 대상이 없으므로 커버리지 행을 만들 수 없다.
- Go 커버리지는 `_test.go` 를 계측하지 않으므로, 이 함수는 삭제되기 전에도 커버리지로
  잴 수 없었다. base 에서의 처분은 "실행됐다/안 됐다"가 아니라 "스위트에 포함됐다"이다.

base 리비전에서 이 함수는 `internal/app/engine` 의 무태그 스위트에 포함되어 매 실행마다
1회 돌았다(파일에 빌드 태그가 없다). 이 lot 에서 파일째 삭제됐다.

이 함수가 고정하던 성질을 지금 무엇이 고정하는지에 대한 뮤테이션 영수증
(production source mutated, run, restored from a pristine copy taken before the mutation):

| # | mutation | result | killed by |
|---|---|---|---|
| M-E1 | 중재 거절을 `continue` 로 바꿔 그 종목만 목록에서 뺀다 | KILLED | `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol` |

| Branch | Anchor (base) | Measured disposition |
|---|---|---|
| B1 | if at 30:2 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B2 | if at 36:3 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B3 | if at 41:2 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B4 | if at 48:3 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B5 | if at 54:4 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B6 | if at 57:4 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B7 | range at 62:3 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B8 | if at 64:5 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B9 | if at 70:3 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B10 | if at 75:2 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B11 | if at 83:3 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B12 | if at 87:3 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B13 | if at 91:3 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B14 | if at 94:3 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B15 | if at 95:4 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B16 | if at 101:2 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |
| B17 | if at 104:2 | 함수와 파일이 함께 삭제됐다. base 스위트에서는 무태그 실행마다 1회 돌았고, Go 커버리지는 `_test.go` 를 계측하지 않으므로 이 팔의 진입 횟수는 base 에서도 측정 불가였다. |

A row states what was measured, not what is intended.
