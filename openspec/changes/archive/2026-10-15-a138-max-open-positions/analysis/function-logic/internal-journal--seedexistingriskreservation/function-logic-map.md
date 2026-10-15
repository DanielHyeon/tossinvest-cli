# Function Logic Map: `seedExistingRiskReservation`

- Source: `internal/journal/risk_bucket_admission_test.go` (406-422)
- Qualified: `seedExistingRiskReservation`
- AST evidence: `ast.json`(현재) — 현재 AST branches 3 · returns 0
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 67d1349d05075d75…) — 편집 전 갈래 수 2
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture — 원장에 직접 심는 기존 HELD 예약.

**a138 편집.** 결정 preimage 를 `'{}'` 에서 생산 RecordDecision 이 쓰는 정규형 RiskIntent 로 — 개수 판정이 HELD 예약의 종목을 읽으므로(옛 모양은 생산이 만들 수 없음).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 413) | — | 분기 본문 참조 | `go test ./internal/journal -run A066` · `TestA066StaleUsageRuleEdgesAtBothAdmissionSites` |
| B2 | `if _, err := j.db.Exec(`INSERT INTO decisions(id,account_ref,generation,safety_class,preimage_kind,risk_preimage,risk_hash,nonce,issued_at,expires_at) VALUES(?,?,0,'EXPOSURE_RAISING','RISK_INTENT',?,?,?,?,?)`, decisionID, account, preimage, "hash-"+id, "nonce-"+id, "2026-03-30T00:29:00Z", "2026-03-30T01:30:00Z"); err != nil {` (if 416) | — | 분기 본문 참조 | `go test ./internal/journal -run A066` · `TestA066StaleUsageRuleEdgesAtBothAdmissionSites` |
| B3 | `if _, err := j.db.Exec(`INSERT INTO risk_reservations(id,decision_id,account_ref,kind,amount,currency,snapshot_as_of,state) VALUES(?,?,?,'OPEN_EXPOSURE','1','KRW','2026-03-30T00:29:00Z','HELD')`, id, decisionID, account); err != nil {` (if 419) | — | 분기 본문 참조 | `go test ./internal/journal -run A066` · `TestA066StaleUsageRuleEdgesAtBothAdmissionSites` |

Exact AST return positions: (없음)

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 4종 — 대표: `Canonical`, `j.db.Exec`, `t.Fatal`, `t.Helper`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
