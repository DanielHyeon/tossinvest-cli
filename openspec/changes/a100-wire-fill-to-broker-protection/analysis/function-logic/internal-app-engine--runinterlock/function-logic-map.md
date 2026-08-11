# Function Logic Map: `runInterlock`

- Source: `internal/app/engine/interlock.go` (L376-407)
- AST evidence: `ast.json` — 분기 3, return 4, 외부 호출 8
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `gate.Enabled` | bool | `config.AutomationGate` | false면 아무것도 검증하지 않고 아무것도 허용하지 않는다 |
| `status.Protection` | `WIRED`/`UNWIRED` | `execgw.ProfileProtection` 유래 | **현재 빌드는 항상 UNWIRED** |
| `facts` | `gateFacts` | 생성 시점 계좌·attestation 읽기 | `verifyGate` 실패 시 startup 거부 |
| `log` | `*audit.Log` | 호출부 | `Record` 실패는 **에러로 전파** (B3) |

**불변식**: `EntryPermitted`는 `Protection == ProtectionWired`와 **동치**다(L395). 다른 입력이
이 값을 올릴 수 없다. 따라서 보호가 UNWIRED인 한 진입 허용은 구조적으로 불가능하다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error |
|---|---|---|---|
| B1 (L379) | `!gate.Enabled` | `logGateDecision` 1행 | `(status, nil)` — OFF가 기본이자 전부 |
| B2 (L386) | `verifyGate(&status, facts) != nil` | `refuseStartup` audit + `logGateDecision` | `(status, err)` — 기동 거부 |
| B3 (L396) | `log.Record(ActionGateAccepted) != nil` | audit 기록 시도 | `(status, fmt.Errorf(...))` |
| fall-through (L406) | 정상 | `status.Verified = true`; `status.EntryPermitted = (Protection == Wired)`; `logGateDecision` | `(status, nil)` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry 계약 | Evidence |
|---|---|---|---|
| `logGateDecision` | 기동 모드 1행 기록 | 무오류 | AST L382/388/405 |
| `verifyGate` | 게이트 검증 (첫 실패에서 중단) | error 반환 | AST L386 |
| `refuseStartup` | 거부 audit | 반환값 무시(`_ =`) | AST L387 |
| `log.Record` | 수락 audit | **error 전파** | AST L396 |
| `acceptanceDetail` | 한도 전문 문자열 | 순수 | AST L401 |

## State mutations and fallbacks

- `status.Verified`, `status.EntryPermitted` 두 필드만 변경한다.
- fallback 없음. audit 기록 실패는 삼키지 않고 기동을 실패시킨다(B3).
- **a100이 바꿀 지점**: `status.Protection`의 **출처**다. L395의 등식 자체는 올바르므로 유지한다.

## Safety conclusion

- Safe edit boundary: L395 등식은 **바꾸지 않는다.** a100은 `Protection`이 `WIRED`가 될 수 있는
  경로를 만들 뿐이다. 등식을 완화하면 보호 없는 진입이 가능해진다.
- High-risk impact: **yes** — 운영 모드와 진입 허용의 단일 결정 지점.
