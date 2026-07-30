# Function Logic Map: `runInterlock`

- Source: `internal/app/engine/interlock.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

검증 성공 후 `status.EntryPermitted`를 설정하는 줄이 추가됐다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| status.Protection | WIRED | UNWIRED | Options.protectionReadiness() | UNWIRED면 EntryPermitted=false |

## Branches and early returns

분기는 전부 기존 것 — 게이트 OFF 조기 반환, verifyGate 오류, log.Record 오류. 추가된 대입은 분기 없이 실행된다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ internal/app/engine/interlock.go:374 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/app/engine/interlock.go:381 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ internal/app/engine/interlock.go:391 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| verifyGate | 조항 1~8 | 오류면 refuseStartup 후 반환 | 기존 |
| log.Record | 수락 감사 기록 | 오류를 감싸 반환 | 기존 |

## State mutations and fallbacks

- status.Verified와 status.EntryPermitted를 설정한다. 후자가 이 change의 추가분.

## Safety conclusion

- Safe edit boundary: 게이트 OFF 분기는 무수정(§0.3). 추가는 검증 성공 경로의 대입 한 줄.
- High-risk impact: yes
