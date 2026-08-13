# Function Logic Map: `Start`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (34-116)
- AST evidence: `ast.json` — **revision `base`** (비교 기준 커밋 `75df9bf9`의 파일). AST 분기 15
- Risk scan: `risk-pattern-report.md`

## 왜 base revision인가 (숨기지 않고 적는다)

**이 change는 `Start`의 본문을 한 줄도 바꾸지 않았다.** Pre-Edit 선언 T1-3이 그것을
무변경으로 못박았고, 근거는 "listen → chmod → token → descriptor" 순서가 생존 판정의
전제(design D2)라 건드리면 그 순서 자체가 리뷰 대상이 된다는 것이다.

그런데 바로 아래 함수 `reclaimStaleControlDirectory`에 doc comment를 붙이면서 삽입 지점이
base 파일의 117행이 되었고, `check_analysis.py`의 인접 규칙(`start <= line <= end+1`)이
그것을 "`Start`(34-116)에 닿는 편집"으로 읽는다. 그 규칙은 함수 끝에 덧붙이는 편집을
잡으려고 있는 것이고, 여기서는 위양성이다. 규칙을 피하려고 diff 모양을 만들지 않고
base revision 증거를 정직하게 남긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reader` | non-nil | 호출부(`internal/app/engine`의 projection 주입) | B1 거부 |
| `engineDir` | 실디렉터리 · symlink 아님 · group/other 쓰기 비트 없음 | `os.Lstat` | B2 `engine directory is unsafe` |
| `.strategy-runtime-read/` | 없어야 한다. 있으면 회수 대상 | `os.Mkdir`의 `ErrExist` | B3-B5: 회수 실패는 그대로 기동 실패 |
| socket bind | listen 성공이 descriptor 발행의 **선행 조건** | `net.Listen` | B7 거부 + 디렉터리 되돌림 |
| token | 32바이트 난수 | `crypto/rand` | B9 거부 + listener 되돌림 |

**이 함수가 회수 판정에 주는 전제**: descriptor는 listen·chmod·token이 전부 성공한 뒤에만
쓰인다(B7→B8→B9→B15 순). 그래서 "수락하지 않는 socket 파일"은 죽은 주인의 것이라고
말할 수 있다 — design D2의 근거가 이 순서다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `reader == nil` | 없음 | `reader is required` | 없음(무변경·기존 무테스트) |
| B2 | 엔진 디렉터리가 안전하지 않다 | 없음 | `engine directory is unsafe` | 없음(무변경) |
| B3 | `os.Mkdir` 실패 | 없음 | — | `TestStartRecoversFromDescriptorOnlyLeftover` |
| B4 | 실패가 `ErrExist`가 아니다 | 없음 | `create control directory: %w` | 없음(무변경) |
| B5 | **회수가 거부했다** | 없음 | 회수의 오류 그대로 | `TestStartRefusesControlDirectoryWithUnknownEntry` |
| B6 | 회수 후 재`Mkdir` 실패 | 없음 | `recreate control directory: %w` | 없음(무변경) |
| B7 | `net.Listen` 실패 | 디렉터리 제거 | 원본 오류 | 없음(무변경) |
| B8 | `os.Chmod` 실패 | listener·socket·디렉터리 제거 | 원본 오류 | 없음(무변경) |
| B9 | `rand.Read` 실패 | 위와 같음 | 원본 오류 | 없음(무변경) |
| B10 | 토큰 불일치 | 없음 | HTTP 401 | `TestUnixEndpointStrictMethodBodyQueryAuthAndRouteGuards` |
| B11 | GET·HEAD가 아니다 | 없음 | HTTP 405 | 같은 테스트 |
| B12 | 본문·쿼리가 있다 | 없음 | HTTP 400 | 같은 테스트 |
| B13 | 스냅샷 읽기·검증 실패 | 없음 | HTTP 503 | 없음(무변경) |
| B14 | HEAD 요청 | 없음 | 헤더만 200 | 없음(무변경) |
| B15 | descriptor 발행 실패 | listener·socket·디렉터리 제거 | 원본 오류 | 없음(무변경) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reclaimStaleControlDirectory` | `ErrExist`일 때 잔재 회수 | 오류는 **그대로 기동 실패**가 된다 — 이 한 줄이 8/13 사고의 전파 경로다 | AST calls(48행), B5 |
| `net.Listen("unix", …)` | 소켓 bind | 실패 시 디렉터리 되돌림 | AST calls |
| `writeDescriptor` | 토큰·PID 발행 (O_EXCL) | 실패 시 listener까지 되돌림 | AST calls, B15 |
| `strategyprojection.Validate/Clone` | 응답 경계 | 실패는 503 | AST calls |

## State mutations and fallbacks

- 되돌림은 3단계(`cleanupDir` → `cleanupListener`)로 계단식이며, 각 실패 지점이 자기
  앞까지 만든 것을 지운다. fallback(강등 기동)은 이 함수에 없다 — 강등은 호출부
  `runEngineRun`의 몫이고 그것이 겹2(T2 소관)다.
- 이 change로 바뀐 것은 **B5가 얼마나 자주 발동하는가**뿐이다. 코드는 그대로다.

## Safety conclusion

- Safe edit boundary: 무변경. 편집이 필요하면 B7→B15의 순서(listen이 descriptor보다 앞)를
  깨지 않는 범위여야 한다 — 그 순서가 D2 생존 판정의 전제다.
- High-risk impact: yes (엔진 기동 경로) — 다만 이번 판에서는 편집하지 않았다.
