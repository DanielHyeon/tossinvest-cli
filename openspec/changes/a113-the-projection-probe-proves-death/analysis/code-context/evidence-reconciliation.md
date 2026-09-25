# Evidence reconciliation — a113

- 기준 HEAD `54004f44` · CodeGraph 1.6.0 · Go AST 번들 4개(`analysis/function-logic/`)

| 사실 | CodeGraph | AST/HEAD | 결론 |
|---|---|---|---|
| probe 호출자 | reclaim · Dial · 판정 표 테스트 | AST calls: reclaim :293, Dial :417 | 일치 |
| probe callee `conn.Close` | `internal/journal/readonly_schema_arms_test.go:51` 의 `Close` | AST :390 `conn.Close` — `net.Conn` 메서드 | **불일치 — CodeGraph 오해석**(이름 기반 해소). HEAD 가 정본. 편집 판단에 영향 없음 |
| 영향 범위 | impact 가 패키지 안 5 심볼에서 멈춤 | HEAD grep: `Dial` 소비자 cmd/tossctl(console·httpapi), `Start` 소비자 engine.go | CodeGraph impact 는 패키지 경계 밖 호출을 못 셈 — HEAD grep 으로 보완. 소비자 동작 변화는 Dial 경합 창 하나(design) |
| 영향 테스트 | 기본 affected 0 Go, `--filter` 3 파일 | cmd/tossctl a108·a109 재부착 테스트 3 파일이 Dial 을 탄다 | `--filter` 결과 + cmd 회귀 실행으로 덮는다 |
| CGC | 심볼 없음 | — | 보조 문맥 not-applicable |

불일치 해소 후 편집 차단 사유 없음.
