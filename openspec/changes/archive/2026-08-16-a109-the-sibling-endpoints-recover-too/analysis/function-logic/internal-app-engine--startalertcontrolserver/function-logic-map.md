# Function Logic Map: `StartAlertControlServer`

- Source: `internal/app/engine/alert_control_transport_unix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**a109 T1이 편집한 두 발행 경로 중 하나다.** 편집 **후**의 AST 분기 10개다.

§2b.3 G9가 Mkdir → 회수 → 재Mkdir 세 절(옛 B4~B7)을 형제와 공유하는
`openReclaimedControlDirectory`로 옮겨 그 자리가 분기 하나(B4)가 됐다 — 동작·오류
문자열은 불변이고, 옮긴 절의 커버리지는 아래 표가 그 기계를 가리킨다.

편집 전 이 함수의 발행 의례는 B9(`PreparePrivateSocket`)·B10(`net.Listen` **최종 경로**)·
B11(`os.Chmod` 0600)·B12(`ValidatePrivateSocket`)의 넷이었고, **B10과 B11 사이의 죽음**이
최종 이름에 pre-chmod socket을 남겼다(design 병 표, a108 A1 F3). 편집 후에는 그 구간이
회수(B6 `ReclaimStalePrivateEndpoint`)와 staged listen(B9 `ListenStagedPrivateSocket`)
한 쌍이 되고, 발행 확인(B10 `ValidatePrivateSocket`, **정확-0600**)은 남는다 — 그 절이 곧
"완화가 회수 밖으로 새지 않았다"는 증거이기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ops` | non-nil | 호출자 `cmd/tossctl/engine.go` | "the alert control server has no operations" |
| `engineDir` | 공백 제거 후 비어 있지 않음 · 안전한 엔진 디렉터리 | `ValidateEngineDirectory` | error 반환(현재 호출부에서 **fatal** — T2가 강등) |
| control 디렉터리 | 없으면 생성(0700), 있으면 **회수 후 재생성** | `os.Mkdir` → `ReclaimStalePrivateEndpoint` → `os.Mkdir` | 회수가 거부하면 기동 실패(강등은 T2) |
| 최종 socket 자리 | 회수가 비운 자리 | `ReclaimStalePrivateEndpoint` | 산 주인이면 거부 · pre-chmod 0700 잔재는 회수 |
| descriptor | stage(`privateDescriptorStagingPrefix`)+rename | `publishPrivateDescriptor` | 이미 원자적 — 병 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `ops == nil` | 없음 | "has no operations" | 배선 실수 거부 |
| B2 | `dir == ""` | 없음 | "has no engine directory" | 커버 없음 |
| B3 | `ValidateEngineDirectory(dir) != nil` | 없음 | 감싼 error | 안전하지 않은 엔진 디렉터리 |
| B4 | `openReclaimedControlDirectory(...) != nil` | Mkdir · 회수(제거) · 재Mkdir | "creating/reclaiming/recreating the alert control directory" | **낯선 엔트리·산 주인 거부** / 잔재 회수 (옛 B4~B7이 이 기계 안이다) |
| B5 | `ValidatePrivateControlDirectory(controlDir) != nil` | cleanup | 감싼 error | 0700 아닌 디렉터리 |
| B6 | `ListenStagedPrivateSocket(...) != nil` | cleanup | "publishing the alert control socket" | **stage+rename 발행** — pre-chmod 상태가 최종 이름을 갖지 않는다 |
| B7 | `ValidatePrivateSocket(socketPath) != nil` | cleanupListener | "validating" | 발행 후 **정확-0600** 확인(완화 누출 감시) |
| B8 | `rand.Read` error | cleanupListener | "generating token" | 커버 없음 |
| B9 | `json.Marshal` error | cleanupListener | "encoding descriptor" | 도달 불가에 가깝다 |
| B10 | `publishPrivateDescriptor != nil` | cleanupListener | 그 error | descriptor 발행 실패 정리 |
| go(종단) | 정상 종단 | `server.server.Serve(listener)` goroutine | `(server, nil)` | 기동 후 실제 수락 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `positionpolicyrpc.ValidateEngineDirectory` | 부모 위생 | error 전파, 재시도 없음 | AST calls |
| `openReclaimedControlDirectory` | control 디렉터리 확보(Mkdir·회수·재Mkdir) | error 전파, 재시도 없음 | AST calls (a109 G9) |
| `positionpolicyrpc.ValidatePrivateControlDirectory` | leaf 0700 | error 전파 | AST calls |
| `alertControlEndpointNames` | 아는-이름 집합(회수 기계에 넘긴다) | 오류 없음 | AST calls |
| `positionpolicyrpc.ListenStagedPrivateSocket` | staged bind → 0600 → rename | error 전파 | AST calls |
| `positionpolicyrpc.ValidatePrivateSocket` | 발행 확인(정확-0600) | error 전파 | AST calls |
| `crypto/rand.Read` | 토큰 | error 전파 | AST calls |
| `publishPrivateDescriptor` | descriptor stage+rename | error 전파 | AST calls |

## State mutations and fallbacks

- 디스크: 회수(제거) · control 디렉터리 생성 · staged socket 생성/chmod/rename ·
  descriptor 발행. 실패 경로는 `cleanupControlDir`/`cleanupListener`가 되돌린다. 회수가
  끝난 뒤의 디렉터리는 **언제나 이번 기동이 만든 것**이므로 조건 없이 지워도 된다.
- 프로세스: `go Serve(listener)` — 반환 후에도 사는 유일한 상태.
- fallback 없음. 모든 실패는 error이고, 오늘 호출부에서 그 error는 fatal이다(T2가 강등).

## Safety conclusion

- Safe edit boundary: 편집한 것은 B4~B9 구간(회수·재생성·staged 발행)뿐이다. 토큰 생성·
  라우트·descriptor 발행·`go Serve` 순서는 건드리지 않았다 — 그 순서가 "socket이 있으면
  listener가 있었다"는 생존 판정의 전제다.
- High-risk impact: yes — 엔진 기동 경로. 거부는 운영자 ack 표면 상실, 오제거는 산 주인 탈취.
