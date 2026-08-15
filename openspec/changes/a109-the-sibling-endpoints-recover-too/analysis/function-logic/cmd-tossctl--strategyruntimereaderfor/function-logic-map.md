# Function Logic Map: `strategyRuntimeReaderFor`

- Source: `cmd/tossctl/httpapi.go` (254-286)
- AST evidence: `ast.json` — AST 분기 4 · return 4 · defer 0
  (source_sha256 `37c2ed8990337096cb98973cba78f1d0e9cc61abade8df115106d754761ef2df`,
  a109 base `016da624`)
- Risk scan: `risk-pattern-report.md`
- **a109 T2 편집 대상: 함수 전체의 반환 계약** — 오늘 이 함수는 부팅 **1회**의 판정을
  값 하나로 굳힌다(nil · sentinel · live client). a109 D4 는 그 세 값을 감싸는
  wrapper 를 돌려주게 바꾸고, 재부착 시도를 wrapper 가 백그라운드에서 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 데몬의 부팅 ctx | `runHTTPAPI` | 오늘 `Dial` 에 그대로 넘긴다 — 재부착 후에는 **부팅 ctx 를 시도 goroutine 이 상속하면 안 된다**(종료 시 취소) |
| `root.configDir` | 비면 XDG 기본 | `engineJournalDir` | 해석 실패 → 경고 한 줄 + **nil** (B1) = 「이 배포는 전략 화면을 안 쓴다」 |
| descriptor 파일 | 있으면 엔진이 발행했다 | `strategyprojectionrpc.DescriptorPath(dir)` | 부재(NotExist) → 조용히 nil (B2, B3 아님). 비-NotExist stat 오류 → 경고 + nil (B2·B3) |
| `Dial` 결과 | 성공하면 live client | `strategyprojectionrpc.Dial` | 실패 → 경고 + **sentinel**(`unavailableStrategyRuntime`) (B4) |
| **화면 값의 구분(a108 D4-2)** | nil = dormant(NOT_CONFIGURED) · non-nil 인데 Read 실패 = unavailable(RUNTIME_UNAVAILABLE) | `httpAPIReader.Snapshot` (:565–573) · `internal/httpapi/strategy_runtime.go` (:18) | **접으면 운영자는 「기능을 안 켰다」와 「엔진이 죽었다」를 구별할 수 없다** — a109 wrapper 는 정의상 non-nil 이므로 이 구분을 별도 상태 신호로 옮겨야 한다(freeze P1-4) |

**불변식 1**: 이 데몬은 전략 endpoint 때문에 죽지 않는다(a108 D4-2). 어떤 분기도
오류를 반환하지 않는다 — 반환 타입에 error 가 없다는 것이 그 계약의 형태다.

**불변식 2 (a109 가 더한다)**: 요청 goroutine 은 `Dial` 을 부르지 않는다.
`strategyprojectionrpc.Dial` 본문에 200ms connect probe 가 있으므로(transport_unix.go
:417 `projectionSocketAccepts`, `projectionProbeTimeout`) 요청 경로 동기 호출은
「요청 경로 비차단」 SHALL NOT 을 구현 시점에 거짓으로 만든다(freeze P0-2).

**불변식 3**: 이 데몬은 남의 control 디렉터리를 치우지 않는다 — 잔재 정리는 엔진의
일이다(`TestAnAbsentDescriptorAndADeadOneBootTheSame` 가 핀).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (:257) | `engineJournalDir` 실패 | stderr 경고 | **nil** (:262) → dormant | 미고정 (경로 해석 실패) |
| B2 (:265) | descriptor `os.Stat` 실패 | (B3 에 따라) | **nil** (:273) → dormant | `TestADialFailureRendersUnavailableRatherThanNotConfigured`("descriptor 가 없다" 행) |
| B3 (:266) | 그 실패가 **NotExist 가 아니다** | stderr 경고 (환경 이상 문구) | 없음 — B2 의 nil 로 떨어진다 | `TestAnUninspectableDescriptorDegradesLikeTheConsole` |
| B4 (:276) | `Dial` 실패 | stderr 경고 | **sentinel** `unavailableStrategyRuntime{cause}` (:283) → unavailable | `TestADeadDescriptorDoesNotStopTheDaemon` · `TestASocketFileWithNoOwnerDegradesTheDaemon` · `TestADialFailureRendersUnavailableRatherThanNotConfigured` |
| — (:285) | 위 넷을 다 지났다 | 없음 | **live `*strategyprojectionrpc.Client`** | `TestAnAbsentDescriptorAndADeadOneBootTheSame`(대조군) |

