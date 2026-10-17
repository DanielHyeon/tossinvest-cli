# Function Logic Map: `LoadRawEngineGate`

- Source: `internal/config/limits_io.go` (42-77, 편집 뒤)
- Qualified: `Service.LoadRawEngineGate`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 9d08b616b9a8c621…, 편집 전 분기 수 4) — 현재 branches 4 · returns 4
- Risk scan: `risk-pattern-report.md`

**역할.** 화면용 게이트 블록 읽기.

**a136 편집.** `MaxDailyBuyNotional` 복사 한 줄(표시용).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if errors.Is(err, os.ErrNotExist) {` (if 44) | — | 분기 본문 참조 | `TestA136TheScreenSaysWhatAnUnsetDailyBuyCeilingDoes` · `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` |
| B2 | `if err != nil {` (if 47) | — | 분기 본문 참조 | `TestA136TheScreenSaysWhatAnUnsetDailyBuyCeilingDoes` · `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` |
| B3 | `if err := json.Unmarshal(data, &doc); err != nil {` (if 56) | — | 분기 본문 참조 | `TestA136TheScreenSaysWhatAnUnsetDailyBuyCeilingDoes` · `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` |
| B4 | `if raw.Enabled != nil {` (if 73) | — | 분기 본문 참조 | `TestA136TheScreenSaysWhatAnUnsetDailyBuyCeilingDoes` · `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` |

Exact AST return positions: 45 · 48 · 57 · 76

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST |

## State mutations and fallbacks

- 설정 파일/audit(해당 시) — 기존과 같음.

## Safety conclusion

- Safe edit boundary: 칸 · 문장 하나 추가 — 기존 분기 무변경.
- High-risk impact: no(읽기)
