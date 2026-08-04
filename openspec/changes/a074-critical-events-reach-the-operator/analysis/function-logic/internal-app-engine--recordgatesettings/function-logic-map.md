# Function Logic Map: `recordGateSettings`

- Source: `internal/app/engine/interlock.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a074-critical-events-reach-the-operator/base-commit.txt`
- 위험 등급: **High-risk** — §0.5의 audit trail을 만드는 함수다. 실패하면
  `NewContext`가 기동을 거부한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `log *audit.Log` | 열린 audit 로그 | `openAuditLog` | write 실패 → error 반환 → **기동 거부** |
| `gate config.AutomationGate` | 게이트 설정 | 설정 파일 | — |
| `adoption config.Adoption` | 편입 설정 | 설정 파일 | 거부 사유는 detail에 |
| `attestationFile string` | attestation 경로 | 상태 | — |

**불변식 1 (유지)**: 이 함수는 **어떤 거부보다 먼저** 호출된다(`engine.go:418-427`).
주석이 이유를 적는다 — "an operator's settings change is worth recording whether or not
the engine then agrees to start on it."

**불변식 2 (유지)**: write 실패는 기동 거부다. `NewContext`의 주석 — "A settings change
we cannot record is a settings change nobody can audit."

**불변식 3 (유지)**: 항목은 `changes` 리터럴의 **선언 순서대로** 기록된다. 순서를
바꾸면 기존 audit trail과 새 trail을 나란히 읽을 수 없다.

**a074가 바꾸는 것**: `changes` 리터럴에 알림 설정 4항목을 추가하고, 그 값을 받기 위해
매개변수를 하나 더한다.

**a074가 바꾸지 않는 것**: 기존 11항목의 순서·action·setting 이름·값 형식. 새 항목은
**끝에** 붙는다.

## Inputs and invariants — 새 매개변수

| Input | Valid range | 기록되는 값 |
|---|---|---|
| `notifications config.Notifications` | 설정 파일의 알림 블록 | `enabled`(bool) · `base_url`(문자열) · `topic_configured`(bool) · `token_configured`(bool) |

**topic과 token의 값은 기록하지 않는다** (design D5). ntfy.sh 구성에서 topic 이름은
유일한 접근 제어이며 토큰과 같은 성질이다. §0.8이 시크릿을 로그에 남기는 것을
금지한다. §0.5가 요구하는 "언제·누가·무엇을 바꿨나"는 **설정 여부**로 답해진다.

`token_configured`는 환경에서 해석된 결과이므로 호출자가 계산해 넘긴다 — 이 함수도,
`internal/config`도 환경변수를 읽지 않는다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (712) | `adoption.Rejected != ""` | `adoptionDetail` 채움 | — | 기존 |
| **신규** | `notifications.Rejected != ""` | `notificationDetail` 채움 | — | **6.3** |
| B2 (742) | `changes` 순회 | audit write | — | 기존 |
| B3 (743) | `log.RecordChange` 실패 | 없음 | `err` → 기동 거부 | 기존 |

새 분기는 B1과 같은 모양의 detail 조립 하나뿐이다. 순회(B2)와 실패 전파(B3)는 그대로
새 항목에도 적용된다 — 알림 설정을 audit에 못 남기면 기동이 거부된다. 그것이 §0.5의
방향이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `log.RecordChange` | 항목 1건 기록 | 실패 즉시 반환 | AST |
| `strconv.FormatBool` | bool → 문자열 | — | AST |
| `limitString` | float → 문자열 | — | AST |
| `strings.Join` | 심볼 목록 | — | AST |
| `strings.ToUpper`/`TrimSpace` | 통화 정규화 | — | AST |

새 항목은 `strconv.FormatBool`과 문자열 그대로만 쓴다. 새 의존이 없다.

## State mutations and fallbacks

- audit 로그 파일에 append. 그 외 상태 없음.
- fallback 없음. 실패는 전파된다.
- 새 항목이 실패해도 앞선 항목은 이미 기록되어 있다 — 이것은 기존 동작이며 audit이
  append-only이므로 부분 기록이 곧 사실의 부분 기록이다.

## Safety conclusion

- Safe edit boundary: 시그니처에 매개변수 1개, B1 옆에 detail 조립 1개, `changes`
  리터럴 **끝에** 항목 4개.
- High-risk impact: **yes** — 기존 항목을 재배열하거나 이름을 바꾸면 audit trail이
  깨진다. 이 편집은 append만 한다.
- §0.5: 알림 설정이 처음으로 audit 대상이 된다. `engine.go`의 `Publisher` 주석이
  요구한 조건이 이것으로 충족된다.
- §0.8: topic·token 값이 audit에 나타나지 않음을 6.2가 테스트로 고정한다.
