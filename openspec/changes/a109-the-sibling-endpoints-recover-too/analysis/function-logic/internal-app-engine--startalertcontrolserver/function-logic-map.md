# Function Logic Map: `StartAlertControlServer`

- Source: `internal/app/engine/alert_control_transport_unix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**a109 T1이 편집하는 두 발행 경로 중 하나다.** AST 분기 15개 중 발행 의례에 해당하는
것은 B9(`PreparePrivateSocket`)·B10(`net.Listen` 최종 경로)·B11(`os.Chmod` 0600)·
B12(`ValidatePrivateSocket`)의 넷이고, 그 **B10과 B11 사이의 죽음**이 최종 이름에
pre-chmod socket을 남긴다(design 병 표). a109는 B9~B11을 회수(`ReclaimStalePrivateEndpoint`)
+ staged listen(`ListenStagedPrivateSocket`) 한 쌍으로 바꾸고 B12는 남긴다 — 발행이
0600임을 확인하는 절이 곧 "완화가 회수 밖으로 새지 않았다"는 증거이기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ops` | non-nil | 호출자 `cmd/tossctl/engine.go` | "the alert control server has no operations" |
| `engineDir` | 공백 제거 후 비어 있지 않음 · 안전한 엔진 디렉터리 | `ValidateEngineDirectory` | error 반환(현재 호출부에서 **fatal** — T2가 강등) |
| control 디렉터리 | 없으면 생성(0700), 있으면 **오늘은 그대로 씀** | `os.Mkdir` + `ValidatePrivateControlDirectory` | 잔재 회수 없음 → staging 무한 축적(design 병 표) |
| 최종 socket 자리 | 없거나 정확-0600 socket | `PreparePrivateSocket` | pre-chmod 0700이면 영구 거부 · 산 주인이면 **탈취** |
| descriptor | stage(`.endpoint-*`)+rename | `publishPrivateDescriptor` | 이미 원자적 — 병 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `ops == nil` | 없음 | "has no operations" | 배선 실수 거부 |
| B2 | `dir == ""` | 없음 | "has no engine directory" | 커버 없음 |
| B3 | `ValidateEngineDirectory(dir) != nil` | 없음 | 감싼 error | 안전하지 않은 엔진 디렉터리 |
| B4 | `os.Mkdir(controlDir, 0o700) != nil` | 디렉터리 생성 시도 | (B6로) | 기존 디렉터리 위에서의 기동 |
| B5 | `else` (Mkdir 성공) | `createdControlDir = true` | — | 첫 기동 |
| B6 | `!errors.Is(err, os.ErrExist)` | 없음 | "creating the alert control directory" | 생성 불가 |
| B7 | `createdControlDir` (cleanup 클로저 안) | 실패 시 디렉터리 제거 | — | 실패 정리가 남의 디렉터리를 지우지 않음 |
| B8 | `ValidatePrivateControlDirectory(controlDir) != nil` | cleanup | 감싼 error | 0700 아닌 디렉터리 |
| B9 | `PreparePrivateSocket(socketPath) != nil` | cleanup | "preparing the alert control socket" | **pre-chmod 0700 영구 거부(A1 F3)** — a109가 없앤다 |
| B10 | `net.Listen("unix", socketPath) != nil` | cleanup | "binding" | **최종 경로 bind** — a109가 staged bind로 바꾼다 |
| B11 | `os.Chmod(socketPath, 0o600) != nil` | cleanupListener | "securing" | B10과의 사이가 사고 창 |
| B12 | `ValidatePrivateSocket(socketPath) != nil` | cleanupListener | "validating" | 발행 후 정확-0600 확인(유지) |
| B13 | `rand.Read` error | cleanupListener | "generating token" | 커버 없음 |
| B14 | `json.Marshal` error | cleanupListener | "encoding descriptor" | 도달 불가에 가깝다 |
| B15 | `publishPrivateDescriptor != nil` | cleanupListener | 그 error | descriptor 발행 실패 정리 |
| go(:151) | 정상 종단 | `server.server.Serve(listener)` goroutine | `(server, nil)` | 기동 후 실제 수락 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `positionpolicyrpc.ValidateEngineDirectory` | 부모 위생 | error 전파, 재시도 없음 | AST calls |
| `os.Mkdir` | control 디렉터리 | ErrExist만 관용 | AST calls |
| `positionpolicyrpc.ValidatePrivateControlDirectory` | leaf 0700 | error 전파 | AST calls |
| `positionpolicyrpc.PreparePrivateSocket` | 잔재 제거 | **probe 없음** — a109가 교체 | AST calls |
| `net.Listen` | socket bind(최종 경로) | error 전파 | AST calls |
| `os.Chmod` | 0600 | error 전파 | AST calls |
| `positionpolicyrpc.ValidatePrivateSocket` | 발행 확인 | error 전파 | AST calls |
| `crypto/rand.Read` | 토큰 | error 전파 | AST calls |
| `publishPrivateDescriptor` | descriptor stage+rename | error 전파 | AST calls |

## State mutations and fallbacks

- 디스크: control 디렉터리 생성 · socket 생성/chmod · descriptor 발행. 실패 경로는
  `cleanupControlDir`/`cleanupListener`가 자기가 만든 것만 되돌린다.
- 프로세스: `go Serve(listener)` — 반환 후에도 사는 유일한 상태.
- fallback 없음. 모든 실패는 error이고, 오늘 호출부에서 그 error는 fatal이다(T2가 강등).

## Safety conclusion

- Safe edit boundary: B9~B11(회수·bind·chmod) 구간만 교체한다. 토큰 생성·라우트·descriptor
  발행·`go Serve` 순서는 건드리지 않는다 — 그 순서가 "socket이 있으면 listener가 있었다"는
  생존 판정의 전제다.
- High-risk impact: yes — 엔진 기동 경로. 거부는 운영자 ack 표면 상실, 오제거는 산 주인 탈취.
