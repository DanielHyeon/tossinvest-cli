# Function Logic Map: `StartPositionPolicyCommandServer`

- Source: `internal/app/engine/position_policy_transport.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a063-operator-can-lift-a-quarantine/base-commit.txt`
- 위험 등급: **High-risk** (원장 write를 노출하는 인증 endpoint). Pre-Edit 선언은 `review.md`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `engineDir string` | 비어 있지 않고 `ValidateEngineDirectory` 통과 | `cmd/tossctl/engine.go:254` | 거부, listener 미생성 |
| `commands positionPolicyCommands` | non-nil | `NewPositionPolicyCommandService` | 거부 |
| control dir | `0700`, engine dir 하위 | 파일시스템 | 생성 실패·검증 실패 모두 거부 |
| bearer token | `crypto/rand` 32바이트 | 프로세스 메모리 | 생성 실패 시 listener 닫고 거부 |

**호출부**: `cmd/tossctl/engine.go:254` 한 곳뿐 (repo 전역 grep, 현재 HEAD).

**a063이 추가하는 것**: `commands`가 선택적 격리 capability를 함께 구현하면
`/v1/quarantines`, `/v1/quarantine/preview`, `/v1/quarantine/release` 세 라우트를
같은 `server.auth(token, …)`로 추가 등록한다. **함수 시그니처와 호출부는 바뀌지
않는다** — 배선 지점(`runEngineRun`, 36분기의 `runConsole`)을 건드리지 않기 위한
의도적 선택이며, `http.Flusher`류의 선택적 capability 탐지와 같은 형태다.

## Branches and early returns

번호는 수정 **전** AST 기준이며 a063은 어떤 분기도 제거·재배치하지 않는다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `commands == nil` | 없음 | error, 서버 없음 | 기존 transport 테스트 |
| B2 | `dir == ""` | 없음 | error | 기존 |
| B3 | `ValidateEngineDirectory` 실패 | 없음 | error | 기존 |
| B4 | `os.Mkdir(controlDir)` 오류 | 없음 | 아래 두 갈래 | 기존 |
| B5 | 오류가 `ErrExist`가 아님 | 없음 | error | 기존 |
| B6 | Mkdir 성공 | `createdControlDir = true` | 계속 | 기존 |
| B7 | `ValidateControlDirectory` 실패 | 아래 | error | 기존 |
| B8 | 실패 시 `createdControlDir` | control dir 삭제 | error | 기존 |
| B9 | `cleanupControlDir` 안의 `createdControlDir` | control dir 삭제 | — | 기존 |
| B10 | `net.Listen` 오류 | control dir 정리 | error | 기존 |
| B11 | `rand.Read` 오류 | listener 닫기 + 정리 | error | 기존 |
| B12 | `/v1/health` non-GET | 없음 | 405 | 기존 |
| B13 | `/v1/positions` non-GET | 없음 | 405 | 기존 |
| B14 | `commands.List` 오류 | 없음 | RPC error | 기존 |
| B15 | `writePositionPolicyDescriptor` 오류 | listener 닫기 + 정리 | error | 기존 |

**a063이 추가하는 분기**: B16 — `commands`가 격리 capability를 구현하는가
(타입 단언). 참이면 라우트 3개를 추가 등록하고, 거짓이면 **현재와 완전히 동일한
mux**가 만들어진다. 이것이 §0.2(OFF = 기존 동작)를 이 함수에서 만족시키는 방법이다.
B17~B19는 새 라우트 각각의 method 검사다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `positionpolicyrpc.ValidateEngineDirectory` | 경로 신뢰 | 오류면 즉시 거부 | AST |
| `positionpolicyrpc.ValidateControlDirectory` | 0700 사설 디렉터리 확인 | 같음 | AST |
| `net.Listen("tcp4","127.0.0.1:0")` | loopback 전용 바인드 | 오류면 거부 | AST |
| `rand.Read` | bearer token | 오류면 거부 | AST |
| `server.auth(token, …)` | 모든 라우트 인증 | 불일치는 401 | AST |
| `commands.List/Preview/Apply` | 정책 명령 위임 | 오류는 `writePositionPolicyRPCError` | AST |
| `writePositionPolicyDescriptor` | 사설 descriptor 발행 | 오류면 전부 롤백 | AST |
| **(a063)** `quarantines.List/Preview/Release` | 격리 명령 위임 | 오류는 전용 RPC error 매핑 | 신규 |

## State mutations and fallbacks

- 파일시스템: control dir 생성(실패 시 되돌림), descriptor 파일 기록.
- 네트워크: loopback listener 하나, goroutine 하나.
- a063은 mux에 핸들러를 **추가**할 뿐 기존 라우트의 인증·메서드·오류 매핑을 바꾸지
  않는다. 새 라우트도 같은 `server.auth(token, …)`를 통과한다 — 인증 없는 표면을
  만들지 않는다.

## Safety conclusion

- Safe edit boundary: mux 등록 구간. listener·token·descriptor·정리 경로는 불변.
- High-risk impact: **yes** — 이 endpoint 뒤에 원장 write가 있다. 새 라우트도 기존과
  같은 bearer 인증·사설 descriptor·loopback 바인드 안에 들어간다.
- §0.2: 격리 capability가 없으면 mux는 현재와 같은 라우트 집합을 갖는다.
- §0.7: 새 라우트는 사람이 콘솔에서 승인한 capability만 소비한다. 자동 호출자가 없다.
