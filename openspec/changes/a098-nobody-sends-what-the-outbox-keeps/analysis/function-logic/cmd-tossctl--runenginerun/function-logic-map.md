# Function Logic Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go` (**179-292**) — 편집 전 `179-280`
- AST evidence: `ast.json` — 편집 후 branches **18** / returns **14** / calls 51
  (편집 전 **16** / **12** / 48)
- Risk scan: `risk-pattern-report.md`
- source SHA-256: **`301f68c6…`**

## ⛔ 이 번들도 §5.2의 표에 **없던 자리**다 — 두 번째다

4.2 때 `engineRuntime`이 그랬고(옆 번들), 이번엔 그 **위**의 부팅 시퀀스다.
§5.2는 `cmd/tossctl` diff를 *"4.4의 운영자 표면"*으로 적었지만 **어느 함수인지는
안 적었고**, 표면을 실제로 띄우는 자리는 조립부가 아니라 **여기**다.

**두 번 다 `check_analysis`가 먼저 말했다** (*"missing evidence for modified
function cmd/tossctl/engine.go:runEngineRun"*). 같은 검사가 같은 파일에서 두 번
잡았다는 사실 자체가 §5.2의 「어느 함수」 칸이 비어 있다는 뜻이다 — 5.2를 고쳤다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx` | 조립이 끝난 `*engine.Context` | `engineAssemble` (B7) | 조립 실패는 즉시 반환 |
| `dir` | 엔진 디렉터리 — **group/other writable 이면 안 된다** | `positionpolicyrpc.ValidateEngineDirectory` | **B17** — 알림 소켓이 조립을 거절한다 |
| `ectx.Journal`·`Notifier`·`Entry` | 셋 다 non-nil | `engine.New` | **B16** — `AlertOperations`가 `ErrRuntimeUnavailable` |

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — **편집 후** 분기 18 · 이탈 14.
**a098이 더한 것은 B16·B17 둘뿐이고, B1~B15는 조건도 순서도 그대로다.**

| Branch | 위치 | Condition | Return |
|---|---|---|---|
| B1~B15 | `:181`~`:261` | 기존 부팅 시퀀스 — **a098이 한 자도 안 바꾼다** | 기존 |
| **B16 (a098 신설)** | `:266` | **`ectx.AlertOperations()` 오류** | `:267` |
| **B17 (a098 신설)** | `:274` | **`engine.StartAlertControlServer(dir, alertOps)` 오류** | `:275` |
| B18 | `:278` | 기존 — 편집 전 `:266`에 있던 갈래가 밀린 것 | `:279` |

> **⛔ 번호가 밀렸다. 옛 B16이 새 B18이다.**
> 조건은 안 바뀌었고 **id 만 움직였다.** 4.1의 `NewRuntime` 때와 같은 함정이고
> (옛 B10~B12 → B15~B17), 같은 이유로 여기 적는다: branch-test map 은 id 로
> 칸을 가리키므로 **조용히 다른 갈래를 가리키게 된다.**

**신설 둘 다 fail-closed 다.** 다른 선택지는 *"소켓을 못 열면 로그만 남기고
엔진은 띄운다"*였고 **안 골랐다** — 그러면 운영자가 승인할 수단이 없는 채로
엔진이 돌고, 진입이 잠기면 **푸는 방법이 재시작뿐**이 된다. design D7.1 이
*"「재시작해야 풀린다」는 운영 표면이 아니다"*라고 적는 그 상태다.

> **⚠ 그 대가는 적어 둔다.** 소켓을 못 여는 이유에는 *"엔진 디렉터리 권한이
> 느슨하다"*도 있고, 그때 **엔진 자체가 안 뜬다.** 손절을 도는 프로세스가
> 알림 표면 때문에 안 뜨는 것은 가벼운 대가가 아니다. 그래도 고른 이유는
> `ValidateEngineDirectory`가 거절하는 조건(group/other writable)이 **이미
> 다른 두 엔드포인트의 기동 조건**이기 때문이다 — `StartPositionPolicyCommandServer`
> (`:255`)와 `StartPositionPolicyRuntimeServer`(`:260`)가 **먼저** 같은 검사로
> 거절한다. 즉 이 갈래가 새로 만드는 「안 뜨는 경우」는 **없다.**

## Calls and live bindings

| Callee | Why | 오류 계약 |
|---|---|---|
| `engineRuntimeFactory` | 감독 루프 + 보조 실행자 조립 | 즉시 반환 |
| `engine.StartPositionPolicyCommandServer` | 기존 loopback TCP 표면 | 기존 갈래 |
| `engine.StartPositionPolicyRuntimeServer` | 기존 Unix 읽기 표면 | 기존 갈래 |
| `strategyprojectionrpc.Start` | 기존 전략 투영 표면 | 기존 갈래 |
| **`ectx.AlertOperations` (a098)** | **밀린 알림의 읽기·승인 의미** | **B16** |
| **`engine.StartAlertControlServer` (a098)** | **그 의미를 여는 소켓** | **B17** |
| `rt.Run` | 루프 | 반환값이 이 명령의 반환값 |

## State mutations and fallbacks

| Mutation | 무엇 | Fallback |
|---|---|---|
| **알림 control 디렉터리·소켓·descriptor 생성** | `<dir>/.alert-control/` 0700 · `alerts.sock` 0600 · `endpoint.json` 0600 | 실패하면 만든 것을 되감고 **거절**한다 |
| `defer alertControl.Close()` | 종료 시 셋 다 제거 | descriptor 가 엔진보다 오래 살면 안 듣는 소켓을 가리키는 토큰이 남는다 |

**순서가 중요하다** — 소켓은 `rt.Run` **앞**에서 뜬다. 뒤였으면 엔진이 도는
동안에도 한동안 승인 수단이 없다.

## Safety conclusion

- Safe edit boundary: 호출 둘 + 갈래 둘 + `defer` 하나. **B1~B15 는 안 건드린다.**
- High-risk impact: **yes** — 엔진 부팅 시퀀스다. 방향은 **보수적**이다
  (fail-closed 둘 추가, 기존 판정 0건 변경, 기존 표면 셋의 순서 불변).
- 되돌리기: 그 열두 줄을 지우면 오늘로 돌아간다 — **그리고 오늘이 「운영자가
  승인할 수단이 없다」다.**
