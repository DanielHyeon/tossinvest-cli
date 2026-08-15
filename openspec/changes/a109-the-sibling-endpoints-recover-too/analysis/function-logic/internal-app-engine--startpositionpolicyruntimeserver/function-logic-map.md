# Function Logic Map: `StartPositionPolicyRuntimeServer`

- Source: `internal/app/engine/position_policy_runtime_transport_unix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**alert control과 같은 병의 다른 사본이다.** AST 분기 16개 중 발행 의례는
B9(`PrepareRuntimeSocket`)·B10(`net.Listen` 최종 경로)·B11(`os.Chmod` 0600)·
B12(`ValidateRuntimeSocket`)이고, B10과 B11 사이의 죽음이 최종 이름에 pre-chmod socket을
남긴다. B14·B15는 발행이 아니라 등록된 핸들러 본문의 분기다(같은 함수 리터럴 안에 있어
AST가 함께 센다) — a109는 그 둘을 건드리지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reader` | non-nil | 호출자 `cmd/tossctl/engine.go` | "has no reader" |
| `engineDir` | 공백 제거 후 비어 있지 않음 · 안전한 엔진 디렉터리 | `ValidateEngineDirectory` | error(오늘 호출부에서 fatal — T2가 강등) |
| control 디렉터리 | 없으면 0700 생성, 있으면 **오늘은 그대로 씀** | `os.Mkdir` + `ValidateRuntimeControlDirectory` | 잔재 회수 없음 → staging 축적 |
| 최종 socket 자리 | 없거나 정확-0600 socket | `PrepareRuntimeSocket` | pre-chmod 0700이면 영구 거부 · 산 주인이면 **탈취** |
| descriptor | stage(`.position-policy-runtime-*`)+rename | `writePositionPolicyRuntimeDescriptor` | 이미 원자적 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `reader == nil` | 없음 | "has no reader" | 배선 실수 거부 |
| B2 | `dir == ""` | 없음 | "has no engine directory" | 커버 없음 |
| B3 | `ValidateEngineDirectory(dir) != nil` | 없음 | 감싼 error | 안전하지 않은 엔진 디렉터리 |
| B4 | `os.Mkdir(controlDir, 0o700) != nil` | 디렉터리 생성 시도 | (B6로) | 기존 디렉터리 위 기동 |
| B5 | `else` | `createdControlDir = true` | — | 첫 기동 |
| B6 | `!errors.Is(err, os.ErrExist)` | 없음 | "creating position policy runtime directory" | 생성 불가 |
| B7 | `createdControlDir` (cleanup 클로저) | 실패 시 디렉터리 제거 | — | 남의 디렉터리 비삭제 |
| B8 | `ValidateRuntimeControlDirectory != nil` | cleanup | 감싼 error | 0700 아닌 디렉터리 |
| B9 | `PrepareRuntimeSocket(socketPath) != nil` | cleanup | "preparing position policy runtime socket" | **pre-chmod 0700 영구 거부** — a109가 없앤다 |
| B10 | `net.Listen("unix", socketPath) != nil` | cleanup | "binding" | **최종 경로 bind** — staged bind로 교체 |
| B11 | `os.Chmod(socketPath, 0o600) != nil` | cleanupListener | "securing" | B10과의 사이가 사고 창 |
| B12 | `ValidateRuntimeSocket != nil` | cleanupListener | "validating" | 발행 후 정확-0600 확인(유지) |
| B13 | `rand.Read` error | cleanupListener | "generating token" | 커버 없음 |
| B14 | (핸들러) `r.Method != http.MethodGet` | 없음 | 405 | 라우트 계약 |
| B15 | (핸들러) `reader.Runtime` error | 없음 | RPC error 매핑 | 라우트 계약 |
| B16 | `writePositionPolicyRuntimeDescriptor != nil` | cleanupListener | 그 error | descriptor 발행 실패 정리 |
| go(:124) | 정상 종단 | `server.server.Serve(listener)` goroutine | `(server, nil)` | 기동 후 실제 응답 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `positionpolicyrpc.ValidateEngineDirectory` | 부모 위생 | error 전파 | AST calls |
| `os.Mkdir` | control 디렉터리 | ErrExist만 관용 | AST calls |
| `positionpolicyrpc.ValidateRuntimeControlDirectory` | 이름 고정 + 0700 | error 전파 | AST calls |
| `positionpolicyrpc.PrepareRuntimeSocket` | 잔재 제거 | **probe 없음** — a109가 교체 | AST calls |
| `net.Listen` | socket bind(최종 경로) | error 전파 | AST calls |
| `os.Chmod` | 0600 | error 전파 | AST calls |
| `positionpolicyrpc.ValidateRuntimeSocket` | 발행 확인 | error 전파 | AST calls |
| `crypto/rand.Read` | 토큰 | error 전파 | AST calls |
| `writePositionPolicyRuntimeDescriptor` | descriptor stage+rename | error 전파 | AST calls |

## State mutations and fallbacks

- 디스크: control 디렉터리 · socket · descriptor. 실패 경로는 `cleanupControlDir`/
  `cleanupListener`가 자기가 만든 것만 되돌린다.
- 프로세스: `go Serve(listener)`.
- fallback 없음 — 모든 실패가 error이고 오늘 호출부에서 fatal이다.

## Safety conclusion

- Safe edit boundary: B9~B11 구간(회수·bind·chmod)만 교체. 핸들러 분기(B14·B15)·토큰·
  descriptor 발행·`go Serve` 순서는 불변.
- High-risk impact: yes — 엔진 기동 경로. 거부는 콘솔 조회 표면 상실, 오제거는 산 주인 탈취.
