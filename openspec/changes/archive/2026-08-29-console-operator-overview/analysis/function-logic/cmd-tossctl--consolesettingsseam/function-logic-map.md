# Function Logic Map: `consoleSettingsSeam`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=base, L252–257, 분기 1개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `47672c6f` — 본문 **byte 동일**. 인접 삽입의 diff hunk 교차로 evidence가 요구되었다 (revision=base)

콘솔의 **유일한 config 쓰기 표면**(편입 설정)을 인터페이스로 옮기는 어댑터다. 이 change의 수정 대상이 아니며 인접 삽입(`consoleGateLimitsSeam` 등)의 diff hunk 교차로 evidence가 요구되었다 — 본문은 base와 byte 동일하다.

- **넘어가는 것**: `console.AdoptionSettings` — `Load`와 `Save` 두 메서드. 편입 블록(enabled, default_stop_pct, include/exclude 심볼)만 다룬다.
- **넘어가지 않는 것**: `*config.Service` 자체, Guardian 게이트 블록의 쓰기, 그리고 주문 능력 일체.
- **typed-nil**: nil 판정을 구체 포인터에서 한다. 인터페이스에 담긴 typed-nil은 콘솔의 `Settings != nil` 배선 검사를 무력화한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root *rootOptions` | nil 허용 | `--config-dir`/기본 경로 | 해석 실패 → nil seam → 화면은 미배선 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `s := newAdoptionSettingsSeam(root); s != nil` | 없음 | `s` (인터페이스에 담김) | `TestATypedNilSeamNeverReachesTheInterface` |
| (else) | s == nil | 없음 | 리터럴 `nil` — typed-nil 아님 | 동일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newAdoptionSettingsSeam` | 구체 포인터를 먼저 받는다 | nil이 유일한 실패 신호 | adoptionsettings.go L28 |

## State mutations and fallbacks

- 상태 변이 없음.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입만 존재.
- High-risk impact: yes (손절 파라미터 경로) — 이 seam이 저장하는 `default_stop_pct`는 편입 포지션의 손절 폭이다. 쓰기 능력이 콘솔에 넘어가는 유일한 지점이며, 주문 능력은 넘기지 않는다.
