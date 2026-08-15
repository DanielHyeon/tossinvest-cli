# Function Logic Map: `Dial`

- Source: `internal/strategyprojectionrpc/transport_unix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

a109 §2b.3 G5가 이 함수를 건드리는 범위는 **transport 리터럴의 필드 하나**다
(`DisableKeepAlives: true`). 분기(B1~B3)·반환 5개·호출 집합은 그대로다 — AST가 그것을
편집 전후로 보여 준다.

왜 이 함수인가: 재부착 wrapper가 자리를 갈아끼울 때마다 여기서 만든 client 하나가
밀려난다. 밀려난 client가 **유휴 연결을 쥔 채** 남으면 엔진이 오르내릴 때마다 연결
한 벌이 데몬 수명 내내 쌓인다(security #1 CRITICAL). 두 갈래로 닫는다: 자리는
`Client.Close`를 부르고(밀려난 값), transport는 애초에 유휴 연결을 두지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `descriptorPath` | 공백 제거 후 실재하는 descriptor 경로 | 호출자(`resolveStrategyRuntimeReader`) | `readDescriptor` error 전파 |
| descriptor 내용 | schema·socket 이름·토큰 길이 | `readDescriptor` | error 전파(B1) |
| socket 파일 | socket 비트 · 비symlink · **정확 0600** | `os.Lstat` | "socket is invalid"(B2) |
| socket 주인 | 지금 수락 중 | `projectionSocketAccepts` | "socket has no listener"(B3) |
| ctx | 쓰이지 않는다(`_`) | — | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `readDescriptor` error | 없음 | 그 error | a108 핀(잘린·낯선 descriptor) |
| B2 | `Lstat` error \|\| socket 아님 \|\| perm != 0600 | 없음 | "socket is invalid" | a108 클라이언트 정확-0600 핀 |
| B3 | `!projectionSocketAccepts(socketPath)` | probe 연결 1회(≤200ms) | "socket has no listener" | a108 D4-2 핀 · a109 재부착 시나리오 ② |
| 분기 밖 종단 | 통과 | `http.Transport` 생성(**유휴 연결 없음** — a109 G5) | `(*Client, nil)` | `TestTheDialedTransportKeepsNoIdleConnections` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readDescriptor` | descriptor 파싱·검증 | error 전파 | AST calls (:403) |
| `filepath.Join`·`filepath.Dir` | socket 경로 구성 | — | AST calls (:407) |
| `os.Lstat` | symlink를 따라가지 않는 모양·권한 확인 | error → B2 | AST calls (:408) |
| `projectionSocketAccepts` | 주인 생사 확인(회수와 같은 원시) | false → B3 | AST calls (:417) |
| `net.Dialer.DialContext` | unix socket 연결(요청 시점) | http.Client Timeout 5s | AST calls (:421) |

## State mutations and fallbacks

- 디스크 상태를 바꾸지 않는다. 만드는 것은 프로세스 안의 client 하나뿐이다.
- a109 G5 이후 그 client는 **놓아 줄 수 있다**: `Client.Close`가
  `http.Client.CloseIdleConnections`를 부르고, `DisableKeepAlives: true`이므로 애초에
  요청 사이에 남는 연결이 없다. 진행 중인 읽기는 어느 쪽으로도 끊기지 않는다.

## Safety conclusion

- Safe edit boundary: transport 리터럴의 필드 추가 1개. 회수·발행 의례(a108 확정 코드)와
  descriptor 검증·probe 절은 건드리지 않는다.
- High-risk impact: no — 이 함수는 **조회 클라이언트**를 만든다. 주문·손절 경로에 닿지
  않고, 실패는 이미 강등으로 흡수된다(a108 D4-2). 방향은 보수적이다: 연결을 덜 쥔다.
