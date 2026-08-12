# Function Logic Map: `TestConcurrentObservationsOfOneConditionSendOnce`

- Source: `internal/obs/a096_one_send_per_condition_test.go`
- AST evidence: `ast.json` (4 branches)
- Risk scan: `risk-pattern-report.md`

테스트 함수지만 `check_analysis`가 **수정된 기존 Go 함수**로 잡는다. a097은 이 함수의
본문을 바꿨으므로 산출물을 만든다.

## 왜 바꿨나 — 그리고 리뷰의 처방이 왜 틀렸나

a096 2라운드 리뷰의 P2 ⑤는 "start barrier가 없어 `GOMAXPROCS=1`에서 30회 중 6회
오통과한다"였다. **진단은 맞고 처방은 틀렸다.** claim-and-send 잠금을 제거한 뮤턴트에
대해 `GOMAXPROCS=1`에서 100회씩 측정한 결과:

| 변형 | 뮤턴트 탐지 |
|---|---|
| 원본 (barrier 없음) | 96 / 100 |
| barrier만 추가 | 91 / 100 — **개선 없음** (차이 1.4σ) |
| **차단 publisher + barrier** | **100 / 100** |

정상 코드에서 오검출은 0/30이었다.

고치는 것은 barrier가 아니라 **전송을 `Publish` 안에 붙잡아 경합 창을 구조적으로 여는
것**이다. 첫 관측이 Publish에서 멈춰 있는 동안 그 행은 아직 PENDING이므로, 배제되지 않은
관측은 그것을 owed로 읽고 함께 발행한다 — publisher가 자기 동시 진입을 세므로 그 사건이
그대로 판정이 된다. barrier는 남긴다: 비용이 없고, "동시 관측"이라는 이름을 우연이 아니라
구성으로 참이 되게 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `pub` | `reentrantPublisher` — 첫 호출만 park | `a097_exclusion_is_an_event_test.go` | 두 번째 진입이 `peak > 1` |
| `n` | `a096Notifier`가 조립 | 같은 파일 | — |
| `observers` | 8 | 상수 | — |
| `ready`/`done` | 각각 8 | `sync.WaitGroup` | `ready.Wait()`가 전원 기동을 보장 |
| `start` | 닫히면 전원 출발 | 테스트 | — |

**불변식**: `ready.Wait()` 반환 시점에 8개 goroutine이 전부 존재하고 barrier에 park되어
있다. `close(start)` 뒤에는 아무도 다른 관측보다 먼저 끝날 수 없다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@218 | `for i := 0; i < observers; i++` | goroutine 8개 기동 | — | 자기 자신 |
| B2@223 | goroutine 안의 `Notify` 오류 | `t.Errorf` | — | 오류 경로 (정상 실행에서 미진입) |
| B3@242 | `peak > 1` | `t.Errorf` | — | 뮤턴트에서만 진입 |
| B4@246 | `calls != 1` | `t.Errorf` | — | 뮤턴트에서만 진입 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newReentrantPublisher` | 동시 진입 계수기 | — | AST calls |
| `a096Notifier` | 조립 | `t.Fatalf`로 중단 | AST calls |
| `n.Notify` | 관측 8회 | 오류는 B2 | AST calls |
| `pub.stats` | `calls`·`peak` 회수 | — | AST calls |

`time.Sleep(50ms)`는 **기회 제공이지 판정이 아니다.** 판정은 `peak`와 `calls`다.

## State mutations and fallbacks

- 프로덕션 상태 변경 없음. journal은 `t.TempDir()`에 격리된다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 테스트 본문 전체. 프로덕션 계약을 바꾸지 않는다.
- High-risk impact: no (테스트)
- 이 테스트가 지키는 프로덕션 한 줄: `claimAndDeliver`의 `n.mu.Lock()`.
