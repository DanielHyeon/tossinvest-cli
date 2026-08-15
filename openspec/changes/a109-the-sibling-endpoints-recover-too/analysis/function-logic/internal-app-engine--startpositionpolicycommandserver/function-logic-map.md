# Function Logic Map: `StartPositionPolicyCommandServer`

- Source: `internal/app/engine/position_policy_transport.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**socket이 없는 endpoint다** — loopback TCP(`net.Listen("tcp4", "127.0.0.1:0")`, B10)이므로
pre-chmod 병도, 산 주인 탈취도 없다(probe할 대상 자체가 없다). 남는 병은 하나뿐이다:
`os.CreateTemp(dir, ".position-policy-control-*")`가 crash 시 남기는 **staging 잔재를
아무도 치우지 않는다.** 어떤 검증도 디렉터리를 열거하지 않으므로 조용히 쌓인다.

a109는 여기에 **staging 위생만** 더했다(design D2a — B9와 B10 사이의 문장 하나,
`SweepPrivateStagingLeftovers`). 열거+낯선-엔트리 거부는 넣지 않았다 — 이물 하나가 격리
해제 표면을 매 부팅 지우는 새 실패 경로가 되기 때문이다(freeze P1-2). 낯선 엔트리는
전과 같이 무시한다. 분기 수는 편집 전후 모두 16이다(위생은 분기가 아니라 문장이다).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `commands` | non-nil | 호출자 | "has no service" |
| `engineDir` | 공백 제거 후 비어 있지 않음 · 안전한 엔진 디렉터리 | `ValidateEngineDirectory` | error(오늘 fatal — T2가 강등) |
| control 디렉터리 | 없으면 0700 생성, 있으면 그대로 | `os.Mkdir` + `ValidateControlDirectory` | staging 잔재 무한 축적 |
| listener | `127.0.0.1:0` (ephemeral) | `net.Listen` | error |
| 격리 해제 라우트 | `commands`가 `exitQuarantineCommands`면 등록 | 타입 단언(B15) | 능력 없으면 라우트 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `commands == nil` | 없음 | "has no service" | 배선 실수 거부 |
| B2 | `dir == ""` | 없음 | "has no engine directory" | 커버 없음 |
| B3 | `ValidateEngineDirectory(dir) != nil` | 없음 | 감싼 error | group-writable 엔진 디렉터리 거부 |
| B4 | `os.Mkdir(controlDir, 0o700) != nil` | 디렉터리 생성 시도 | (B6로) | 기존 디렉터리 위 기동 |
| B5 | `else` | `createdControlDir = true` | — | 첫 기동 |
| B6 | `!errors.Is(err, os.ErrExist)` | 없음 | "creating position policy control directory" | 생성 불가 |
| B7 | `ValidateControlDirectory(controlDir) != nil` | (B8) | 감싼 error | 0750 디렉터리 거부 |
| B8 | `createdControlDir` (B7 본문 안) | 우리가 만든 디렉터리만 제거 | — | 남의 디렉터리 비삭제 |
| B9 | `createdControlDir` (cleanup 클로저) | 같음 | — | 같음 |
| B10 | `net.Listen("tcp4", "127.0.0.1:0") != nil` | cleanup | "binding position policy control" | 커버 없음 |
| B11 | `rand.Read` error | listener close + cleanup | "generating token" | 커버 없음 |
| B12 | (핸들러) `/v1/health` 메서드 != GET | 없음 | 405 | 라우트 계약 |
| B13 | (핸들러) `/v1/positions` 메서드 != GET | 없음 | 405 | 라우트 계약 |
| B14 | (핸들러) `commands.List` error | 없음 | RPC error 매핑 | 라우트 계약 |
| B15 | `commands.(exitQuarantineCommands)` ok | 격리 해제 라우트 등록 | — | 능력 유무로 라우트가 갈린다 |
| B16 | `writePositionPolicyDescriptor != nil` | listener close + cleanup | 그 error | descriptor 발행 실패 정리 |
| go(:132) | 정상 종단 | `server.server.Serve(listener)` goroutine | `(server, nil)` | 기동 후 응답 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `positionpolicyrpc.ValidateEngineDirectory` | 부모 위생 | error 전파 | AST calls |
| `os.Mkdir` | control 디렉터리 | ErrExist만 관용 | AST calls |
| `positionpolicyrpc.ValidateControlDirectory` | 이름 고정 + 0700 | error 전파 | AST calls |
| `net.Listen("tcp4")` | loopback ephemeral | error 전파 | AST calls |
| `crypto/rand.Read` | 토큰 | error 전파 | AST calls |
| `registerExitQuarantineRoutes` | 능력이 있을 때만 | — | AST calls |
| `writePositionPolicyDescriptor` | descriptor stage+rename | error 전파 | AST calls |

## State mutations and fallbacks

- 디스크: control 디렉터리 · descriptor. **socket 없음** — 이 endpoint의 잔재는
  디렉터리·descriptor·staging 파일뿐이다.
- 프로세스: `go Serve(listener)`.
- descriptor 잔재는 rename이 덮으므로 기동을 막지 않는다(기동은 이전 descriptor를 읽지 않는다).

## Safety conclusion

- Safe edit boundary: `ValidateControlDirectory` 통과 직후에 **staging 위생 한 문장**만
  더했다. 낯선 엔트리 열거·거부는 넣지 않았다(D2a). 라우트·토큰·발행 순서 불변.
- High-risk impact: yes — 이 표면이 없으면 **격리 해제**(격리된 포지션의 손절 포함 미판정
  상태를 푸는 유일한 장중 경로)가 사라진다. 그래서 새 실패 경로를 만들지 않는 것이 요구다.
