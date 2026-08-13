# Function Logic Map: `Dial`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (334-356)
- AST evidence: `ast.json` — revision `current`. AST 분기 3
- Risk scan: `risk-pattern-report.md`

첫 라운드는 이 함수를 바꾸지 않았다. Fix 라운드가 connect probe를 넣는다(design D4-2,
A2 F3). 분기는 2 → 3으로 늘었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `descriptorPath` | `…/.strategy-runtime-read/endpoint.json` | 호출부(콘솔·httpapi) | `readDescriptor`가 경로 모양부터 검사 |
| descriptor 내용 | 스키마 v1 · socket 이름 · 토큰 32자 이상 · PID > 0 | 디스크 | B1 — 오류 그대로 전파 |
| socket 파일 | socket 타입 · perm 정확히 0600 | 디스크 | B2 `socket is invalid` |
| socket **주인** | 지금 수락해야 한다 | `projectionSocketAccepts` | B3 `socket has no listener` |
| `ctx` | **무시된다** | — | 첫 인자는 `_`다. 타임아웃은 probe 200ms와 client 5초 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `readDescriptor` 실패 | 없음 | 원본 오류 | `TestStartRecoversFromDescriptorOnlyLeftover` (성공 쪽) |
| B2 | socket lstat 실패 · socket 아님 · perm != 0600 | 없음 | `socket is invalid` | `TestCloseToleratesLeftoverAlreadyRemoved` |
| B3 | 아무도 수락하지 않는다 | 없음 | `socket has no listener` | `TestDialRefusesSocketWithNoListener` |

## 왜 여기서 연결해 보는가 (design D4-2)

첫 라운드의 `Dial`은 **연결하지 않았다** — transport를 만들어 돌려줄 뿐이고 실제 연결은
첫 요청에서 일어났다. A2가 실측한 결과가 그 대가다: 전원이 끊기면 unlink할 기회가 없어
socket 파일이 그대로 남고(design D1의 S3, **전원 단절의 기본 모양**), 그 위에서 `Dial`은
성공한다. 실패는 첫 `Read`에서야 나타나고, 그때는 소비자가 이미 "붙었다"고 판단한
뒤라 강등할 기회를 놓친다 — httpapi에서는 집계 스냅샷 전체(engine·positions·orders)가
전략 하나 때문에 함께 죽었다.

판정 원시는 회수와 같은 것을 쓴다(`projectionSocketAccepts`). 그 함수는 **연결 거부와
파일 부재만** 사망으로 읽고 나머지(권한·타임아웃)는 살아 있다고 본다. `Dial`에서 그
보수성의 방향은 "애매하면 client를 준다"이고, 그 경우의 실패는 예전처럼 첫 `Read`가
받는다 — 즉 이 분기는 **확실한 사망에서만** 발동한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readDescriptor` | 토큰·socket 이름 획득 + 파일 검증 | 오류 그대로 전파 | AST calls, B1 |
| `os.Lstat` | socket 타입·권한 | 오류는 `socket is invalid`로 뭉갠다 | AST calls, B2 |
| `projectionSocketAccepts` | **주인 생존 확인** | 200ms 1회. 재시도 없음 | AST calls, B3 |
| `net.Dialer.DialContext("unix", …)` | 요청 시점의 실제 연결 | client `Timeout` 5초. 재연결·백오프 없음 | AST calls |

## State mutations and fallbacks

- 없다. 조회 전용이고 디스크를 바꾸지 않는다. probe가 여는 연결은 즉시 닫는다.
- fallback도 없다 — 실패는 호출부로 올라가고, 그것을 강등으로 받는 것이 콘솔·httpapi의
  결정이다(design D4-2, T2 소관). 두 호출부 모두 dial 오류를 dormant 강등으로 받는다.

## Safety conclusion

- Safe edit boundary: 검사(0600·socket 타입·descriptor 경로 모양·probe)를 완화하는 편집은
  금지다. 느슨해지면 잘못된 socket에 토큰을 보내게 된다.
- High-risk impact: no (조회 전용 클라이언트) — 다만 실패의 **처리 방식**은 High-risk다.
  probe 추가는 실패를 **더 이르게** 만들 뿐 더 치명적으로 만들지 않는다.
