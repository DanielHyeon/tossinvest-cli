# Branch Test Map: `TestSchemaTablesAndColumns`

- Source: `internal/journal/schema_test.go` (78-406); AST branch positions are authoritative.
- L5 5.3.3 이 이 함수의 본문(얼린 테이블 목록)을 편집했다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 174:2 — 테이블 질의 실패 | 자기 자신 | 아니오 | 아니오 — 진입하지 않는 arm |
| B2 | for at 178:2 — 실제 테이블 열거 | 자기 자신 | — | 예 |
| B3 | if at 180:3 — 스캔 실패 | 자기 자신 | 아니오 | 아니오 — 진입하지 않는 arm |
| B4 | if at 185:2 — rows.Err | 자기 자신 | 아니오 | 아니오 — 진입하지 않는 arm |
| B5 | if at 189:2 — **얼린 목록과 실제가 다르다** | 자기 자신 | **예 — v32 마이그레이션 직후 실패했고 문구가 `strategy_lane_latches`·`strategy_lane_latch_recoveries` 를 이름으로 가리켰다** | 예 — 두 줄을 선언한 뒤 PASS |
| B6 | for at 399:2 — 칼럼 대조 | 자기 자신 | — | 예 |
| B7 | if at 402:3 — 칼럼 불일치 | 자기 자신 | 아니오 | 아니오 — 진입하지 않는 arm |

RED/GREEN 은 `go test -count=1 -run TestSchemaTablesAndColumns ./internal/journal/` 의 실제
출력이다. 이 열거표는 새 테이블이 조용히 생기는 것을 막는 장치이고, 이번에 실제로 막았다.

> 재기준화: a112 게이트 위생(2026-10-01) — a066 5.5(bce793a7, v35)가 기대 표 · 열 목록에 줄을 더함(분기 불변) — 분기 (id · 종류) · return 수 동일, 좌표는 id 끼리 사상, 호출 표는 새 AST.
