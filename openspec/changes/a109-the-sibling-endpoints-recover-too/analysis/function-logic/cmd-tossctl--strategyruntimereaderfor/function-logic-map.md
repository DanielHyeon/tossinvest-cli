# Function Logic Map: `strategyRuntimeReaderFor`

- Source: `cmd/tossctl/httpapi.go` (257-272)
- AST evidence: `ast.json` — AST 분기 1 · return 1 · defer 0
  (source_sha256 `fadda2b53b0ba6a6e851cdca81fcc26b598383bcfcb910a0c11bbd86ff5801cd`,
  **a109 §2.3 편집 후 재생성**)
- Risk scan: `risk-pattern-report.md`
- **편집 요약**: 편집 전 이 함수는 분기 4개로 세 값(nil·sentinel·live)을 직접 골라
  돌려줬다. a109 이후 그 **판정은 `resolveStrategyRuntimeReader` 로 그대로 옮겨 갔고**
  (같은 네 분기·같은 경고 문구), 이 함수는 그 결과를 재부착 wrapper 로 감싸는 일만 한다.
  판정 자체는 바뀌지 않았다 — 바뀐 것은 그 결과가 **굳지 않는다**는 것뿐이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 데몬의 수명 ctx | `runHTTPAPI` | nil 이면 `context.Background()` (B1). 이 ctx 가 **시도 goroutine 의 수명**이다 — 요청 ctx 를 쓰면 요청이 끝나는 순간 시도가 취소된다 |
| `root` | 프로파일 | CLI | 해석 실패는 `resolveStrategyRuntimeReader` 의 경고 + 부재 |
| `errOut` | 기동 stderr | cobra | wrapper 의 **전이 보고**가 계속 여기로 간다. 재시도 자체의 경고는 `io.Discard` 로 버린다(반복 로그 금지) |
| rate limit | 기본 30s | `strategyRuntimeRedialInterval` (package var) | 테스트 주입 가능. 값 자체는 `TestTheProductionRedialIntervalIsThirtySeconds` 가 핀 |
| 부팅 1회 해석 | **동기** | `resolveStrategyRuntimeReader(ctx, root, errOut)` | 오늘과 같다 — 경고도 같은 자리에서 찍힌다 |

**불변식 1**: 반환값은 **항상 non-nil** 이다. 그래서 소비자의 부재 판정은 nil 검사가 아니라
`httpapi.StrategyRuntimeAbsent` 다(그 함수 주석이 정본).

**불변식 2**: 이 함수도, 반환된 wrapper 의 요청 경로도 `strategyprojectionrpc.Dial` 을
**요청 goroutine 에서** 부르지 않는다. 부팅 1회만 동기이고 그 뒤는 전부 백그라운드다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (:259) | `ctx == nil` | 없음 | 없음 (배경 ctx 대입) | 간접 — `TestTheDaemonAttachesWhenTheEngineComesUpLater` 는 실제 ctx 를 넘긴다 |
| — (:272) | 언제나 | wrapper 생성 + **부팅 1회 동기 해석**(경고 포함) | 항상 non-nil wrapper | `TestADialFailureRendersUnavailableRatherThanNotConfigured` · `TestTheDaemonReattachesAfterTheEngineRestarts` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `resolveStrategyRuntimeReader` | 부팅 1회 해석 + 재시도의 몸통 | 오류를 돌려주지 않는다 — 세 값과 「live 인가」로 답한다 | AST · httpapi.go:265·269 |
| `strategyRuntimeAttachment.attach` | 부팅 결과를 wrapper 의 출발점으로 | 없음 | AST · httpapi_strategy_attach.go |
| `time.Now` | rate limit 의 시계 | 테스트는 가짜 시계를 넣는다 | AST |

## State mutations and fallbacks

- 디스크에 아무것도 쓰지 않는다. 부작용은 **부팅 1회의 stderr 경고**뿐이다.
- 새로 생기는 상태는 전부 wrapper 안의 메모리다(현재 reader·직전 Read 성패·마지막 시도
  시각·single-flight 플래그). 디스크 상태를 새로 만들지 않는다는 것이 D3a-2 와 같은 규율이다.
- fallback 사다리는 그대로다: live → sentinel(unavailable) → nil(부재/dormant).
  a109 가 더한 것은 **위로 올라가는 경로**다.

## Safety conclusion

- Safe edit boundary: 이 함수의 반환 타입과 wrapper 조립. 네 판정 조건은
  `resolveStrategyRuntimeReader` 로 옮겨졌고 그 조건은 a108 계약이므로 바꾸지 않는다.
- High-risk impact: **no** — 조회 전용 데몬의 화면 경로다. 주문·손절·사이징·Guardian·
  원장 어디에도 닿지 않는다. 잘못 만들면 **화면 값이 조용히 틀린다**(dormant ↔ unavailable
  접힘)는 것이 이 표면의 실제 위험이고, 그래서 세 소비자의 화면 값을 테스트가 핀한다.
- 금지: 요청 경로에서 dial·probe. 실패 반복 로그. 부재 판정의 두 번째 사본.
