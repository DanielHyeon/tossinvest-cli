# Branch Test Map: `ExitObserver.sellIntent`

AST 기준 분기 4 / 이탈 4. 판정 HEAD `102d4e99`. 커버리지는 `go test ./internal/app/engine/ -count=1 -coverprofile` 실측(2026-09-30).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:1573` 관측가 공백 → 기준선 폴백 | 없음 — 커버리지 0, 생산 도달 불가(FLM 「불변식」) | no | no |
| B2 | `:1576` 관측가·기준선 모두 공백 → 거부 | 없음 — 커버리지 0, 거부문 단언 시험 0 | no | no |
| B3 | `:1581` 수량 파싱 실패 | 없음 — 커버리지 0 | no | no |
| B4 | `:1585` 가격 float64 변환 실패 | 없음 — 커버리지 0 | no | no |
| 정상 | `:1588` LIMIT 매도, 가격 = 관측가 | engine e2e 경로(커버리지 1 블록) | no | yes |

Phase 1(P1.2) RED 는 **작성하지 않았다** — 대상 분기(B2 앞 새 단)가 생산에서 도달 불가라는 측정이 먼저 나왔고, 편집 전
Manager 판단을 요청했다(`../../../issues.md` I-P1).
