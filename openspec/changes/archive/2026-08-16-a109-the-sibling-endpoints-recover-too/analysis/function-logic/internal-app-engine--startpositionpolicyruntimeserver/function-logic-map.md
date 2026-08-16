# Function Logic Map: `StartPositionPolicyRuntimeServer`

- Source: `internal/app/engine/position_policy_runtime_transport_unix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**alert control과 같은 병의 다른 사본이었다.** 편집 **후**의 AST 분기 11개다.

§2b.3 G9가 Mkdir → 회수 → 재Mkdir 세 절(옛 B4~B7)을 형제와 공유하는
`openReclaimedControlDirectory`로 옮겨 그 자리가 분기 하나(B4)가 됐다 — 동작·오류
문자열은 불변이다.

편집 전 발행 의례는 B9(`PrepareRuntimeSocket`)·B10(`net.Listen` **최종 경로**)·
B11(`os.Chmod` 0600)·B12(`ValidateRuntimeSocket`)이었고, B10과 B11 사이의 죽음이 최종
이름에 pre-chmod socket을 남겼다. 편집 후에는 회수(B6)와 staged listen(B9) 한 쌍이 되고
발행 확인(B10, **정확-0600**)은 남는다.

B9·B10은 발행이 아니라 등록된 핸들러 본문의 분기다(같은 함수 리터럴 안에 있어 AST가
함께 센다) — a109는 그 둘을 건드리지 않았다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reader` | non-nil | 호출자 `cmd/tossctl/engine.go` | "has no reader" |
| `engineDir` | 공백 제거 후 비어 있지 않음 · 안전한 엔진 디렉터리 | `ValidateEngineDirectory` | error(오늘 호출부에서 fatal — T2가 강등) |
| control 디렉터리 | 없으면 0700 생성, 있으면 **회수 후 재생성** | `os.Mkdir` → `ReclaimStalePrivateEndpoint` → `os.Mkdir` | 회수가 거부하면 기동 실패(강등은 T2) |
| 최종 socket 자리 | 회수가 비운 자리 | `ReclaimStalePrivateEndpoint` | 산 주인이면 거부 · pre-chmod 0700 잔재는 회수 |
| descriptor | stage(`positionPolicyRuntimeStagingPrefix`)+rename | `writePositionPolicyRuntimeDescriptor` | 이미 원자적 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `reader == nil` | 없음 | "has no reader" | 배선 실수 거부 |
| B2 | `dir == ""` | 없음 | "has no engine directory" | 커버 없음 |
| B3 | `ValidateEngineDirectory(dir) != nil` | 없음 | 감싼 error | 안전하지 않은 엔진 디렉터리 |
| B4 | `openReclaimedControlDirectory(...) != nil` | Mkdir · 회수(제거) · 재Mkdir | "creating/reclaiming/recreating position policy runtime directory" | **낯선 엔트리·산 주인 거부** / 잔재 회수 (옛 B4~B7이 이 기계 안이다) |
| B5 | `ValidateRuntimeControlDirectory != nil` | cleanup | 감싼 error | 0700 아닌 디렉터리 |
| B6 | `ListenStagedPrivateSocket(...) != nil` | cleanup | "publishing position policy runtime socket" | **stage+rename 발행** |
| B7 | `ValidateRuntimeSocket != nil` | cleanupListener | "validating" | 발행 후 **정확-0600** 확인 |
| B8 | `rand.Read` error | cleanupListener | "generating token" | 커버 없음 |
| B9 | (핸들러) `r.Method != http.MethodGet` | 없음 | 405 | 라우트 계약 |
| B10 | (핸들러) `reader.Runtime` error | 없음 | RPC error 매핑 | 라우트 계약 |
| B11 | `writePositionPolicyRuntimeDescriptor != nil` | cleanupListener | 그 error | descriptor 발행 실패 정리 |
| go(종단) | 정상 종단 | `server.server.Serve(listener)` goroutine | `(server, nil)` | 기동 후 실제 응답 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `positionpolicyrpc.ValidateEngineDirectory` | 부모 위생 | error 전파 | AST calls |
| `openReclaimedControlDirectory` | control 디렉터리 확보(Mkdir·회수·재Mkdir) | error 전파, 재시도 없음 | AST calls (a109 G9) |
| `positionpolicyrpc.ValidateRuntimeControlDirectory` | 이름 고정 + 0700 | error 전파 | AST calls |
| `positionPolicyRuntimeEndpointNames` | 아는-이름 집합 | 오류 없음 | AST calls |
| `positionpolicyrpc.ListenStagedPrivateSocket` | staged bind → 0600 → rename | error 전파 | AST calls |
| `positionpolicyrpc.ValidateRuntimeSocket` | 발행 확인(정확-0600) | error 전파 | AST calls |
| `crypto/rand.Read` | 토큰 | error 전파 | AST calls |
| `writePositionPolicyRuntimeDescriptor` | descriptor stage+rename | error 전파 | AST calls |

## State mutations and fallbacks

- 디스크: 회수(제거) · control 디렉터리 · staged socket · descriptor. 실패 경로는
  `cleanupControlDir`/`cleanupListener`가 되돌린다. 회수 뒤의 디렉터리는 언제나 이번
  기동이 만든 것이므로 조건 없이 지워도 된다.
- 프로세스: `go Serve(listener)`.
- fallback 없음 — 모든 실패가 error이고 오늘 호출부에서 fatal이다.

## Safety conclusion

- Safe edit boundary: 편집한 것은 B4~B9 구간(회수·재생성·staged 발행)뿐이다. 핸들러
  분기(B9·B10)·토큰·descriptor 발행·`go Serve` 순서는 불변.
- High-risk impact: yes — 엔진 기동 경로. 거부는 콘솔 조회 표면 상실, 오제거는 산 주인 탈취.
