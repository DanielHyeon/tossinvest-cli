# Function Logic Map: `TestA126CorruptReceiptedRowsAreUnreadable`

- Source: `internal/journal/a126_filled_exposure_leaves_the_bucket_test.go` (`336`–`374`)
- Qualified: `TestA126CorruptReceiptedRowsAreUnreadable`
- AST evidence: `ast.json` (`source_sha256` d2df830b7ee0d70b…, revision current) — **편집 뒤**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 7 · return 0 · 호출 16

**역할.** 영수증 행 손상 · 불일치 × scope latch 조합의 판독 불가 시험. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. base 재고정(review 2.1) 뒤 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | range | `:349` `for name, corrupt := range cases {` |
| B2 | range | `:350` `for _, reverted := range []bool{false, true} {` |
| B3 | if | `:353` `if reverted {` |
| B4 | if | `:354` `if _, err := j.db.Exec(`INSERT INTO risk_bucket_scope_latches(account_ref,market,symbol,prospective_generat…` |
| B5 | if | `:359` `if _, err := j.db.Exec(corrupt, o.key.ProspectiveGeneration); err != nil {` |
| B6 | if | `:362` `if _, err := riskbucket.ReadJournalBucketUsage(context.Background(), j.db, a126Account, riskbucket.Dimensio…` |
| B7 | if | `:368` `if err := a126AdmitSymbol(t, j, "after-corrupt", "MSFT", "1000", "50", 1); !errors.Is(err, ErrRiskBucketSna…` |

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: 사례 셋 추가(HELD·held 0 — R4, symbol · market 사본 불일치 — R5)와 admission 거절 단언(R8). 기존 사례 · 단언 불변.
- **High-risk impact**: no — 시험. 생산 동작 변화 0(`64ad3058` 비시험 `.go` 변경 0 — codex 재리뷰 확인).
