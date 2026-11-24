# Branch Test Map: `Context.ExitObserver`

**편집 전 좌표**(AST 6 분기, base `064f59fd9` → 재고정 `b93ed133a`, `exitwiring.go` `319`–`365`). 실행 0 — 덮는 시험은 미측정(이름 grep 으로 특정하지 않았다). 편집할지는 Manager 결정(판독기 · 결정 발행기 주입 자리) 뒤 확정.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil Context 거절 | 미측정 | no | no |
| B2 | automation gate 미검증 ∧ 청산 전용 아님 → 거절(a199) | 미측정 | no | no |
| B3 | Guardian 이 감소 결정 발행기가 아님 | 미측정 | no | no |
| B4 | 이름 표 기본값 | 미측정 | no | no |
| B5 | 알림기 있을 때 기록 전용 · critical 입구 배선 | 미측정 | no | no |
| B6 | 확정 하한 기본값 | 미측정 | no | no |
