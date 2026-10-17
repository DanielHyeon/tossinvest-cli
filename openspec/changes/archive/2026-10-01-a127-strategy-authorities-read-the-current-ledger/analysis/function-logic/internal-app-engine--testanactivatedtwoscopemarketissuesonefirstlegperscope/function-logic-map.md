# Function Logic Map: `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`

- Source: `internal/app/engine/a112_owner_scope_trading_test.go` (`259`–`300`)
- Qualified: `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`
- AST evidence: `ast.json` (`source_sha256` 9f190f0d32244592…, revision current) — **편집 뒤**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 10 · return 0 · 호출 30

**역할.** a112 두 범위 첫 레그 시험. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | if | `:263` `if risk, account := len(fixture.risk.kr.scopes), len(fixture.accounts.kr.scopes); risk != 2 \|\| account !=…` |
| B2 | if | `:269` `if got := strings.Join(fixture.placedSymbols(), ","); got != "005930" {` |
| B3 | if | `:272` `if firstWave == nil \|\| !strings.Contains(firstWave.Error(), "BUCKET_USAGE_STALE") {` |
| B4 | if | `:275` `if scope := (*strategyScopeRefusal)(nil); errors.As(firstWave, &scope) {` |
| B5 | if | `:279` `if err := sqlOpenReadOnlyCount(t, fixture.journal.Path(), `SELECT count(*) FROM risk_bucket_reservations`, …` |
| B6 | if | `:283` `if err := fixture.deliverKR(t); err != nil {` |
| B7 | if | `:286` `if got := strings.Join(fixture.placedSymbols(), ","); got != "000660,005930" {` |
| B8 | range | `:292` `for _, scope := range fixture.risk.kr.scopes {` |
| B9 | if | `:293` `if scope.ready {` |
| B10 | if | `:297` `if len(digests) != 2 {` |

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: 원장 → stub 복사 다리 제거 — 둘째 파도는 적재기가 실제 원장을 다시 읽음; 준비 단언은 원장 예약 행 수(a127 D4).
- **High-risk impact**: no — 시험. 시험 전용 — 생산 동작 무관.