**a109 가 여는 구멍**: 위 표의 어느 값도 **다시 시도하지 않는다.** nil·sentinel 은
영원히 그 값이고, live client 는 엔진 재시작(새 socket·새 토큰) 후 영원히 실패한다
(freeze P0-1). 그래서 a109 의 재부착 트리거는 「부재 **또는 직전 Read 실패**」이고,
wrapper 가 **live 를 포함한 세 값 전부**를 감싼다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineJournalDir` | 이 프로파일의 엔진 디렉터리 | 실패는 nil + 경고 (B1) | AST · httpapi.go:256 |
| `strategyprojectionrpc.DescriptorPath` | descriptor 경로 계산 | 순수 함수 | AST · httpapi.go:264 |
| `os.Stat` | descriptor 존재 조사 | NotExist 와 그 밖을 가른다 (B2·B3) | AST · httpapi.go:265 |
| `strategyprojectionrpc.Dial` | live client 확보 | **본문에 200ms connect probe 포함** — 이 호출은 요청 goroutine 에 있으면 안 된다 | AST · httpapi.go:275 · transport_unix.go:402–424 |
| `fmt.Fprintf(errOut, …)` | 무경고 강등 금지 | 부작용은 stderr 뿐 | AST · httpapi.go:258·267·277 |

호출자: `runHTTPAPI` (httpapi.go:147) — 반환값을 `resources.reader.strategyRuntime`
(집계 스냅샷 경로)과 `httpapi.NewRouter(Options{StrategyRuntime: …})`(SSE 경로) **두
곳**에 같은 값으로 꽂는다. 재부착 wrapper 는 그래서 두 소비자를 동시에 고쳐야 한다.

## State mutations and fallbacks

- **프로세스 상태 변경 없음**: 디스크에 아무것도 쓰지 않고 아무것도 지우지 않는다.
  부작용은 stderr 경고 세 종류뿐이다.
- fallback 사다리: live → sentinel(unavailable) → nil(dormant). **아래로만 간다** —
  올라가는 경로가 없다는 것이 a109 가 닫는 결함이다.
- a109 이후 추가되는 상태: wrapper 안의 현재 reader, 직전 Read 성패, 마지막 시도
  시각, single-flight 진행 플래그. 전부 **메모리**이고 디스크 상태를 새로 만들지 않는다.

## Safety conclusion

- Safe edit boundary: 이 함수의 **반환 타입과 네 분기의 반환값**. 각 분기의 판정
  조건(무엇이 dormant 이고 무엇이 unavailable 인가)은 a108 계약이므로 **바꾸지 않는다** —
  wrapper 는 같은 판정을 상태 신호로 옮겨 담을 뿐이다.
- High-risk impact: **no** — 조회 전용 데몬의 화면 경로다. 주문·손절·사이징·Guardian·
  원장 어디에도 닿지 않는다. 다만 잘못 만들면 **화면 값이 조용히 틀린다**(dormant ↔
  unavailable 접힘), 그래서 두 화면 값을 테스트로 핀하는 것이 이 표면의 안전 요구다.
- 금지: 요청 경로(Snapshot·SSE)에서 `Dial`·connect probe 호출. 실패 반복 로그(30초마다
  stderr)도 금지 — 보고는 상태 전이 시 1회다(freeze P2-4).
