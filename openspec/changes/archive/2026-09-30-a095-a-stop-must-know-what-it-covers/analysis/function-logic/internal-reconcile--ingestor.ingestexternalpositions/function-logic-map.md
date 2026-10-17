# Function Logic Map: `Ingestor.IngestExternalPositions`

- Source: `internal/reconcile/external.go` (`183`–`297`)
- Qualified: `Ingestor.IngestExternalPositions`
- AST evidence: `ast.json` (`source_sha256` df79472e469f738f…)
- Risk scan: `risk-pattern-report.md`
- 분기 13 · return 10 · 호출 30

**역할.** 로컬 인스턴스가 없는 보유를 원장에 접어 넣고 알린다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in.Alert` | 알림 어댑터 | 배선 | B12 — nil이면 알림 없이 `continue` |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:185` `if len(diff.ExternalPos) == 0 {` | `len` | :186 | 예 |
| B2 | if | `:188` `if in == nil \|\| in.Journal == nil {` | `firstNonEmpty`, `fmt.Errorf`, `strings.TrimSpace` | :189 | 예 |
| B3 | if | `:193` `if account == "" {` | `fmt.Errorf`, `strings.TrimSpace` | :194 | 아니오 |
| B4 | if | `:197` `if asOf == "" {` | `fmt.Errorf` | :198 | 예 |
| B5 | range | `:204` `for _, external := range diff.ExternalPos {` | `strings.ToLower`, `strings.ToUpper`, `strings.TrimSpace` | — | 예 |
| B6 | if | `:207` `if market == "" {` | `strings.ToLower`, `strings.TrimSpace` | — | 예 |
| B7 | if | `:210` `if symbol == "" \|\| market == "" {` | `fmt.Errorf`, `in.Journal.FillWatermark` | :211 | 예 |
| B8 | if | `:218` `if err != nil {` | `fmt.Sprintf`, `in.Journal.ApplyPositionAdjustment`, `string` | :219 | 아니오 |
| B9 | if | `:240` `if errors.Is(err, journal.ErrAdjustmentStale) {` | `errors.Is`, `fmt.Errorf` | :241 | 예 |
| B10 | if | `:245` `if err != nil {` | `fmt.Errorf`, `result.Position.Adopted`, `result.Position.ExitEligible` | :246 | 예 |
| B11 | if | `:258` `if strings.TrimSpace(result.Position.EntryDecisionID) != "" {` | `append`, `fmt.Errorf`, `strings.TrimSpace` | :272 | 아니오 |
| B12 | if | `:278` `if !folded.Applied \|\| in.Alert == nil {` | — | — | 예 |
| B13 | if | `:281` `if err := in.Alert.ExternalPositionFound(ctx, ExternalPositionAlert{` | `append`, `errors.Join`, `fmt.Errorf`, `in.Alert.ExternalPositionFound` | :296 | 예 |

## Calls and live bindings

`in.Journal.FillWatermark` · `in.Journal.ApplyPositionAdjustment` · `in.Alert.ExternalPositionFound`(B13).

결과는 `(Report, error)`다 — 폴드 · 조정 오류는 되던지고, 알림 오류는 모아 끝에 `errors.Join`으로 돌려준다(B13 창).

## State mutations and fallbacks

`position_adjustments` · `positions`(조정 경유) · 알림.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** B12가 `in.Alert == nil`이면 알림을 건너뛴다 — 생산의 `ReconcileDriver`는 이 필드를 nil로 둔다. 따라서 네 번째 발신 자리는 생산에서 도달하지 않는다(보이스 B B-P1-7의 주장을 이 분기로 확인).
- **High-risk impact**: no — 3판에서 무변화.
