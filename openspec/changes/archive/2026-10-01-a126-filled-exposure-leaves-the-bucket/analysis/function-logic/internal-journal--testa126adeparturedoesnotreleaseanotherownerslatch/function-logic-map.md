# Function Logic Map: `TestA126ADepartureDoesNotReleaseAnotherOwnersLatch`

- Source: `internal/journal/a126_filled_exposure_leaves_the_bucket_test.go` (`496`–`532`)
- Qualified: `TestA126ADepartureDoesNotReleaseAnotherOwnersLatch`
- AST evidence: `ast.json` (`source_sha256` d2df830b7ee0d70b…, revision current) — **편집 뒤**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 8 · return 0 · 호출 27

**역할.** 델타 「떠남은 다른 owner 의 latch 를 풀지 않는다」 시험. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. base 재고정(review 2.1) 뒤 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | if | `:500` `if err := a126AdmitSymbol(t, j, "b", "MSFT", "100", "50", 10); err != nil {` |
| B2 | if | `:507` `if overage, _, _ := ownerFlags(t, j, b); overage != 1 {` |
| B3 | if | `:511` `if overage, _, _ := ownerFlags(t, j, b); overage != 1 {` |
| B4 | if | `:515` `if usage.FilledMinor != "24" \|\| usage.HeldMinor != "30" \|\| !usage.OverageLatched {` |
| B5 | if | `:518` `if err := a126Admit(t, j, "c", "1000", 1); !errors.Is(err, ErrRiskBucketEntryBlocked) {` |
| B6 | if | `:522` `if _, err := j.ReleaseRiskOverageLatch(ctx, latchRelease(b, ownerLatchView(t, j, b).StateDigest, &relaxatio…` |
| B7 | if | `:526` `if overage, unknown, rows := ownerFlags(t, j, b); overage != 0 \|\| unknown != 0 \|\| rows != 0 {` |
| B8 | if | `:529` `if usage := a126Usage(t, j, riskbucket.DimensionSector, "sector-tech"); usage.FilledMinor != "30" \|\| usag…` |

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: B latch 를 SQL 주입에서 생산 재계산으로 바꾸고 운영자 해제 뒤 다음 체결 판정을 더함(R1). 단언은 강해지기만 함.
- **High-risk impact**: no — 시험. 생산 동작 변화 0(`64ad3058` 비시험 `.go` 변경 0 — codex 재리뷰 확인).
