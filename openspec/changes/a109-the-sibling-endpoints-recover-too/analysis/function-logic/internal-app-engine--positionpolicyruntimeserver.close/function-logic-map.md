# Function Logic Map: `PositionPolicyRuntimeServer.Close`

- Source: `internal/app/engine/position_policy_runtime_transport_unix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

`AlertControlServer.Close`와 달리 이 구조체에는 **`listener` 필드가 이미 있다**
(`position_policy_runtime_transport_unix.go`의 타입 정의). 그런데 Close는 그것을 쓰지
않는다 — 닫는 일을 `Shutdown`에 맡긴다. 필드가 있는데 안 쓰는 것이 더 나쁘다: 소유권이
있다는 표시만 있고 행사는 없다. a109는 `Shutdown` 직후 이 listener를 직접 닫는다
(design D2b).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s` | nil 허용 | 호출자 defer | nil이면 `nil` |
| `s.once` | 한 번만 | `sync.Once` | 두 번째 Close는 무동작 |
| Shutdown 시한 | 2초 | 본문 상수 | 초과 error가 `result` |
| 제거 대상·순서 | descriptor → socket → controlDir | 본문 슬라이스 | ErrNotExist만 관용 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `s == nil` | 없음 | `nil` | nil 수신자 안전 |
| B2 | `range []string{descriptor, socket, controlDir}` | 세 경로 제거 | — | Close 후 셋 다 사라짐 |
| B3 | `err != nil && !errors.Is(err, os.ErrNotExist) && result == nil` | 없음 | 첫 오류만 보존 | ErrNotExist 관용 |
| 분기 밖 종단 | — | `Shutdown` 결과가 기본값 | `result` | 정상 Close는 nil |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `context.WithTimeout` | 2초 시한 | `defer cancel()` | AST defers |
| `s.server.Shutdown` | 요청 정리 | 시한 초과 error 보존 | AST calls |
| `os.Remove` ×3 | 산출물 제거 | ErrNotExist 관용 | AST calls |
| `errors.Is` | 관용 판정 | — | AST calls |

## State mutations and fallbacks

- 세 경로 제거. `s.listener`는 **쓰이지 않는다** — 이것이 a109가 채우는 빈칸이다.

## Safety conclusion

- Safe edit boundary: `Shutdown` 직후 `s.listener.Close()` 절 추가(`net.ErrClosed`는 성공과 같다).
  제거 루프·순서·관용은 불변.
- High-risk impact: yes — 늦은 unlink는 후계자의 socket을 지운다.
