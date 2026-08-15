# Function Logic Map: `PositionPolicyRuntimeServer.Close`

- Source: `internal/app/engine/position_policy_runtime_transport_unix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

`AlertControlServer.Close`와 달리 이 구조체에는 **`listener` 필드가 이미 있었다**
(`position_policy_runtime_transport_unix.go`의 타입 정의). 그런데 Close는 그것을 쓰지
않았다 — 닫는 일을 `Shutdown`에 맡겼다. 필드가 있는데 안 쓰는 것이 더 나쁘다: 소유권이
있다는 표시만 있고 행사는 없다.

a109가 `Shutdown` 직후 이 listener를 직접 닫는다(design D2b). 그리고 §2b.3 G9가 그 해체를
형제와 공유하는 `closePrivateEndpointFiles`로 옮겼다 — 지금 이 함수의 AST 분기는
**1개**(nil 수신자)다(편집 전 3개).

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
| 분기 밖 종단 | — | `Shutdown` → `closePrivateEndpointFiles`(listener 해체 + 세 경로 제거) | `result` | 정상 Close는 nil |

옛 B2~B5(listener 직접 닫기 · `net.ErrClosed` 관용 · 세 경로 제거 · `os.ErrNotExist`
관용)는 a109 §2b.3 G9가 `closePrivateEndpointFiles`로 옮겼다. 조건도 순서도 그대로이고,
그 절들의 자기 문서는 이제 그 기계에 있다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `context.WithTimeout` | 2초 시한 | `defer cancel()` | AST defers |
| `s.server.Shutdown` | 요청 정리 | 시한 초과 error 보존 | AST calls |
| `s.listener.Close` | 늦은 unlink 차단(직접 소유) | `net.ErrClosed` 관용 | AST calls |
| `os.Remove` ×3 | 산출물 제거 | ErrNotExist 관용 | AST calls |
| `errors.Is` | 관용 판정 ×2 | — | AST calls |

## State mutations and fallbacks

- 세 경로 제거. `s.listener`는 이제 **여기서 닫힌다** — 그 빈칸이 a109가 채운 것이다.
  경로를 지울 권한은 제거 루프 하나뿐이다(`SetUnlinkOnClose(false)` + 임시 이름).

## Safety conclusion

- Safe edit boundary: `Shutdown` 직후 `s.listener.Close()` 절만 추가했다(`net.ErrClosed`는
  성공과 같다). 제거 루프·순서·관용은 불변.
- High-risk impact: yes — 늦은 unlink는 후계자의 socket을 지운다.
