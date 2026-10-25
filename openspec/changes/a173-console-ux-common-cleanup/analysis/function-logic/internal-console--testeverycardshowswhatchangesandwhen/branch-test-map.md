# Branch Test Map: `TestEveryCardShowsWhatChangesAndWhen`

- Source: `internal/console/settings_card_test.go` (83-112) — 편집 뒤. AST branches 6.
- 이 함수 자체가 시험이다. 순회 갈래(B1 · B2 · B4)는 통과 실행이 지나고, 실패 갈래(B3 · B5 · B6)는 회귀 때만 탄다. B6 의 새 리터럴은 a173 GREEN 뒤 `go test ./internal/console/ -count=1` 통과로 성립(실패 갈래 변이는 종결에서 다시 돌리지 않음).

| Branch | Anchor | Scenario | Test | 측정 n | a173 |
|---|---|---|---|---|---|
| B1 | `86:2` | 설정 탭 순회 | `TestEveryCardShowsWhatChangesAndWhen` | — | 무변경 |
| B2 | `87:3` | 카드 순회 | `TestEveryCardShowsWhatChangesAndWhen` | — | 무변경 |
| B3 | `88:4` | 미리보기 없음 → Errorf | `TestEveryCardShowsWhatChangesAndWhen` | — | 무변경(실패 갈래) |
| B4 | `100:2` | adoption · gate 순회 | `TestEveryCardShowsWhatChangesAndWhen` | — | 무변경 |
| B5 | `103:3` | 반영 시점 마커 없음 → Errorf · continue | `TestEveryCardShowsWhatChangesAndWhen` | — | 무변경(실패 갈래) |
| B6 | `107:3` | 마커 앞 class 없음 → Errorf | `TestEveryCardShowsWhatChangesAndWhen` | — | a173: class 리터럴 교체(실패 갈래) |
