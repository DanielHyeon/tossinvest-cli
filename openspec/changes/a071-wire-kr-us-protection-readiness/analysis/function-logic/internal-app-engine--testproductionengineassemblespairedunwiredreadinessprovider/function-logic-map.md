# Function Logic Map: `TestProductionEngineAssemblesPairedUnwiredReadinessProvider`

- Source: `internal/app/engine/a071_readiness_assembly_test.go` (10-25)
- Revision: current — HEAD `648df8ef`; source_sha256 `dc5937ea7eaf9e788842e4ae27381456ab0983f12e031b35ea6409a83997600d`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 3 · returns 0 · calls 9
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| isolated engine config and official stub | production defaults, no manifest pin | engine test harness | fail startup/assertion; never open automation |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 16:2 | `if err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B2 | if at 19:2 | `if !eng.Gateway.HasProtectionReadinessProvider() {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B3 | if at 22:2 | `if eng.Automation.EntryPermitted {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | engine startup succeeds | local test DB/config creation | fail test on error | named test |
| N2 | readiness provider is absent | none | fail test | named test |
| N3 | default readiness opens entry | none | fail test | named test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `startEngine` | exercise real production assembly with stubbed official host | fail test on startup error | CodeGraph + AST |

## State mutations and fallbacks

- Test-only config/database mutation inside an isolated temporary directory.

## Safety conclusion

- Safe edit boundary: production assembly assertions only
- High-risk impact: no (test); protects default OFF behavior
