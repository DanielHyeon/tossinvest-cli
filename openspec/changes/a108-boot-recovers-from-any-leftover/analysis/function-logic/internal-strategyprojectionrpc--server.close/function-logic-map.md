# Function Logic Map: `Server.Close`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (232-248)
- AST evidence: `ast.json` (AST 분기 3, revision `current`)
- Risk scan: `risk-pattern-report.md`

**이 change는 이 함수를 바꾸지 않았다.** 여기 있는 이유는 잔재의 **생산자**이기 때문이다 —
회수가 다뤄야 하는 상태 집합은 이 함수가 어디서 중단될 수 있는가로 결정된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s` | nil 허용 | 호출부 | B1에서 nil이면 무동작 `nil` |
| `s.once` | 정확히 1회 실행 | `sync.Once` | 두 번째 `Close`는 첫 결과를 재사용하지 않고 `nil` 반환 |
| `s.descriptor` · `s.socket` · `s.controlDir` | `Start`가 만든 세 경로 | 구조체 필드 | B3: `ErrNotExist`는 실패가 아니다 |
| 종료 예산 | 2초 | `context.WithTimeout` | 초과해도 제거 루프는 실행된다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `s == nil` | 없음 | `nil` | 없음(무변경) |
| B2 | descriptor → socket → controlDir 순회 | **파일·디렉터리 제거** | — | `TestCloseToleratesLeftoverAlreadyRemoved` |
| B3 | 제거 실패이고 `ErrNotExist`가 아니며 아직 오류가 없다 | 첫 오류만 보존 | 그 오류 | `TestCloseToleratesLeftoverAlreadyRemoved` |

## 이 순서가 만드는 잔재 (회수 설계의 입력)

`Shutdown` → descriptor → socket → controlDir. 어디서 중단되느냐가 다음 부팅이 보는
디스크 상태를 정한다.

| 중단 지점 | 남는 것 | design D1 행 |
|---|---|---|
| `Shutdown` 직후 (Go의 unix listener가 socket을 unlink한 뒤) | 디렉터리 + descriptor | **S1 — 8/13 사고** |
| descriptor 제거 후 | 디렉터리 + socket | S2 |
| socket 제거 후 | 빈 디렉터리 | S0 |
| `Shutdown` 전 (SIGKILL·전원 단절) | 둘 다 | S3 |

**네 상태가 전부 이 함수의 중단 지점이다.** 예방(종료 순서 조정)을 채택하지 않은 근거가
이 표다 — 순서를 어떻게 바꿔도 "그 사이에서 죽는" 상태는 남는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.server.Shutdown(ctx)` | listener 닫기 + 진행 중 요청 배수 | 2초 타임아웃. **listener를 닫는 것이 socket 파일을 unlink한다**(Go 기본) | AST calls |
| `os.Remove` ×3 | 잔재 제거 | `ErrNotExist` 용인 | AST calls, B3 |

## State mutations and fallbacks

- 경합: 다음 부팅의 회수가 같은 경로를 먼저 지울 수 있다. B3의 `ErrNotExist` 용인이
  그 경합을 양성으로 만든다(design D2) — `TestCloseToleratesLeftoverAlreadyRemoved`가 핀.
- **관측된 잔여 위험(이 change 밖)**: `Shutdown`이 `Serve`의 listener 등록보다 먼저
  지나가면 listener 정리가 `Serve` goroutine의 defer로 밀리고, unix listener의 unlink는
  **경로 기준**이라 그 늦은 정리가 그 사이에 만들어진 **새 socket**을 지운다. 실측으로
  재현했고(`-count=200`에서 재현), 테스트는 Close 전에 수락을 확인해 그 경합을
  배제했다. 운영에서는 journal flock이 두 엔진의 겹침을 막아 창이 닫힌다.

## Safety conclusion

- Safe edit boundary: 무변경. 제거 순서를 바꾸면 D1 표의 행이 바뀌므로 회수와 함께
  다시 봐야 한다.
- High-risk impact: yes (기동·종료 경로) — 이번 판에서는 편집하지 않았다.
