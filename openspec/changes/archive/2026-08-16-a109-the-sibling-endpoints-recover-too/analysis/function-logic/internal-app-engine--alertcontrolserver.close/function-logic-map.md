# Function Logic Map: `AlertControlServer.Close`

- Source: `internal/app/engine/alert_control_transport_unix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**편집 전에는 listener 필드가 없었다.** 구조체가 `server *http.Server`만 들고 있어서
Close는 `Shutdown`이 listener를 닫아 주기를 기다렸다. a108이 실측한 late-unlink 경합
(A1 F5, 300라운드 중 3회)이 그 위임에서 나온다: `Shutdown`이 `Serve`의 listener 등록을
앞지르면 정리가 Serve goroutine의 defer로 밀리고, 늦게 도착한 unlink가 이미 그 경로에
앉은 **후계자의 socket**을 지운다.

a109가 listener 필드를 추가하고 Close가 직접 닫는다(design D2b). 그리고 §2b.3 G9가 그
해체를 형제와 공유하는 `closePrivateEndpointFiles`로 옮겼다 — 지금 이 함수의 AST 분기는
**1개**(nil 수신자)다(편집 전 3개).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s` | nil 허용 | 호출자 defer | nil이면 `nil` 반환 |
| `s.once` | 정확히 한 번 실행 | `sync.Once` | 두 번째 Close는 무동작·`nil` |
| Shutdown 시한 | 2초 | 본문 상수 | 초과 시 그 error가 `result` |
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
| `context.WithTimeout` | Shutdown 시한 2초 | `defer cancel()` | AST defers |
| `s.server.Shutdown` | 진행 중 요청 정리 | 시한 초과 error를 `result`로 | AST calls |
| `closePrivateEndpointFiles` | listener 직접 해체 + descriptor·socket·controlDir 제거 | `net.ErrClosed`·ErrNotExist 관용, 첫 오류만 보존 | AST calls (a109 G9) |

## State mutations and fallbacks

- 디스크에서 세 경로가 사라진다. 순서가 계약이다 — descriptor를 먼저 지워야 "socket은
  있는데 descriptor가 없는" 잔재 모양이 나오고, 그 모양은 회수가 아는 모양이다.
- listener 소유권이 이제 명시적이다. 그리고 경로를 지울 권한은 제거 루프 하나뿐이다 —
  `SetUnlinkOnClose(false)`이고 listener가 기억하는 이름은 이미 rename으로 사라진 임시
  이름이라 두 겹으로 막힌다.

## Safety conclusion

- Safe edit boundary: 구조체에 `listener` 필드를 더하고 `Shutdown` 직후 그것을 닫는 절만
  추가했다. **제거 루프·ErrNotExist 관용·제거 순서는 그대로다**(design D2b).
- High-risk impact: yes — 늦은 unlink는 후계자의 endpoint를 지운다.
