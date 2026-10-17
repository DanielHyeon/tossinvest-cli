# Function Logic Map: `productionFixture`

- Source: `internal/strategyproposal/production_test.go` (103-106)
- Function: `productionFixture` in package `strategyproposal`
- Signature: `productionFixture(params=3, results=3)`
- File SHA-256: `abfb3e4b1b06d32000fa3b8c4d8ee1361d71b16dd864172e850361a9c63e8969`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

서명된 KR/US 생산 매니페스트와 증거 저장소를 갖춘 픽스처를 만든다.

이 lot 전까지 이 함수가 그 90 줄을 직접 들고 있었다. 태스크 5.4.3 은 **서명 전에**
스코프 하나를 손볼 수 있어야 했고(서명 뒤에 고치면 매니페스트 검증이 먼저 거절해서
정작 보려던 안쪽 경로에 닿지 못한다), 경로 결정과 스코프를 **함께** 다른 레인으로
옮길 수 있어야 했다(하나만 옮기면 자격 집합이 그 스코프를 안 받아 정상 부재로 걸러진다).
그래서 본문을 `productionFixtureOn` 으로 옮기고 이 함수는 "아무것도 바꾸지 않는"
기본값을 넘기는 세 줄이 되었다.

기존 부르는 쪽의 동작은 바뀌지 않는다 — `mutate` 와 `override` 가 둘 다 nil 이면
`productionFixtureOn` 은 옛 본문과 같은 값을 만든다. 그 동치성은 이 파일의 기존
시험들이 그대로 통과하는 것으로 확인했다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

분기가 없다. 근거와 유일한 행은 `branch-test-map.md` 에 있다.

Exact AST return positions: 105:2.

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 104:2 |
| `productionFixtureWith` | 105:9 |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.

## Safety conclusion

시험 전용 도우미다. 생산 바이너리에 들어가지 않으며(`_test.go`), 실계좌 경로에 닿지 않는다.
