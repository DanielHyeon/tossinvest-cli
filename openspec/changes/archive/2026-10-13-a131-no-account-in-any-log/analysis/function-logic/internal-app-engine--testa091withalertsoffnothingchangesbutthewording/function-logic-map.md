# Function Logic Map: `TestA091WithAlertsOffNothingChangesButTheWording`

- Source: `internal/app/engine/a091_stop_sold_nothing_test.go` (226-310)
- Qualified: `TestA091WithAlertsOffNothingChangesButTheWording`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` 6901ac61ba493906…) — AST branches 12 · returns 0
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 시험 함수(a131 이 기대값만 바꿈). a091 B2 로그 줄의 `position_id` 기대값을 치환된 꼴로 바꿈. payload 기대값은 원문 그대로. 편집 뒤 작성 — 생산 코드가 아니라 기대값의 정정이며 시험의 분기·흐름은 그대로.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 하네스 | — | 같은 파일의 픽스처 | 실패는 t.Fatal |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 235:3 | `if n := len(h.alerts.events); n != 0 {` | 무편집 |
| B2 | if at 247:3 | `if h.alerts.count(obs.EventExitStopSoldNothing) != 0 \|\| h.alerts.count(obs.EventExitProposalCapped) != 1 {` | 무편집 |
| B3 | if at 251:3 | `if want := "005930 청산이 확정 하한에 걸려 한 주도 나가지 않았다"; a.Title != want {` | 무편집 |
| B4 | if at 260:3 | `if len(lines) != 1 {` | 무편집 |
| B5 | if at 264:3 | `if l["level"] != "ERROR" \|\| l["cause"] != "floor_unknown" \|\| l["symbol"] != "005930" \|\| l["position_id"]` | 무편집 |
| B6 | if at 267:3 | `if _, ok := l["account"]; ok \|\| strings.Contains(fmt.Sprint(l), exitAccount) {` | 무편집 |
| B7 | if at 270:3 | `if len(r.log.ofEvent(t, obs.EventExitStopSoldNothing)) != 0 {` | 무편집 |
| B8 | if at 284:3 | `if a.Body != want {` | 무편집 |
| B9 | range at 288:2 | `for _, cause := range []struct {` | 무편집 |
| B10 | if at 299:4 | `if rows, _ := r.journal.PendingAlerts(context.Background(), 0); len(rows) != 0 {` | 무편집 |
| B11 | if at 302:4 | `if rej := r.gate.CheckEntry(); rej != nil {` | 무편집 |
| B12 | if at 305:4 | `if got := r.mode(); got != journal.ModeNormal {` | 무편집 |

반환문 없음(AST returns 0).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 하네스·대상 함수 | 시험 | — | AST calls |
| `redact.String`/`redact.OpenAccount`(일부) | 치환·봉인 기대값 계산 | 순수 | a131 |

## State mutations and fallbacks

- 시험 상태만.

## Safety conclusion

- Safe edit boundary: 기대값 식만 바꿈 — 분기·흐름 불변.
- High-risk impact: no — 시험 기대값.
