# Function Logic Map: `closeRiskBucketOwnerLifecycle`

- Source: `internal/journal/risk_bucket_owner_test.go` (`808`–`834`)
- Qualified: `closeRiskBucketOwnerLifecycle`
- AST evidence: `ast.json` (`source_sha256` 40d2f896e7e09c8e…) — **편집 뒤**(시험 전용 fixture)
- Risk scan: `risk-pattern-report.md`
- AST branches 7 · return 0 · 호출 21

**역할.** owner 해제 시험의 수명주기를 닫는 fixture — Position · campaign · claim · legacy 예약을 종결 상태로. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님, 게이트가 수정 함수로 세므로 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` · 경로 · 원장 | 시험 임시 원장 | 시험 | `t.Fatal` |

## Branches and early returns

> 표는 `analysis/harness/branch_table.py` 가 만들었다(시험 코드라 커버리지 블록 없음 — `—`).

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:811` `if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_orders o JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHER…` | — | — |
| B2 | if | `:815` `if orders == 0 {` | — | — |
| B3 | if | `:818` `if _, err := j.db.Exec(`UPDATE positions SET state='CLOSED',quantity='0',closed_at='2026-03-30T00:35:00Z' WHERE account_ref=? AND market=…` | — | — |
| B4 | if | `:821` `if _, err := j.db.Exec(`UPDATE position_campaigns SET state='CLOSED',entry_blocked=1,version=version+1,updated_at='2026-03-30T00:35:00Z' …` | — | — |
| B5 | if | `:824` `if _, err := j.db.Exec(`DELETE FROM position_campaign_claims WHERE campaign_id=(SELECT campaign_id FROM risk_bucket_owners WHERE account_…` | — | — |
| B6 | if | `:827` `if _, err := j.db.Exec(`UPDATE risk_reservations SET state='RELEASED',released_at='2026-03-30T00:36:00Z',release_reason='BROKER_TERMINAL'…` | — | — |
| B7 | if | `:831` `if reconcile {` | — | — |

## Calls and live bindings

원장(시험 임시 DB) 문장 실행과 `t.Fatal`. 생산 코드 · 브로커 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: a126 1.4: 주문이 하나도 없는 owner 는 먼저 `fillRiskBucketOwnerInFull`(생산 작성자: 등록 주문 · RecordFill · actual 보완)로 실제 체결을 쌓는다. 예전의 `UPDATE risk_bucket_reservations SET held_minor='0',state='FILLED'`(filled_minor 를 '0' 으로 남겨 갭을 가린 SQL)는 지웠다 — 예약은 체결로 FILLED · held 0 이 되거나(체결 owner) 취소로 RELEASED 가 된다(주문 있는 owner). 가림 주석도 지우고 이유를 새 주석에 적었다. a066 해제 시험 전부 이 fixture 위에서 통과(`analysis/impl/regress-1.log`).
- **High-risk impact**: no — 시험 fixture.
