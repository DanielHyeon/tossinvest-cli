# Function Logic Map: `mergeEngine`

- Source: `internal/config/engine.go` (330-368)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 기준: 편집 뒤(브랜치 `a163-impl` GREEN 로트, 2026-10-22). 편집 전 번들 문서는 커밋 `78efe1951` · 생성 입력 `../../flm-gen/spec-base.json`
- AST 분기 6 · 반환 2 · defer 0 · go 0
- 위험 등급: Normal(설정 파싱) — 단 「키 없음 = 꺼짐」 불변식이 OFF=오늘 동작(§0.2)을 지킴

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw *rawEngine` | nil 또는 파싱된 engine 블록 | config 파일 | nil 이면 기본값(전부 off) |
| `raw.AutomationGate` | nil 이면 B4 조기 반환 | config 파일 | 그 뒤 필드는 병합 안 됨 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 331 `if raw == nil {` | — | `return` | `TestAConfigWithNoNotificationsBlockIsUnchanged` |
| B2 | if 334 `if raw.Autostart != nil {` | `cfg.Autostart` | — | `TestAutostartDefaultsOffAndLoadsExplicitValues` |
| B3 | if 338 `if raw.ExitPolicy != nil {` | `cfg.ExitPolicy` | — | `TestUnknownCommonExitPolicyIsRetainedAsRejected` |
| B4 | if 349 `if raw.ExitSourceAgeReport != nil {` | **(a163 새 분기)** `cfg.ExitSourceAgeReport` | — | `TestA163TheSourceAgeReportToggleDefaultsOff` |
| B5 | if 352 `if raw.AutomationGate == nil {` | — | `return`(조기, 옛 B4) | `TestNotificationsMergeWithoutAnAutomationGate` |
| B6 | if 356 `if gate.Enabled != nil {` | `cfg.AutomationGate.Enabled`(옛 B5) | — | `TestAutomationGateParsed` |

반환 좌표(AST): 332 · 353

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `mergeAdoption` | 편입 블록 | 거부 시 0 블록 | AST |
| `mergeNotifications` | 알림 블록(B4 앞) | — | AST |

## State mutations and fallbacks

- `cfg` 의 engine 필드들.

## a163 편집

- a163 3.6: B4 조기 반환 **앞**에 `if raw.ExitSourceAgeReport != nil { cfg.ExitSourceAgeReport = *raw.ExitSourceAgeReport }` — 분기 5→6, 반환 2 불변(design D8). 새 분기는 B2 와 같은 모양. — **적용됨**, 편집 전·뒤 AST 분기·반환·defer 수 대조는 `../../flm-gen/README.md`

## Safety conclusion

- Safe edit boundary: 새 `if` 하나, 기존 분기 무편집.
- High-risk impact: no(설정 파싱) — 변이 ⑱⑲ 는 2.13 이 잡는다.
