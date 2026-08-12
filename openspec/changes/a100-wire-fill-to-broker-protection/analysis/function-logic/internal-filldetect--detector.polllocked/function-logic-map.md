# Function Logic Map: `Detector.pollLocked`

- Source: `internal/filldetect/detect.go` (L283-357)
- AST evidence: `ast.json` — 분기 10, return 6
- Risk scan: `risk-pattern-report.md`

## 이 함수가 대상인 이유

a100은 `internal/filldetect`를 **편집하지 않는다.** 그러나 설계가 "체결 감지 경로 안에서
보호를 계획하지 않는다"(D8)를 **이 함수의 분기를 근거로** 주장한다. 근거로 쓰므로 AST가
선행이다(`.claude/CLAUDE.md` 「단계 건너뛰기 금지」 4항).

**AST가 즉시 정정한 것.** 이전 초안은 이 로직을 `PollOnce`의 것으로 적었다. `PollOnce`는
L277-281의 5줄 래퍼이고 **분기가 0**이다 — 잠금을 잡고 `pollLocked`에 위임할 뿐이다.
인용하려던 L316-322는 전부 `pollLocked` 안에 있다. 손으로 읽은 인용이 **함수를 잘못
지목**하고 있었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `d.Ledger` | non-nil, 1-메서드 인터페이스 | `cmd/tossctl/engine.go:428-430`이 `JournalLedger`를 조립 | `Apply` 에러는 **사이클 실패 + outage**(B6) |
| `snaps` | `d.collect`가 만든 주문별 스냅샷 | 브로커 open/closed 주문 조회 | 수집 실패는 B2에서 사이클 종료 |
| `applied.CommittedAt` | journal이 커밋에 찍은 시각 | `ledger.go:106-107` | zero면 `clk.Now()`로 대체(B9) |
| `snap.BrokerVisibleAt` | `execution.filledAt` | 브로커 응답 | 신선도 SLO의 시작점 |
| `d.slo` / `d.outage` | 누적 관측기 | 이 함수만 갱신 | `evaluateSLO`가 `Gate.Block` 유발 |

**불변식 1 — 루프는 직렬이다.** L316의 `for`는 스냅샷을 하나씩 처리하고, 각 반복은 앞
반복의 `Ledger.Apply`가 반환한 **뒤에만** 시작한다. `Apply` 안에서 시간을 쓰면 그 시간은
같은 사이클 뒤쪽 스냅샷 전부의 대기 시간이 된다.

**불변식 2 — 신선도 SLO는 이 루프 안에서 계산된다.** `latency := committed.Sub(snap.BrokerVisibleAt)`
(L335) → `d.slo.observe`(L342) → `d.evaluateSLO(now)`(L349). 측정 구간의 끝점은 `Apply`가
**반환한 시각이 아니라** journal이 커밋에 찍은 시각이므로, 자기 자신의 지연은 섞이지 않는다.
그러나 **뒤 스냅샷의 커밋 시각은 앞 스냅샷의 `Apply` 소요만큼 밀린다.**

**불변식 3 — `Apply` 에러는 남은 스냅샷을 버린다.** B6은 `continue`가 아니라 `return`이다.
한 주문의 `Apply` 실패가 같은 사이클의 **아직 처리되지 않은 다른 주문들의 체결 반영까지**
중단시킨다.

## Branches and early returns

| Branch | 조건 (L) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `d.validate() != nil` (285) | 없음 | `Cycle{}, err` | 구성 오류 |
| B2 | `d.collect` 실패 (294) | `outage.failure` | `cycle, err` | 주문 조회 실패 |
| B3 | `d.Positions.Positions` 실패 (301) | `outage.failure` | wrap된 err | **미실행** |
| B4 | `d.sweepBalance` 실패 (309) | `outage.failure` | `cycle, err` | **미실행** |
| B5 | `range snaps` 본문 (316) | 스냅샷별 처리 | — | 정상 경로 |
| B6 | `d.Ledger.Apply` 실패 (318) | `outage.failure` + **남은 스냅샷 폐기** | wrap된 err | **미실행 — a100의 핵심** |
| B7 | `applied.FailClosed` (324) | `blockSymbol` | `continue` | 모순 스냅샷 |
| B8 | `applied.Delta > 0` (329) | `Fills++`, latency 관측 | — | 신규 체결 |
| B9 | `committed.IsZero()` (332) | `committed = clk.Now()` | — | **미실행** |
| B10 | `latency < 0` (336) | `latency = 0` | — | **미실행** (시계 skew 바닥) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `d.Ledger.Apply` | 체결 스냅샷을 journal에 durable 커밋 | 에러 = 사이클 실패 + outage + 잔여 폐기(B6) | AST L317, B6 |
| `d.collect` | 주문 조회 → 스냅샷 | 에러 = 사이클 실패(B2) | AST L293 |
| `d.slo.observe` | 신선도 표본 적재 | 없음 | AST L342 |
| `d.evaluateSLO` | 표본이 임계 초과면 `Gate.Block(ReasonFillDetectionSLO)` | 게이트 차단 | AST L349, `detect.go:639` |
| `d.outage.failure/success` | outage 판정 | 없음 | AST L295/302/310/319, L347 |

## State mutations and fallbacks

- `cycle.Applied`는 `Apply` 성공분만 누적한다. B6에서 반환하면 그 사이클의 나머지는 없다.
- `d.lastPollAt`은 **성공 사이클에서만** 갱신된다(L352, B6 경로는 도달하지 못함).
- `Run`은 사이클이 끝난 **뒤에** `PollInterval`을 잔다(`detect.go:523`). 사이클이 길어지면
  다음 관측이 그만큼 늦다.

## Safety conclusion

- Safe edit boundary: **a100은 이 함수를 편집하지 않는다.** 조립 지점(`Ledger` 구현)만 바꾼다.
- High-risk impact: **yes.** 이 함수는 체결 감지·outage·신선도 SLO의 단일 지점이다.
- **설계 귀결(D8):** 보호 왕복을 `Ledger.Apply` 안에 두면 (1) 같은 사이클 뒤 스냅샷의
  `CommittedAt`이 밀려 신선도 표본이 오염되고 → `evaluateSLO` → `Gate.Block`, (2) 보호 실패가
  `Apply` 에러로 새면 B6이 남은 체결 반영을 버린다. 두 결과 모두 안전 불변식 §0-3(손절·비상
  청산 즉시성 불약화)에 걸린다. **따라서 보호는 이 경로 밖에서 수렴한다.**
