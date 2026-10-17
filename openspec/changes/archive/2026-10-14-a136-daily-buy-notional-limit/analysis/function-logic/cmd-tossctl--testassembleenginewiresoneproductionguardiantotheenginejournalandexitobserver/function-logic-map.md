# Function Logic Map: `TestAssembleEngineWiresOneProductionGuardianToTheEngineJournalAndExitObserver`

- Source: `cmd/tossctl/engine_assembly_test.go` (28-143, 편집 뒤)
- Qualified: `TestAssembleEngineWiresOneProductionGuardianToTheEngineJournalAndExitObserver`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` f44202eda0384d01…, 편집 전 분기 수 16) — 현재 branches 16 · returns 0
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture/시험(생산 코드 아님, tossos_testseams).

**a136 편집.** USD 게이트 fixture 에 하루 매수 상한 명시(D5). 시험이 재는 것 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := attest.Save(filepath.Join(dir, attest.FileName), attest.Attestation{` (if 48) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B2 | `switch r.URL.Path {` (switch 62) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B3 | `case "/oauth2/token":` (case 63) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B4 | `case "/api/v1/accounts":` (case 65) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B5 | `default:` (case 67) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B6 | `if err != nil {` (if 86) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B7 | `if constructions != 1 {` (if 89) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B8 | `if !ok {` (if 93) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B9 | `if got := guardian.ExposureLimits(); !reflect.DeepEqual(got, wantLimits) {` (if 104) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B10 | `if !reflect.DeepEqual(ectx.Automation.Limits, wantLimits) {` (if 107) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B11 | `if err != nil {` (if 122) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B12 | `if err != nil {` (if 127) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B13 | `if decision.SafetyClass != journal.SafetyClassRiskReducing {` (if 130) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B14 | `if _, err := ectx.ExitObserver(engine.ExitObserverOptions{Costs: costs.DefaultModel()}); err != nil {` (if 134) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B15 | `if err := ectx.Close(); err != nil {` (if 137) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |
| B16 | `if _, err := sharedJournal.MaxDecisionTTL(context.Background()); err == nil {` (if 140) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./cmd/tossctl` |

Exact AST return positions: 

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 칸 하나.
- High-risk impact: no(시험 전용)
