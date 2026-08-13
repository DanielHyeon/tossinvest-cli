# Function Logic Map: `Dial`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (216-230)
- AST evidence: `ast.json` (AST 분기 2, revision `current`)
- Risk scan: `risk-pattern-report.md`

**이 change는 이 함수를 바꾸지 않았다.** 여기 있는 이유는 잔재의 **소비자**이기 때문이다 —
httpapi가 반쪽 잔재에서 crash loop에 빠진 경로(겹3, T2 소관)가 이 함수의 실패다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `descriptorPath` | `…/.strategy-runtime-read/endpoint.json` | 호출부(콘솔·httpapi) | `readDescriptor`가 경로 모양부터 검사(transport.go:93) |
| descriptor 내용 | 스키마 v1 · socket 이름 · 토큰 32자 이상 · PID > 0 | 디스크 | B1 — 오류 그대로 전파 |
| socket 파일 | socket 타입 · perm 정확히 0600 | 디스크 | B2 `socket is invalid` |
| `ctx` | **무시된다** | — | 첫 인자는 `_`다. 타임아웃은 client의 5초뿐 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `readDescriptor` 실패 | 없음 | 원본 오류 | `TestStartRecoversFromDescriptorOnlyLeftover` (회수 후 성공 쪽) |
| B2 | socket lstat 실패 · socket 아님 · perm != 0600 | 없음 | `socket is invalid` | `TestCloseToleratesLeftoverAlreadyRemoved` (경합 배제 전 실측으로 이 오류를 관측) |

`Dial`은 **연결하지 않는다** — transport를 만들어 돌려줄 뿐이고 실제 연결은 첫 요청에서
일어난다. 그래서 "socket 파일이 있다"까지만 확인하며, 그 파일에서 아무도 수락하지 않는
경우는 `Client.Read`의 오류가 된다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readDescriptor` (transport.go:92) | 토큰·socket 이름 획득 + 파일 검증 | 오류 그대로 전파 | AST calls, B1 |
| `os.Lstat` | socket 타입·권한 | 오류는 `socket is invalid`로 뭉갠다 | AST calls, B2 |
| `net.Dialer.DialContext("unix", …)` | 요청 시점의 실제 연결 | client `Timeout` 5초. 재연결·백오프 없음 | AST calls |

## State mutations and fallbacks

- 없다. 조회 전용이고 디스크를 바꾸지 않는다.
- fallback도 없다 — 실패는 호출부로 올라가고, 그것을 **강등으로 받을지 fatal로 받을지**가
  겹3의 결정이다(design D4, T2 소관).

## Safety conclusion

- Safe edit boundary: 무변경. 이 함수가 느슨해지면 잘못된 socket에 토큰을 보내게 되므로,
  검사(0600·socket 타입·descriptor 경로 모양)를 완화하는 편집은 금지다.
- High-risk impact: no (조회 전용 클라이언트) — 다만 실패의 **처리 방식**은 High-risk다.
