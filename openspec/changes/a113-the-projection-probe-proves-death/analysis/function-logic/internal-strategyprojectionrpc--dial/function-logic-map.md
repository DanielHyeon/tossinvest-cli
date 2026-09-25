# Function Logic Map: `Dial`

- Source: `internal/strategyprojectionrpc/transport_unix.go`
- AST evidence: `ast.json` — 구현 후 재생성(:405–436, 분기 3·반환 5). 코드 0줄 변경 — 줄 이동과 주석 정정뿐이다.
- Risk scan: `risk-pattern-report.md`

**a113 은 이 함수를 편집하지 않는다.** 이 번들은 design 이 이 함수의 분기를 근거로 쓰기 때문에
만들었다(FLM-before-claiming): `projectionSocketAccepts` 의 B3 삭제가 `Dial` 에 무엇을 바꾸는가.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `descriptorPath` | 실재 descriptor | 소비자(`resolveStrategyRuntimeReader`, `runConsole`) | B1 |
| socket 파일 | socket · **정확 0600** | `os.Lstat` :408 | B2 |
| 주인 | 수락 중 | `projectionSocketAccepts` :417 | B3 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `readDescriptor` 실패 | 없음 | 그 error (:405) | a108 descriptor 핀 |
| B2 | Lstat 실패 · socket 아님 · `Perm() != 0o600` | 없음 | "socket is invalid" (:410) | a108 정확-0600 핀 |
| B3 | `!projectionSocketAccepts(socketPath)` | probe 1회 | "socket has no listener" (:418) | `TestDialRefusesSocketWithNoListener` (a108) |
| 종단 | 통과 | client 생성(DisableKeepAlives) | `(*Client, nil)` (:431) | `TestTheDialedTransportKeepsNoIdleConnections` (a109) |

B2 가 B3 보다 먼저 정확-0600 을 요구하므로, B3 에 닿는 socket 은 Lstat 시점에 owner 쓰기 비트를
가진다 — 삭제되는 `projectionSocketAccepts` B3(owner-write 추정)은 `Dial` 경로에서 Lstat(:408)과
connect(:417) 사이에 권한이 바뀌는 경합에서만 발동할 수 있었다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readDescriptor` | descriptor 검증 | error 전파 | AST :403 |
| `os.Lstat` | 모양·권한 | → B2 | AST :408 |
| `projectionSocketAccepts` | 생존 질문(순수 probe — chmod 없음) | false → B3 | AST :417 |

## State mutations and fallbacks

- 디스크를 바꾸지 않는다. a113 이후에도 **조회 클라이언트는 chmod 하지 않는다** — chmod 는 회수
  전용 함수에만 있다(design D1).

## Safety conclusion

- Safe edit boundary: 코드 편집 없음. 주석 한 문단(:417–418)만 정정 — 회수는 원시 앞에 권한 복원을 두고 여기는 묻기만 한다.
- High-risk impact: no. 행동 차이는 위 경합 창 하나이고, 그때도 결과는 "client 를 주고 첫 Read 가
  실패" — 소비자 재부착(a109 D4)이 받는다.
