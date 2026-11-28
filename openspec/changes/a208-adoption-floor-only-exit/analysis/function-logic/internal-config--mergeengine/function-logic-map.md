# Function Logic Map: `mergeEngine`

- Source: `internal/config/engine.go` (344-386)
- Qualified: `mergeEngine`
- AST evidence: `ast.json` (source_sha256 `3275a148ad43…`) — AST branches 7 · returns 2
- Risk scan: `risk-pattern-report.md`(일치 0)
- 작성 시점: 편집 **전**(a208 F3 재작업, 2026-11-28, HEAD `468ec87ce` — base `baeddfe31` 와 이 함수 본문 동일)(파일은 a208 이 mergeAdoption · 타입 필드만 고침)

**역할.** 설정 파일의 engine 블록을 기본값 위에 병합한다. 호출자 `Service` 로더(service.go:335) 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw *rawEngine` | nil 또는 파싱 결과 | json.Unmarshal(엄격 아님) | nil → 기본값 |
| 불변식 | 기본값은 전부 꺼짐 | 함수 주석 | — |
| 불변식 | 알림 · a163 · a197 키는 automation_gate 조기 반환 앞 | 주석 | — |

## Branches and early returns

| Branch | Condition | At | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if raw == nil` | 345:2 | — | return (346:3) | 미측정 |
| B2 | Autostart 있음 | 348:2 | cfg.Autostart | — | 미측정 |
| B3 | ExitPolicy 있음 | 352:2 | cfg.ExitPolicy(+Rejected) | — | `TestUnknownCommonExitPolicyIsRetainedAsRejected` |
| B4 | ExitSourceAgeReport 있음 | 363:2 | cfg.ExitSourceAgeReport | — | a163 시험 |
| B5 | CancelExternalOppositeBuy 있음 | 367:2 | cfg.CancelExternalOppositeBuy | — | a197 시험 |
| B6 | AutomationGate 없음 | 370:2 | — | return (371:3) | 미측정 |
| B7 | gate.Enabled 있음 | 374:2 | cfg.AutomationGate.Enabled | — | 미측정 |

Exact AST return positions: 346:3, 371:3

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `mergeAdoption` | 편입 블록 | 거부 시 전면 0 | AST |
| `ExitPolicy.validate` | 공통 정책 ID 등록 검사 | 사유 문자열 | AST |
| `mergeNotifications` | 알림 블록 | — | AST |

Live config binding: 설정 파일 → 기동 때 1 회.

## State mutations and fallbacks

- cfg 필드 대입만. fallback 없음.

## Safety conclusion

- Safe edit boundary(a208 F3): B3 블록 **뒤** · 알림 병합 앞에 한 줄 — `cfg.Adoption.FloorOnlyIgnored = cfg.Adoption.floorOnlyIgnoredReason(cfg.ExitPolicy.CommonPolicy)`
  (새 파일 함수, `json:"-"` 파생 필드 — 저장 바이트 무관). 편입 블록 · 공통 정책 블록이 둘 다 병합된 뒤라야 판정할 수 있어 이 자리. 분기 0 추가.
- High-risk impact: no(기동 audit 문구용 파생 값; 판정은 엔진이 exitpolicy 함수로 따로 함).
