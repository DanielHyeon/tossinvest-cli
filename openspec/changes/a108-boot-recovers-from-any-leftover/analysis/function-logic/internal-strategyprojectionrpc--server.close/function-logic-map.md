# Function Logic Map: `Server.Close`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (358-386)
- AST evidence: `ast.json` — revision `current`. AST 분기 5
- Risk scan: `risk-pattern-report.md`

첫 라운드는 이 함수를 **바꾸지 않았고**, 잔재의 생산자로서만 문서에 있었다. 그때
"관측된 잔여 위험"으로 적어 둔 늦은 unlink를 A1이 300라운드 중 3회로 재현했고,
Fix 라운드가 그것을 닫는다(design D2-2). 분기는 3 → 5로 늘었다: listener nil 검사와
명시적 `Close` 오류 절.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s` | nil 허용 | 호출부 | B1에서 nil이면 무동작 `nil` |
| `s.once` | 정확히 1회 실행 | `sync.Once` | 두 번째 `Close`는 `nil` |
| `s.listener` | `Start`가 만든 listener. 자기 경로를 unlink하지 않는다 | `listenPrivateSocket` | B3: `net.ErrClosed`는 성공과 같다 |
| `s.descriptor` · `s.socket` · `s.controlDir` | `Start`가 만든 세 경로 | 구조체 필드 | B5: `ErrNotExist`는 실패가 아니다 |
| 종료 예산 | 2초 | `context.WithTimeout` | 초과해도 제거는 실행된다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `s == nil` | 없음 | `nil` | 없음(무변경) |
| B2 | `s.listener != nil` | listener를 **여기서** 닫는다 | — | `TestCloseClosesItsOwnListenerWithoutServe` |
| B3 | `Close` 실패이고 `net.ErrClosed`가 아니며 첫 오류다 | 오류 보존 | 그 오류 | `TestCloseClosesItsOwnListenerWithoutServe` |
| B4 | descriptor → socket → controlDir 순회 | **파일·디렉터리 제거** | — | `TestCloseToleratesLeftoverAlreadyRemoved` |
| B5 | 제거 실패이고 `ErrNotExist`가 아니며 첫 오류다 | 오류 보존 | 그 오류 | `TestCloseToleratesLeftoverAlreadyRemoved` |

## 왜 listener를 여기서 닫는가 (design D2-2)

`http.Server.Shutdown`은 **자기가 아는** listener만 닫는다. `Serve`가 아직 listener를
등록하지 않은 사이에 `Shutdown`이 지나가면 정리가 `Serve` goroutine의 defer로 밀리고,
unix listener의 unlink는 **경로 기준**이라 그 늦은 정리가 그 사이에 들어선 후계자의
socket을 지운다. 오늘 그 도달을 막고 있던 것은 flock이 아니라 "예정된 goroutine이
프로세스와 함께 죽는다"는 사실뿐이었다(A1 F5).

Fix 라운드는 두 겹으로 막는다.

1. listener가 기억하는 이름은 최종 경로가 아니라 이미 rename된 임시 이름이다(D1-2).
2. `SetUnlinkOnClose(false)`로 unlink 자체를 끄고, 최종 경로 제거는 B4의 루프가
   **혼자** 소유한다.

그리고 닫는 시점을 goroutine에 맡기지 않는다 — B2가 `Close` 안에서 닫는다.

## 이 순서가 만드는 잔재 (회수 설계의 입력)

`Shutdown` → listener → descriptor → socket → controlDir. 어디서 중단되느냐가 다음
부팅이 보는 디스크 상태를 정한다.

| 중단 지점 | 남는 것 | design D1 행 |
|---|---|---|
| listener 닫은 직후 | 디렉터리 + descriptor + socket | S3 |
| descriptor 제거 후 | 디렉터리 + socket | S2 |
| socket 제거 후 | 빈 디렉터리 | S0 |
| SIGKILL·전원 단절 | 둘 다 | S3 |

첫 라운드 표의 "S1 — 8/13 사고" 행은 **여기서 사라졌다.** 그 행은 "Go의 unix listener가
socket을 unlink한 뒤"를 중단 지점으로 삼았는데, `SetUnlinkOnClose(false)` 이후로 이
함수는 socket을 그렇게 잃지 않는다. S1이 없어졌다는 뜻은 아니다 — 구버전이 남긴
S1과, descriptor 제거가 socket보다 먼저인 순서에서 오는 S2는 그대로 회수 대상이다.

## Calls and live bindings

| Callee | Why called | Error/timeout contract | Evidence |
|---|---|---|---|
| `s.server.Shutdown(ctx)` | 진행 중 요청 배수 | 2초 타임아웃 | AST calls |
| `s.listener.Close()` | listener 소유권 | `net.ErrClosed` 용인 | AST calls, B3 |
| `os.Remove` ×3 | 잔재 제거 | `ErrNotExist` 용인 | AST calls, B5 |

## State mutations and fallbacks

- 경합: 다음 부팅의 회수가 같은 경로를 먼저 지울 수 있다. B5의 `ErrNotExist` 용인이
  그 경합을 양성으로 만든다(design D2).
- **겹2는 강등 후 in-process projection 재시도를 하지 않는다**(design D2-2 명문).
  재시도는 같은 프로세스 안에서 이 창을 다시 연다.

## Safety conclusion

- Safe edit boundary: 제거 순서와 listener 소유권. 순서를 바꾸면 D1 표의 행이 바뀌고,
  소유권을 goroutine에 돌려주면 A1 F5가 되살아난다.
- High-risk impact: yes (기동·종료 경로).
