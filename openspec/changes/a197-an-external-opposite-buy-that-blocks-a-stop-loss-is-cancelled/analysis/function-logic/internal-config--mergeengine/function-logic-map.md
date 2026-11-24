# Function Logic Map: `mergeEngine`

- Source: `internal/config/engine.go` (**편집 전** `330`–`368`, base `b93ed133a`)
- Qualified: `mergeEngine`
- AST evidence: `ast.json` — branches 6 · returns 2
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(a197 재개 개정 2026-11-25, design D8 — 새 토글 칸의 복사 자리). 분기 주장은 `ast.json` 열거에서 왔다.

**역할.** 설정 파일의 `engine` 블록(raw)을 기본값 위에 복사한다. 기본값은 전부 꺼짐. 호출자: 설정 적재(`config` 패키지 안).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cfg` | 기본값이 채워진 `*Engine` | 적재기 | — |
| `raw` | nil 가능 · 각 칸 포인터(nil = 키 없음) | JSON | nil 이면 B1 에서 기본값 유지 |
| 불변식 | 키가 없으면 기본값(꺼짐). `automation_gate` 키가 없는 파일에서도 그 앞의 칸은 반영(a163 D8 · 알림 블록) | 주석 `:343-346` · `:348` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:331` `if raw == nil` | 없음 | `return` `:332` | 기본값 시험 |
| B2 | `:334` `if raw.Autostart != nil` | `cfg.Autostart` | — | `TestAutostartDefaultsOffAndLoadsExplicitValues` |
| B3 | `:338` `if raw.ExitPolicy != nil` | `cfg.ExitPolicy`(검증 포함) | — | exit 정책 설정 시험(미특정) |
| B4 | `:349` `if raw.ExitSourceAgeReport != nil` | `cfg.ExitSourceAgeReport` — **a197 토글의 선례 자리** | — | `TestA163TheSourceAgeReportToggleDefaultsOff` |
| B5 | `:352` `if raw.AutomationGate == nil` | 없음 | `return` `:353` — 뒤의 gate 칸 건너뜀 | gate 시험(미특정) |
| B6 | `:356` `if gate.Enabled != nil` | `cfg.AutomationGate.Enabled` | — | gate 시험(미특정) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `mergeAdoption` | 편입 블록 | 오류 없음 | AST `:337` |
| `strings.TrimSpace` · `next.validate` | exit 정책 | 거절은 값으로 남김 | AST `:339-340` |
| `mergeNotifications` | 알림 블록 | 오류 없음 | AST `:347` |

Live config binding: 이 함수가 설정 값의 원천이다. 엔진은 기동 때만 읽는다(`Context.ExitObserver` `exitwiring.go:358` 선례).

## State mutations and fallbacks

- `cfg` 의 칸만 바꾼다. 키가 없으면 기본값(꺼짐) 유지.

## Safety conclusion

- Safe edit boundary (a197 D8): B4 바로 뒤, B5 의 조기 반환 **앞**에 새 토글 칸 복사 `if raw.<Toggle> != nil { cfg.<Toggle> = *raw.<Toggle> }` 하나를 덧붙인다 — a163 과 같은 이유(키 없는 파일에서 조용히 무시되면 안 됨). 기존 분기 · 순서 무변경. 검증 없음(bool).
- a166 census · 구조 핀 대상 아님(rg). a163 금지 census 의 식별자(`SourceAgeReport` · `SourceTimestamp`)를 새 칸이 쓰지 않으면 무관.
- High-risk impact: yes(운영 토글의 기본값 경로) — 키 없음 · 명시 false = 꺼짐을 시험으로 고정(design §4).
