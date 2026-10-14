# Function Logic Map: `Save`

- Source: `cmd/tossctl/limitsettings.go` (48-71, 편집 뒤)
- Qualified: `consoleLimitSettings.Save`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 06572b0a355b9c26…, 편집 전 분기 수 3) — 현재 branches 5 · returns 2
- Risk scan: `risk-pattern-report.md`

**역할.** 콘솔 한도 seam 의 저장 + 필드별 audit.

**a136 편집.** 저장 뒤 파일 값을 다시 읽어 하루 매수 상한 변화(폼 값 · 통화 전환 지움)를 `recordDailyBuySave` 로 audit.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := s.svc.SaveEngineGateLimits(l); err != nil {` (if 50) | — | 분기 본문 참조 | `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` · `TestTheLimitSeamSavesAuditsAndLeavesTheSwitchAlone` |
| B2 | `if beforeErr == nil {` (if 56) | — | 분기 본문 참조 | `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` · `TestTheLimitSeamSavesAuditsAndLeavesTheSwitchAlone` |
| B3 | `} else {` (else 58) | — | 분기 본문 참조 | `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` · `TestTheLimitSeamSavesAuditsAndLeavesTheSwitchAlone` |
| B4 | `if after, err := s.svc.LoadRawEngineGate(); err == nil {` (if 63) | — | 분기 본문 참조 | `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` · `TestTheLimitSeamSavesAuditsAndLeavesTheSwitchAlone` |
| B5 | `if beforeErr == nil {` (if 65) | — | 분기 본문 참조 | `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` · `TestTheLimitSeamSavesAuditsAndLeavesTheSwitchAlone` |

Exact AST return positions: 51 · 70

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST |

## State mutations and fallbacks

- 설정 파일/audit(해당 시) — 기존과 같음.

## Safety conclusion

- Safe edit boundary: 칸 · 문장 하나 추가 — 기존 분기 무변경.
- High-risk impact: yes(§0.5 audit)
